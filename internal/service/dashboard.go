package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IDashboardService Dashboard service interface | 仪表盘服务接口
type IDashboardService interface {
	// GetDashboardStats Get dashboard statistics | 获取仪表盘统计数据
	GetDashboardStats(ctx context.Context) (*schema.DashboardStatsResponse, error)
	// GetRecentActivity Get recent activity | 获取最近活动
	GetRecentActivity(ctx context.Context) (*schema.RecentActivityResponse, error)
	// GetPopularPosts Get popular posts | 获取热门帖子
	GetPopularPosts(ctx context.Context) (*schema.PopularPostsResponse, error)
	// GetPopularCategories Get popular categories | 获取热门版块
	GetPopularCategories(ctx context.Context) (*schema.PopularCategoriesResponse, error)
}

// DashboardService Dashboard service implementation | 仪表盘服务实现
type DashboardService struct {
	db           *ent.Client
	userRepo     repository.IUserRepository
	postRepo     repository.IPostRepository
	commentRepo  repository.ICommentRepository
	categoryRepo repository.ICategoryRepository
	cache        cache.ICacheService
	logger       *zap.Logger
}

// NewDashboardService Create dashboard service instance | 创建仪表盘服务实例
func NewDashboardService(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, logger *zap.Logger) IDashboardService {
	return &DashboardService{
		db:           db,
		userRepo:     repos.User,
		postRepo:     repos.Post,
		commentRepo:  repos.Comment,
		categoryRepo: repos.Category,
		cache:        cacheService,
		logger:       logger,
	}
}

// GetDashboardStats Get dashboard statistics | 获取仪表盘统计数据
func (s *DashboardService) GetDashboardStats(ctx context.Context) (*schema.DashboardStatsResponse, error) {
	s.logger.Info("获取仪表盘统计数据", tracing.WithTraceIDField(ctx))

	// Get user statistics | 获取用户统计
	userStats, err := s.getUserStats(ctx)
	if err != nil {
		s.logger.Error("获取用户统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户统计失败: %w", err)
	}

	// Get post statistics | 获取帖子统计
	postStats, err := s.getPostStats(ctx)
	if err != nil {
		s.logger.Error("获取帖子统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子统计失败: %w", err)
	}

	// Get comment statistics | 获取评论统计
	commentStats, err := s.getCommentStats(ctx)
	if err != nil {
		s.logger.Error("获取评论统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论统计失败: %w", err)
	}

	// Get category statistics | 获取版块统计
	categoryStats, err := s.getCategoryStats(ctx)
	if err != nil {
		s.logger.Error("获取版块统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块统计失败: %w", err)
	}

	// Get system statistics | 获取系统统计
	systemStats, err := s.getSystemStats(ctx)
	if err != nil {
		s.logger.Error("获取系统统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取系统统计失败: %w", err)
	}

	return &schema.DashboardStatsResponse{
		UserStats:     userStats,
		PostStats:     postStats,
		CommentStats:  commentStats,
		CategoryStats: categoryStats,
		SystemStats:   systemStats,
	}, nil
}

