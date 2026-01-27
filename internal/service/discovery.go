package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// discoveryPageSize 发现页固定分页大小
const discoveryPageSize = 10

// IDiscoveryService 发现服务接口
type IDiscoveryService interface {
	// GetFreshPosts 获取新鲜发布的帖子（最新发布）
	GetFreshPosts(ctx context.Context) (*schema.DiscoveryFreshResponse, error)
	// GetLatestDiscussions 获取最新讨论（按最近被回复的帖子顺序）
	GetLatestDiscussions(ctx context.Context) (*schema.DiscoveryLatestDiscussionResponse, error)
	// GetInteractiveComments 获取互动评论
	GetInteractiveComments(ctx context.Context) (*schema.DiscoveryCommentsResponse, error)
}

// DiscoveryService 发现服务实现
type DiscoveryService struct {
	db           *ent.Client
	postRepo     repository.IPostRepository
	commentRepo  repository.ICommentRepository
	userRepo     repository.IUserRepository
	categoryRepo repository.ICategoryRepository
	log          *zap.Logger
}

// NewDiscoveryService 创建发现服务实例
func NewDiscoveryService(db *ent.Client, repos *repository.Repositories, log *zap.Logger) IDiscoveryService {
	return &DiscoveryService{
		db:           db,
		postRepo:     repos.Post,
		commentRepo:  repos.Comment,
		userRepo:     repos.User,
		categoryRepo: repos.Category,
		log:          log,
	}
}

