package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// IDashboardService 仪表盘服务接口
type IDashboardService interface {
	// GetDashboardStats 获取仪表盘统计数据
	GetDashboardStats(ctx context.Context) (*schema.DashboardStatsResponse, error)
	// GetRecentActivity 获取最近活动
	GetRecentActivity(ctx context.Context) (*schema.RecentActivityResponse, error)
	// GetPopularPosts 获取热门帖子
	GetPopularPosts(ctx context.Context) (*schema.PopularPostsResponse, error)
	// GetPopularCategories 获取热门版块
	GetPopularCategories(ctx context.Context) (*schema.PopularCategoriesResponse, error)
}

// DashboardService 仪表盘服务实现
type DashboardService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewDashboardService 创建仪表盘服务实例
func NewDashboardService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IDashboardService {
	return &DashboardService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// GetDashboardStats 获取仪表盘统计数据
func (s *DashboardService) GetDashboardStats(ctx context.Context) (*schema.DashboardStatsResponse, error) {
	s.logger.Info("获取仪表盘统计数据", tracing.WithTraceIDField(ctx))

	// 获取用户统计
	userStats, err := s.getUserStats(ctx)
	if err != nil {
		s.logger.Error("获取用户统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户统计失败: %w", err)
	}

	// 获取帖子统计
	postStats, err := s.getPostStats(ctx)
	if err != nil {
		s.logger.Error("获取帖子统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子统计失败: %w", err)
	}

	// 获取评论统计
	commentStats, err := s.getCommentStats(ctx)
	if err != nil {
		s.logger.Error("获取评论统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论统计失败: %w", err)
	}

	// 获取版块统计
	categoryStats, err := s.getCategoryStats(ctx)
	if err != nil {
		s.logger.Error("获取版块统计失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块统计失败: %w", err)
	}

	// 获取系统统计
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

// getUserStats 获取用户统计
func (s *DashboardService) getUserStats(ctx context.Context) (schema.UserStats, error) {
	// 总用户数
	totalUsers, err := s.db.User.Query().Count(ctx)
	if err != nil {
		return schema.UserStats{}, err
	}

	// 活跃用户数（30天内创建的用户，简化处理）
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	activeUsers, err := s.db.User.Query().
		Where(user.CreatedAtGTE(thirtyDaysAgo)).
		Count(ctx)
	if err != nil {
		return schema.UserStats{}, err
	}

	// 新增用户数（今日）
	today := time.Now().Truncate(24 * time.Hour)
	newUsers, err := s.db.User.Query().
		Where(user.CreatedAtGTE(today)).
		Count(ctx)
	if err != nil {
		return schema.UserStats{}, err
	}

	// 在线用户数（这里简化处理，实际应该通过Redis或其他方式统计）
	onlineUsers := int64(0) // 需要根据实际在线用户统计逻辑实现

	// 被封禁用户数
	bannedUsers, err := s.db.User.Query().
		Where(user.StatusEQ(user.StatusBlocked)).
		Count(ctx)
	if err != nil {
		return schema.UserStats{}, err
	}

	// 版主数量
	moderatorCount, err := s.db.User.Query().
		Where(user.RoleEQ(user.RoleModerator)).
		Count(ctx)
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

// getPostStats 获取帖子统计
func (s *DashboardService) getPostStats(ctx context.Context) (schema.PostStats, error) {
	// 总帖子数
	totalPosts, err := s.db.Post.Query().Count(ctx)
	if err != nil {
		return schema.PostStats{}, err
	}

	// 已发布帖子数
	publishedPosts, err := s.db.Post.Query().
		Where(post.StatusEQ(post.StatusNormal)).
		Count(ctx)
	if err != nil {
		return schema.PostStats{}, err
	}

	// 草稿帖子数
	draftPosts, err := s.db.Post.Query().
		Where(post.StatusEQ(post.StatusDraft)).
		Count(ctx)
	if err != nil {
		return schema.PostStats{}, err
	}

	// 被锁定帖子数
	lockedPosts, err := s.db.Post.Query().
		Where(post.StatusEQ(post.StatusLocked)).
		Count(ctx)
	if err != nil {
		return schema.PostStats{}, err
	}

	// 精华帖子数
	essencePosts, err := s.db.Post.Query().
		Where(post.IsEssenceEQ(true)).
		Count(ctx)
	if err != nil {
		return schema.PostStats{}, err
	}

	// 置顶帖子数
	pinnedPosts, err := s.db.Post.Query().
		Where(post.IsPinnedEQ(true)).
		Count(ctx)
	if err != nil {
		return schema.PostStats{}, err
	}

	// 今日新增帖子数
	today := time.Now().Truncate(24 * time.Hour)
	todayPosts, err := s.db.Post.Query().
		Where(post.CreatedAtGTE(today)).
		Count(ctx)
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

// getCommentStats 获取评论统计
func (s *DashboardService) getCommentStats(ctx context.Context) (schema.CommentStats, error) {
	// 总评论数
	totalComments, err := s.db.Comment.Query().Count(ctx)
	if err != nil {
		return schema.CommentStats{}, err
	}

	// 精选评论数
	selectedComments, err := s.db.Comment.Query().
		Where(comment.IsSelectedEQ(true)).
		Count(ctx)
	if err != nil {
		return schema.CommentStats{}, err
	}

	// 置顶评论数
	pinnedComments, err := s.db.Comment.Query().
		Where(comment.IsPinnedEQ(true)).
		Count(ctx)
	if err != nil {
		return schema.CommentStats{}, err
	}

	// 今日新增评论数
	today := time.Now().Truncate(24 * time.Hour)
	todayComments, err := s.db.Comment.Query().
		Where(comment.CreatedAtGTE(today)).
		Count(ctx)
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

// getCategoryStats 获取版块统计
func (s *DashboardService) getCategoryStats(ctx context.Context) (schema.CategoryStats, error) {
	// 总版块数
	totalCategories, err := s.db.Category.Query().Count(ctx)
	if err != nil {
		return schema.CategoryStats{}, err
	}

	// 活跃版块数（正常状态的版块）
	activeCategories, err := s.db.Category.Query().
		Where(category.StatusEQ(category.StatusNormal)).
		Count(ctx)
	if err != nil {
		return schema.CategoryStats{}, err
	}

	// 隐藏版块数
	hiddenCategories, err := s.db.Category.Query().
		Where(category.StatusEQ(category.StatusHidden)).
		Count(ctx)
	if err != nil {
		return schema.CategoryStats{}, err
	}

	// 锁定版块数
	lockedCategories, err := s.db.Category.Query().
		Where(category.StatusEQ(category.StatusLocked)).
		Count(ctx)
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

// getSystemStats 获取系统统计
func (s *DashboardService) getSystemStats(ctx context.Context) (schema.SystemStats, error) {
	// 总浏览量（这里简化处理，实际应该从访问日志统计）
	totalViews := int64(0) // 需要根据实际浏览量统计逻辑实现

	// 今日浏览量
	todayViews := int64(0) // 需要根据实际今日浏览量统计逻辑实现

	// 总点赞数（帖子点赞数 + 评论点赞数）
	posts, err := s.db.Post.Query().All(ctx)
	if err != nil {
		return schema.SystemStats{}, err
	}

	totalPostLikes := int64(0)
	for _, p := range posts {
		totalPostLikes += int64(p.LikeCount)
	}

	comments, err := s.db.Comment.Query().All(ctx)
	if err != nil {
		return schema.SystemStats{}, err
	}

	totalCommentLikes := int64(0)
	for _, c := range comments {
		totalCommentLikes += int64(c.LikeCount)
	}

	totalLikes := totalPostLikes + totalCommentLikes

	// 今日点赞数（简化处理）
	todayLikes := int64(0) // 需要根据实际今日点赞数统计逻辑实现

	// 存储使用量（简化处理）
	storageUsed := int64(0) // 需要根据实际存储使用量统计逻辑实现

	// 数据库大小（简化处理）
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

// GetRecentActivity 获取最近活动
func (s *DashboardService) GetRecentActivity(ctx context.Context) (*schema.RecentActivityResponse, error) {
	s.logger.Info("获取最近活动", tracing.WithTraceIDField(ctx))

	// 获取最近帖子（最近10条）
	recentPosts, err := s.db.Post.Query().
		Order(ent.Desc(post.FieldCreatedAt)).
		Limit(10).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取最近帖子失败: %w", err)
	}

	// 获取最近评论（最近10条）
	recentComments, err := s.db.Comment.Query().
		Order(ent.Desc(comment.FieldCreatedAt)).
		Limit(10).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取最近评论失败: %w", err)
	}

	// 获取新用户（最近10个）
	newUsers, err := s.db.User.Query().
		Order(ent.Desc(user.FieldCreatedAt)).
		Limit(10).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取新用户失败: %w", err)
	}

	// 收集需要查询的用户ID和版块ID
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

	// 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, _ := s.db.User.Query().
		Where(user.IDIn(userIDList...)).
		Select(user.FieldID, user.FieldUsername, user.FieldAvatar).
		All(ctx)
	type userInfo struct {
		Username string
		Avatar   string
	}
	userMap := make(map[int]userInfo)
	for _, u := range users {
		userMap[u.ID] = userInfo{Username: u.Username, Avatar: u.Avatar}
	}

	// 批量查询版块信息
	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id)
	}
	categories, _ := s.db.Category.Query().
		Where(category.IDIn(categoryIDList...)).
		Select(category.FieldID, category.FieldName).
		All(ctx)
	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	// 批量查询帖子信息
	postIDList := make([]int, 0, len(postIDs))
	for id := range postIDs {
		postIDList = append(postIDList, id)
	}
	postsData, _ := s.db.Post.Query().
		Where(post.IDIn(postIDList...)).
		Select(post.FieldID, post.FieldTitle).
		All(ctx)
	postMap := make(map[int]string)
	for _, p := range postsData {
		postMap[p.ID] = p.Title
	}

	// 转换格式
	posts := make([]schema.RecentPost, len(recentPosts))
	for i, p := range recentPosts {
		ui := userMap[p.UserID]
		posts[i] = schema.RecentPost{
			ID:           p.ID,
			Title:        p.Title,
			Username:     ui.Username,
			Avatar:       ui.Avatar,
			CategoryName: categoryMap[p.CategoryID],
			CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	comments := make([]schema.RecentComment, len(recentComments))
	for i, c := range recentComments {
		// 截取评论内容前100字符
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
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	newUserList := make([]schema.NewUser, len(newUsers))
	for i, u := range newUsers {
		newUserList[i] = schema.NewUser{
			ID:        u.ID,
			Username:  u.Username,
			Avatar:    u.Avatar,
			Email:     u.Email,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &schema.RecentActivityResponse{
		RecentPosts:    posts,
		RecentComments: comments,
		NewUsers:       newUserList,
	}, nil
}

// GetPopularPosts 获取热门帖子
func (s *DashboardService) GetPopularPosts(ctx context.Context) (*schema.PopularPostsResponse, error) {
	s.logger.Info("获取热门帖子", tracing.WithTraceIDField(ctx))

	// 按浏览量排序获取热门帖子
	popularPosts, err := s.db.Post.Query().
		Order(ent.Desc(post.FieldViewCount)).
		Limit(10).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取热门帖子失败: %w", err)
	}

	// 收集需要查询的用户ID和版块ID
	userIDs := make(map[int]bool)
	categoryIDs := make(map[int]bool)
	postIDs := make([]int, len(popularPosts))
	for i, p := range popularPosts {
		userIDs[p.UserID] = true
		categoryIDs[p.CategoryID] = true
		postIDs[i] = p.ID
	}

	// 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, _ := s.db.User.Query().
		Where(user.IDIn(userIDList...)).
		Select(user.FieldID, user.FieldUsername).
		All(ctx)
	userMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// 批量查询版块信息
	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id)
	}
	categories, _ := s.db.Category.Query().
		Where(category.IDIn(categoryIDList...)).
		Select(category.FieldID, category.FieldName).
		All(ctx)
	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	// 批量查询评论数
	// 注意: 这里简化处理，实际应该使用GroupBy优化性能
	commentCountMap := make(map[int]int)
	for _, postID := range postIDs {
		count, _ := s.db.Comment.Query().
			Where(comment.PostIDEQ(postID)).
			Count(ctx)
		commentCountMap[postID] = count
	}

	// 转换格式
	posts := make([]schema.PopularPost, len(popularPosts))
	for i, p := range popularPosts {
		posts[i] = schema.PopularPost{
			ID:           p.ID,
			Title:        p.Title,
			Username:     userMap[p.UserID],
			CategoryName: categoryMap[p.CategoryID],
			ViewCount:    p.ViewCount,
			LikeCount:    p.LikeCount,
			CommentCount: commentCountMap[p.ID],
			CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &schema.PopularPostsResponse{
		Posts: posts,
	}, nil
}

// GetPopularCategories 获取热门版块
func (s *DashboardService) GetPopularCategories(ctx context.Context) (*schema.PopularCategoriesResponse, error) {
	s.logger.Info("获取热门版块", tracing.WithTraceIDField(ctx))

	// 按帖子数量排序获取热门版块
	popularCategories, err := s.db.Category.Query().
		Order(ent.Desc(category.FieldWeight)).
		Limit(10).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取热门版块失败: %w", err)
	}

	// 转换格式
	categories := make([]schema.PopularCategory, len(popularCategories))
	for i, c := range popularCategories {
		// 获取该版块的帖子数量
		postCount, err := s.db.Post.Query().
			Where(post.CategoryIDEQ(c.ID)).
			Count(ctx)
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