// getUserStats Get user statistics | 获取用户统计
func (s *DashboardService) getUserStats(ctx context.Context) (schema.UserStats, error) {
	// Total user count | 总用户数
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return schema.UserStats{}, err
	}

	// Active user count (users created within 30 days, simplified handling) | 活跃用户数（30天内创建的用户，简化处理）
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	activeUsers, err := s.userRepo.CountWithCondition(ctx, func(q *ent.UserQuery) *ent.UserQuery {
		return q.Where(user.CreatedAtGTE(thirtyDaysAgo))
	})
	if err != nil {
		return schema.UserStats{}, err
	}

	// New user count (today) | 新增用户数（今日）
	today := time.Now().Truncate(24 * time.Hour)
	newUsers, err := s.userRepo.CountWithCondition(ctx, func(q *ent.UserQuery) *ent.UserQuery {
		return q.Where(user.CreatedAtGTE(today))
	})
	if err != nil {
		return schema.UserStats{}, err
	}

	// Online user count (simplified handling, should actually be tracked via Redis or other methods) | 在线用户数（这里简化处理，实际应该通过Redis或其他方式统计）
	onlineUsers := int64(0) // 需要根据实际在线用户统计逻辑实现

	// Banned user count | 被封禁用户数
	bannedUsers, err := s.userRepo.CountWithCondition(ctx, func(q *ent.UserQuery) *ent.UserQuery {
		return q.Where(user.StatusEQ(user.StatusBlocked))
	})
	if err != nil {
		return schema.UserStats{}, err
	}

	// Moderator count | 版主数量
	moderatorCount, err := s.userRepo.CountWithCondition(ctx, func(q *ent.UserQuery) *ent.UserQuery {
		return q.Where(user.RoleEQ(user.RoleModerator))
	})
	if err != nil {
		return schema.UserStats{}, err
	}

	return schema.UserStats{
		TotalUsers:     int64(totalUsers),
		ActiveUsers:    int64(activeUsers),
		NewUsers:       int64(newUsers),
		OnlineUsers:    onlineUsers,
		BannedUsers:    int64(bannedUsers),
		ModeratorCount: int64(moderatorCount),
	}, nil
}

// getPostStats Get post statistics | 获取帖子统计
func (s *DashboardService) getPostStats(ctx context.Context) (schema.PostStats, error) {
	// Total post count | 总帖子数
	totalPosts, err := s.postRepo.Count(ctx)
	if err != nil {
		return schema.PostStats{}, err
	}

	// Published post count | 已发布帖子数
	publishedPosts, err := s.postRepo.CountWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		return q.Where(post.StatusEQ(post.StatusNormal))
	})
	if err != nil {
		return schema.PostStats{}, err
	}

	// Draft post count | 草稿帖子数
	draftPosts, err := s.postRepo.CountWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		return q.Where(post.StatusEQ(post.StatusDraft))
	})
	if err != nil {
		return schema.PostStats{}, err
	}

	// Locked post count | 被锁定帖子数
	lockedPosts, err := s.postRepo.CountWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		return q.Where(post.StatusEQ(post.StatusLocked))
	})
	if err != nil {
		return schema.PostStats{}, err
	}

	// Essence post count | 精华帖子数
	essencePosts, err := s.postRepo.CountWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		return q.Where(post.IsEssenceEQ(true))
	})
	if err != nil {
		return schema.PostStats{}, err
	}

	// Pinned post count | 置顶帖子数
	pinnedPosts, err := s.postRepo.CountWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		return q.Where(post.IsPinnedEQ(true))
	})
	if err != nil {
		return schema.PostStats{}, err
	}

	// Today new post count | 今日新增帖子数
	today := time.Now().Truncate(24 * time.Hour)
	todayPosts, err := s.postRepo.CountWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		return q.Where(post.CreatedAtGTE(today))
	})
	if err != nil {
		return schema.PostStats{}, err
	}

	return schema.PostStats{
		TotalPosts:     int64(totalPosts),
		PublishedPosts: int64(publishedPosts),
		DraftPosts:     int64(draftPosts),
		LockedPosts:    int64(lockedPosts),
		EssencePosts:   int64(essencePosts),
		PinnedPosts:    int64(pinnedPosts),
		TodayPosts:     int64(todayPosts),
	}, nil
}

// getCommentStats Get comment statistics | 获取评论统计
func (s *DashboardService) getCommentStats(ctx context.Context) (schema.CommentStats, error) {
	// Total comment count | 总评论数
	totalComments, err := s.commentRepo.Count(ctx)
	if err != nil {
		return schema.CommentStats{}, err
	}

	// Selected comment count | 精选评论数
	selectedComments, err := s.commentRepo.CountWithCondition(ctx, func(q *ent.CommentQuery) *ent.CommentQuery {
		return q.Where(comment.IsSelectedEQ(true))
	})
	if err != nil {
		return schema.CommentStats{}, err
	}

	// Pinned comment count | 置顶评论数
	pinnedComments, err := s.commentRepo.CountWithCondition(ctx, func(q *ent.CommentQuery) *ent.CommentQuery {
		return q.Where(comment.IsPinnedEQ(true))
	})
	if err != nil {
		return schema.CommentStats{}, err
	}

	// Today new comment count | 今日新增评论数
	today := time.Now().Truncate(24 * time.Hour)
	todayComments, err := s.commentRepo.CountWithCondition(ctx, func(q *ent.CommentQuery) *ent.CommentQuery {
		return q.Where(comment.CreatedAtGTE(today))
	})
	if err != nil {
		return schema.CommentStats{}, err
	}

	return schema.CommentStats{
		TotalComments:    int64(totalComments),
		SelectedComments: int64(selectedComments),
		PinnedComments:   int64(pinnedComments),
		TodayComments:    int64(todayComments),
	}, nil
}