// GetFreshPosts 获取新鲜发布的帖子（最新发布）
func (s *DiscoveryService) GetFreshPosts(ctx context.Context) (*schema.DiscoveryFreshResponse, error) {
	s.log.Info("获取新鲜发布帖子", tracing.WithTraceIDField(ctx))

	// 查询最新发布的帖子，按创建时间降序（StatusNormal 和 StatusLocked 状态）
	posts, err := s.db.Post.Query().
		Where(
			post.StatusIn(post.StatusNormal, post.StatusLocked),
			post.ReadPermissionEQ(post.ReadPermissionPublic),
		).
		Order(ent.Desc(post.FieldCreatedAt)).
		Limit(discoveryPageSize).
		All(ctx)
	if err != nil {
		s.log.Error("查询新鲜发布帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 批量获取用户和版块信息
	userIDs, categoryIDs := s.collectIDs(posts)
	userMap := s.batchGetUsers(ctx, userIDs)
	categoryMap := s.batchGetCategories(ctx, categoryIDs)

	// 批量获取帖子评论数
	postIDs := make([]int, 0, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}
	commentCountMap := s.batchGetCommentCounts(ctx, postIDs)

	// 构建响应
	items := make([]schema.DiscoveryPostItem, 0, len(posts))
	for _, p := range posts {
		item := schema.DiscoveryPostItem{
			ID:           p.ID,
			Title:        p.Title,
			CategoryID:   p.CategoryID,
			CategoryName: categoryMap[p.CategoryID],
			UserID:       p.UserID,
			Username:     userMap[p.UserID].Username,
			Avatar:       userMap[p.UserID].Avatar,
			ViewCount:    p.ViewCount,
			LikeCount:    p.LikeCount,
			CommentCount: commentCountMap[p.ID],
			IsEssence:    p.IsEssence,
			CreatedAt:    p.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:    p.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
		items = append(items, item)
	}

	return &schema.DiscoveryFreshResponse{Posts: items}, nil
}

// GetLatestDiscussions 获取最新讨论（按最近被回复的帖子顺序）
func (s *DiscoveryService) GetLatestDiscussions(ctx context.Context) (*schema.DiscoveryLatestDiscussionResponse, error) {
	s.log.Info("获取最新讨论帖子", tracing.WithTraceIDField(ctx))

	// 先查询最近有评论的帖子ID（通过评论表获取最新回复信息）
	latestComments, err := s.db.Comment.Query().
		Order(ent.Desc(comment.FieldCreatedAt)).
		Limit(100). // 多查一些以确保能获取到足够的不重复帖子
		All(ctx)
	if err != nil {
		s.log.Error("查询最新评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 按帖子ID分组，记录每个帖子的最新评论信息
	type postCommentInfo struct {
		PostID          int
		LastReplyUserID int
		LastReplyAt     string
		CommentCount    int
	}
	postInfoMap := make(map[int]*postCommentInfo)
	orderedPostIDs := make([]int, 0)

	for _, c := range latestComments {
		if _, exists := postInfoMap[c.PostID]; !exists {
			postInfoMap[c.PostID] = &postCommentInfo{
				PostID:          c.PostID,
				LastReplyUserID: c.UserID,
				LastReplyAt:     c.CreatedAt.Format(time_tools.DateTimeFormat),
			}
			orderedPostIDs = append(orderedPostIDs, c.PostID)
		}
	}

	// 限制帖子数量
	if len(orderedPostIDs) > discoveryPageSize {
		orderedPostIDs = orderedPostIDs[:discoveryPageSize]
	}

	if len(orderedPostIDs) == 0 {
		return &schema.DiscoveryLatestDiscussionResponse{Posts: []schema.DiscoveryLatestDiscussionItem{}}, nil
	}

	// 批量获取帖子信息
	posts, err := s.db.Post.Query().
		Where(
			post.IDIn(orderedPostIDs...),
			post.StatusIn(post.StatusNormal, post.StatusLocked),
			post.ReadPermissionEQ(post.ReadPermissionPublic),
		).
		All(ctx)
	if err != nil {
		s.log.Error("查询帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 构建帖子Map
	postMap := make(map[int]*ent.Post)
	for _, p := range posts {
		postMap[p.ID] = p
	}

	// 批量获取用户和版块信息
	userIDs, categoryIDs := s.collectIDs(posts)

	// 收集最后回复用户ID
	for _, info := range postInfoMap {
		if info.LastReplyUserID > 0 {
			userIDs = append(userIDs, info.LastReplyUserID)
		}
	}

	userMap := s.batchGetUsers(ctx, userIDs)
	categoryMap := s.batchGetCategories(ctx, categoryIDs)

	// 批量获取帖子评论数
	commentCountMap := s.batchGetCommentCounts(ctx, orderedPostIDs)

	// 按原始顺序构建响应
	items := make([]schema.DiscoveryLatestDiscussionItem, 0, len(orderedPostIDs))
	for _, postID := range orderedPostIDs {
		p, ok := postMap[postID]
		if !ok {
			continue
		}

		info := postInfoMap[postID]
		item := schema.DiscoveryLatestDiscussionItem{
			ID:                p.ID,
			Title:             p.Title,
			CategoryID:        p.CategoryID,
			CategoryName:      categoryMap[p.CategoryID],
			UserID:            p.UserID,
			Username:          userMap[p.UserID].Username,
			Avatar:            userMap[p.UserID].Avatar,
			CommentCount:      commentCountMap[p.ID],
			LastReplyUserID:   info.LastReplyUserID,
			LastReplyUsername: "",
			LastReplyAt:       info.LastReplyAt,
		}

		// 设置最后回复用户名
		if info.LastReplyUserID > 0 {
			if u, ok := userMap[info.LastReplyUserID]; ok {
				item.LastReplyUsername = u.Username
			}
		}

		items = append(items, item)
	}

	return &schema.DiscoveryLatestDiscussionResponse{Posts: items}, nil
}

// GetInteractiveComments 获取互动评论
func (s *DiscoveryService) GetInteractiveComments(ctx context.Context) (*schema.DiscoveryCommentsResponse, error) {
	s.log.Info("获取互动评论", tracing.WithTraceIDField(ctx))

	// 查询最新评论，按创建时间降序（多查一些以过滤非公开帖子的评论）
	comments, err := s.db.Comment.Query().
		Order(ent.Desc(comment.FieldCreatedAt)).
		Limit(discoveryPageSize * 3).
		All(ctx)
	if err != nil {
		s.log.Error("查询互动评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 收集用户ID和帖子ID
	userIDs := make([]int, 0, len(comments))
	postIDs := make([]int, 0, len(comments))
	for _, c := range comments {
		userIDs = append(userIDs, c.UserID)
		postIDs = append(postIDs, c.PostID)
	}

	// 批量获取用户和帖子信息
	userMap := s.batchGetUsers(ctx, userIDs)
	postMap := s.batchGetPosts(ctx, postIDs)

	// 构建响应
	items := make([]schema.DiscoveryCommentItem, 0, discoveryPageSize)
	for _, c := range comments {
		// 跳过帖子不存在或非公开的评论
		p, ok := postMap[c.PostID]
		if !ok || (p.Status != post.StatusNormal && p.Status != post.StatusLocked) || p.ReadPermission != post.ReadPermissionPublic {
			continue
		}

		item := schema.DiscoveryCommentItem{
			ID:        c.ID,
			Content:   c.Content,
			PostID:    c.PostID,
			PostTitle: p.Title,
			UserID:    c.UserID,
			Username:  userMap[c.UserID].Username,
			Avatar:    userMap[c.UserID].Avatar,
			LikeCount: c.LikeCount,
			CreatedAt: c.CreatedAt.Format(time_tools.DateTimeFormat),
		}
		items = append(items, item)

		// 达到数量限制后停止
		if len(items) >= discoveryPageSize {
			break
		}
	}

	return &schema.DiscoveryCommentsResponse{Comments: items}, nil
}

// collectIDs 收集帖子中的用户ID和版块ID
func (s *DiscoveryService) collectIDs(posts []*ent.Post) ([]int, []int) {
	userIDs := make([]int, 0, len(posts))
	categoryIDs := make([]int, 0, len(posts))
	userIDSet := make(map[int]bool)
	categoryIDSet := make(map[int]bool)

	for _, p := range posts {
		if !userIDSet[p.UserID] {
			userIDs = append(userIDs, p.UserID)
			userIDSet[p.UserID] = true
		}
		if !categoryIDSet[p.CategoryID] {
			categoryIDs = append(categoryIDs, p.CategoryID)
			categoryIDSet[p.CategoryID] = true
		}
	}

	return userIDs, categoryIDs
}

// batchGetUsers 批量获取用户信息
func (s *DiscoveryService) batchGetUsers(ctx context.Context, userIDs []int) map[int]*ent.User {
	userMap := make(map[int]*ent.User)
	if len(userIDs) == 0 {
		return userMap
	}

	// 去重
	uniqueIDs := make([]int, 0, len(userIDs))
	seen := make(map[int]bool)
	for _, id := range userIDs {
		if !seen[id] {
			uniqueIDs = append(uniqueIDs, id)
			seen[id] = true
		}
	}

	users, err := s.userRepo.GetByIDsWithFields(ctx, uniqueIDs, []string{user.FieldID, user.FieldUsername, user.FieldAvatar})
	if err != nil {
		s.log.Warn("批量查询用户信息失败", zap.Error(err))
		return userMap
	}

	for _, u := range users {
		userMap[u.ID] = u
	}

	// 为不存在的用户设置默认值
	for _, id := range uniqueIDs {
		if _, ok := userMap[id]; !ok {
			userMap[id] = &ent.User{ID: id, Username: "未知用户", Avatar: ""}
		}
	}

	return userMap
}

// batchGetCategories 批量获取版块信息
func (s *DiscoveryService) batchGetCategories(ctx context.Context, categoryIDs []int) map[int]string {
	categoryMap := make(map[int]string)
	if len(categoryIDs) == 0 {
		return categoryMap
	}

	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, categoryIDs, []string{category.FieldID, category.FieldName})
	if err != nil {
		s.log.Warn("批量查询版块信息失败", zap.Error(err))
		return categoryMap
	}

	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	return categoryMap
}

// batchGetPosts 批量获取帖子信息
func (s *DiscoveryService) batchGetPosts(ctx context.Context, postIDs []int) map[int]*ent.Post {
	postMap := make(map[int]*ent.Post)
	if len(postIDs) == 0 {
		return postMap
	}

	// 去重
	uniqueIDs := make([]int, 0, len(postIDs))
	seen := make(map[int]bool)
	for _, id := range postIDs {
		if !seen[id] {
			uniqueIDs = append(uniqueIDs, id)
			seen[id] = true
		}
	}

	posts, err := s.postRepo.GetByIDs(ctx, uniqueIDs)
	if err != nil {
		s.log.Warn("批量查询帖子信息失败", zap.Error(err))
		return postMap
	}

	for _, p := range posts {
		postMap[p.ID] = p
	}

	return postMap
}

// batchGetCommentCounts 批量获取帖子评论数
func (s *DiscoveryService) batchGetCommentCounts(ctx context.Context, postIDs []int) map[int]int {
	countMap := make(map[int]int)
	if len(postIDs) == 0 {
		return countMap
	}

	// 查询所有相关帖子的评论
	comments, err := s.db.Comment.Query().
		Where(comment.PostIDIn(postIDs...)).
		All(ctx)
	if err != nil {
		s.log.Warn("批量查询评论数失败", zap.Error(err))
		return countMap
	}

	// 统计每个帖子的评论数
	for _, c := range comments {
		countMap[c.PostID]++
	}

	return countMap
}