// getCategoryStats Get category statistics | 获取版块统计
func (s *DashboardService) getCategoryStats(ctx context.Context) (schema.CategoryStats, error) {
	// Total category count | 总版块数
	totalCategories, err := s.categoryRepo.Count(ctx)
	if err != nil {
		return schema.CategoryStats{}, err
	}

	// Active category count (categories with normal status) | 活跃版块数（正常状态的版块）
	activeCategories, err := s.categoryRepo.CountWithCondition(ctx, func(q *ent.CategoryQuery) *ent.CategoryQuery {
		return q.Where(category.StatusEQ(category.StatusNormal))
	})
	if err != nil {
		return schema.CategoryStats{}, err
	}

	// Hidden category count | 隐藏版块数
	hiddenCategories, err := s.categoryRepo.CountWithCondition(ctx, func(q *ent.CategoryQuery) *ent.CategoryQuery {
		return q.Where(category.StatusEQ(category.StatusHidden))
	})
	if err != nil {
		return schema.CategoryStats{}, err
	}

	// Locked category count | 锁定版块数
	lockedCategories, err := s.categoryRepo.CountWithCondition(ctx, func(q *ent.CategoryQuery) *ent.CategoryQuery {
		return q.Where(category.StatusEQ(category.StatusLocked))
	})
	if err != nil {
		return schema.CategoryStats{}, err
	}

	return schema.CategoryStats{
		TotalCategories:  int64(totalCategories),
		ActiveCategories: int64(activeCategories),
		HiddenCategories: int64(hiddenCategories),
		LockedCategories: int64(lockedCategories),
	}, nil
}

// getSystemStats Get system statistics | 获取系统统计
func (s *DashboardService) getSystemStats(ctx context.Context) (schema.SystemStats, error) {
	// Total views (simplified handling, should actually be from access logs) | 总浏览量（这里简化处理，实际应该从访问日志统计）
	totalViews := int64(0) // 需要根据实际浏览量统计逻辑实现

	// Today views | 今日浏览量
	todayViews := int64(0) // 需要根据实际今日浏览量统计逻辑实现

	// Total likes (post likes + comment likes) | 总点赞数（帖子点赞数 + 评论点赞数）
	posts, err := s.postRepo.GetAll(ctx)
	if err != nil {
		return schema.SystemStats{}, err
	}

	totalPostLikes := int64(0)
	for _, p := range posts {
		totalPostLikes += int64(p.LikeCount)
	}

	comments, err := s.commentRepo.GetAll(ctx)
	if err != nil {
		return schema.SystemStats{}, err
	}

	totalCommentLikes := int64(0)
	for _, c := range comments {
		totalCommentLikes += int64(c.LikeCount)
	}

	totalLikes := totalPostLikes + totalCommentLikes

	// Today likes (simplified handling) | 今日点赞数（简化处理）
	todayLikes := int64(0) // 需要根据实际今日点赞数统计逻辑实现

	// Storage used (simplified handling) | 存储使用量（简化处理）
	storageUsed := int64(0) // 需要根据实际存储使用量统计逻辑实现

	// Database size (simplified handling) | 数据库大小（简化处理）
	databaseSize := int64(0) // 需要根据实际数据库大小统计逻辑实现

	return schema.SystemStats{
		TotalViews:   totalViews,
		TodayViews:   todayViews,
		TotalLikes:   totalLikes,
		TodayLikes:   todayLikes,
		StorageUsed:  storageUsed,
		DatabaseSize: databaseSize,
	}, nil
}

// GetRecentActivity Get recent activity | 获取最近活动
func (s *DashboardService) GetRecentActivity(ctx context.Context) (*schema.RecentActivityResponse, error) {
	s.logger.Info("获取最近活动", tracing.WithTraceIDField(ctx))

	// Get recent posts (last 10) | 获取最近帖子（最近10条）
	recentPosts, err := s.postRepo.ListWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		return q.Order(ent.Desc(post.FieldCreatedAt))
	}, 10)
	if err != nil {
		return nil, fmt.Errorf("获取最近帖子失败: %w", err)
	}

	// Get recent comments (last 10) | 获取最近评论（最近10条）
	recentComments, err := s.commentRepo.ListWithCondition(ctx, func(q *ent.CommentQuery) *ent.CommentQuery {
		return q.Order(ent.Desc(comment.FieldCreatedAt))
	}, 10)
	if err != nil {
		return nil, fmt.Errorf("获取最近评论失败: %w", err)
	}

	// Get new users (last 10) | 获取新用户（最近10个）
	newUsers, err := s.userRepo.ListWithCondition(ctx, func(q *ent.UserQuery) *ent.UserQuery {
		return q.Order(ent.Desc(user.FieldCreatedAt))
	}, 10)
	if err != nil {
		return nil, fmt.Errorf("获取新用户失败: %w", err)
	}

	// Collect user IDs and category IDs to query | 收集需要查询的用户ID和版块ID
	userIDs := make(map[int]bool)
	categoryIDs := make(map[int]bool)
	postIDs := make(map[int]bool)
	for _, p := range recentPosts {
		userIDs[p.UserID] = true
		categoryIDs[p.CategoryID] = true
	}
	for _, c := range recentComments {
		userIDs[c.UserID] = true
		postIDs[c.PostID] = true
	}

	// Batch query user information | 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar})
	if err != nil {
		s.logger.Warn("批量查询用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}
	type userInfo struct {
		Username string
		Avatar   string
	}
	userMap := make(map[int]userInfo)
	for _, u := range users {
		userMap[u.ID] = userInfo{Username: u.Username, Avatar: u.Avatar}
	}

	// Batch query category information | 批量查询版块信息
	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id)
	}
	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, categoryIDList, []string{category.FieldID, category.FieldName})
	if err != nil {
		s.logger.Warn("批量查询版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}
	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	// Batch query post information | 批量查询帖子信息
	postIDList := make([]int, 0, len(postIDs))
	for id := range postIDs {
		postIDList = append(postIDList, id)
	}
	postsData, err := s.postRepo.GetByIDsWithFields(ctx, postIDList, []string{post.FieldID, post.FieldTitle})
	if err != nil {
		s.logger.Warn("批量查询帖子信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}
	postMap := make(map[int]string)
	for _, p := range postsData {
		postMap[p.ID] = p.Title
	}

	// Convert format | 转换格式
	posts := make([]schema.RecentPost, len(recentPosts))
	for i, p := range recentPosts {
		ui := userMap[p.UserID]
		posts[i] = schema.RecentPost{
			ID:           p.ID,
			Title:        p.Title,
			Username:     ui.Username,
			Avatar:       ui.Avatar,
			CategoryName: categoryMap[p.CategoryID],
			CreatedAt:    p.CreatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	comments := make([]schema.RecentComment, len(recentComments))
	for i, c := range recentComments {
		// Truncate comment content to first 100 characters | 截取评论内容前100字符
		content := c.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		ui := userMap[c.UserID]
		comments[i] = schema.RecentComment{
			ID:        c.ID,
			Content:   content,
			Username:  ui.Username,
			Avatar:    ui.Avatar,
			PostTitle: postMap[c.PostID],
			CreatedAt: c.CreatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	newUserList := make([]schema.NewUser, len(newUsers))
	for i, u := range newUsers {
		newUserList[i] = schema.NewUser{
			ID:        u.ID,
			Username:  u.Username,
			Avatar:    u.Avatar,
			Email:     u.Email,
			CreatedAt: u.CreatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.RecentActivityResponse{
		RecentPosts:    posts,
		RecentComments: comments,
		NewUsers:       newUserList,
	}, nil
}

// GetPopularPosts Get popular posts | 获取热门帖子
func (s *DashboardService) GetPopularPosts(ctx context.Context) (*schema.PopularPostsResponse, error) {
	s.logger.Info("获取热门帖子", tracing.WithTraceIDField(ctx))

	// Get popular posts sorted by view count | 按浏览量排序获取热门帖子
	popularPosts, err := s.postRepo.ListWithCondition(ctx, func(q *ent.PostQuery) *ent.PostQuery {
		return q.Order(ent.Desc(post.FieldViewCount))
	}, 10)
	if err != nil {
		return nil, fmt.Errorf("获取热门帖子失败: %w", err)
	}

	// Collect user IDs and category IDs to query | 收集需要查询的用户ID和版块ID
	userIDs := make(map[int]bool)
	categoryIDs := make(map[int]bool)
	postIDs := make([]int, len(popularPosts))
	for i, p := range popularPosts {
		userIDs[p.UserID] = true
		categoryIDs[p.CategoryID] = true
		postIDs[i] = p.ID
	}

	// Batch query user information | 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar})
	if err != nil {
		s.logger.Warn("批量查询用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}
	type userInfo struct {
		Username string
		Avatar   string
	}
	userMap := make(map[int]userInfo)
	for _, u := range users {
		userMap[u.ID] = userInfo{Username: u.Username, Avatar: u.Avatar}
	}

	// Batch query category information | 批量查询版块信息
	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id)
	}
	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, categoryIDList, []string{category.FieldID, category.FieldName})
	if err != nil {
		s.logger.Warn("批量查询版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}
	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	// Batch query comment count | 批量查询评论数
	// Note: simplified handling, should actually use GroupBy to optimize performance | 注意: 这里简化处理，实际应该使用GroupBy优化性能
	commentCountMap := make(map[int]int)
	for _, postID := range postIDs {
		count, err := s.commentRepo.CountByPostID(ctx, postID)
		if err != nil {
			s.logger.Warn("查询评论数失败", zap.Error(err), tracing.WithTraceIDField(ctx), zap.Int("postID", postID))
		}
		commentCountMap[postID] = count
	}

	// Convert format | 转换格式
	posts := make([]schema.PopularPost, len(popularPosts))
	for i, p := range popularPosts {
		ui := userMap[p.UserID]
		posts[i] = schema.PopularPost{
			ID:           p.ID,
			Title:        p.Title,
			Username:     ui.Username,
			Avatar:       ui.Avatar,
			CategoryName: categoryMap[p.CategoryID],
			ViewCount:    p.ViewCount,
			LikeCount:    p.LikeCount,
			CommentCount: commentCountMap[p.ID],
			CreatedAt:    p.CreatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.PopularPostsResponse{
		Posts: posts,
	}, nil
}

// GetPopularCategories Get popular categories | 获取热门版块
func (s *DashboardService) GetPopularCategories(ctx context.Context) (*schema.PopularCategoriesResponse, error) {
	s.logger.Info("获取热门版块", tracing.WithTraceIDField(ctx))

	// Get popular categories sorted by post count | 按帖子数量排序获取热门版块
	popularCategories, err := s.categoryRepo.ListWithCondition(ctx, func(q *ent.CategoryQuery) *ent.CategoryQuery {
		return q.Order(ent.Desc(category.FieldWeight))
	}, 10)
	if err != nil {
		return nil, fmt.Errorf("获取热门版块失败: %w", err)
	}

	// Convert format | 转换格式
	categories := make([]schema.PopularCategory, len(popularCategories))
	for i, c := range popularCategories {
		// Get post count for this category | 获取该版块的帖子数量
		postCount, err := s.postRepo.CountByCategoryID(ctx, c.ID)
		if err != nil {
			postCount = 0
		}

		categories[i] = schema.PopularCategory{
			ID:          c.ID,
			Name:        c.Name,
			PostCount:   postCount,
			Description: c.Description,
		}
	}

	return &schema.PopularCategoriesResponse{
		Categories: categories,
	}, nil
}
