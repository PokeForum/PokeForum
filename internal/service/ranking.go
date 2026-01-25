package service

import (
	"context"
	"errors"
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

// timeRangeAll Time range constant: All-time leaderboard | 时间范围常量：总榜
const timeRangeAll = "all"

// IRankingService Ranking service interface | 排行榜服务接口
type IRankingService interface {
	// GetReadingRanking Get reading leaderboard | 获取阅读排行榜
	// Get list of posts with highest view count based on time range | 根据时间范围获取阅读量最高的帖子列表
	GetReadingRanking(ctx context.Context, req schema.UserRankingListRequest) (*schema.UserRankingListResponse, error)

	// GetCommentRanking Get comment leaderboard | 获取评论排行榜
	// Get list of users with most comments based on time range | 根据时间范围获取评论数最多的用户列表
	GetCommentRanking(ctx context.Context, req schema.UserRankingListRequest) (*schema.UserRankingListResponse, error)
}

// RankingService Ranking service implementation | 排行榜服务实现
type RankingService struct {
	db           *ent.Client
	userRepo     repository.IUserRepository
	categoryRepo repository.ICategoryRepository
	cache        cache.ICacheService
	log          *zap.Logger
}

// NewRankingService Create ranking service instance | 创建排行榜服务实例
func NewRankingService(db *ent.Client, repos *repository.Repositories, cache cache.ICacheService, log *zap.Logger) IRankingService {
	return &RankingService{
		db:           db,
		userRepo:     repos.User,
		categoryRepo: repos.Category,
		cache:        cache,
		log:          log,
	}
}

// GetReadingRanking Get reading leaderboard | 获取阅读排行榜
func (s *RankingService) GetReadingRanking(ctx context.Context, req schema.UserRankingListRequest) (*schema.UserRankingListResponse, error) {
	// Add tracing | 添加链路追踪
	s.log.Info("获取阅读排行榜", tracing.WithTraceIDField(ctx))

	// Validate request parameters | 验证请求参数
	if req.Type != "reading" {
		return nil, errors.New("排行榜类型错误")
	}

	// Calculate time range | 计算时间范围
	startTime, err := s.calculateTimeRange(req.TimeRange)
	if err != nil {
		s.log.Error("计算时间范围失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("时间范围参数错误: %w", err)
	}

	// Calculate pagination parameters | 计算分页参数
	offset := (req.Page - 1) * req.PageSize

	// Build query conditions | 构建查询条件
	query := s.db.Post.Query().
		Where(
			post.StatusEQ("published"), // 只查询已发布的帖子
		).
		Order(ent.Desc(post.FieldViewCount)) // 按阅读数降序排列

	// If not all-time leaderboard, add time range filter | 如果不是总榜，添加时间范围过滤
	if req.TimeRange != timeRangeAll {
		query = query.Where(
			post.CreatedAtGTE(startTime),
		)
	}

	// Get total count | 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.log.Error("查询阅读排行榜总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	// Get paginated data | 获取分页数据
	posts, err := query.
		Offset(offset).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.log.Error("查询阅读排行榜数据失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	// Collect user IDs and category IDs | 收集用户ID和版块ID
	userIDs := make(map[int]bool)
	categoryIDs := make(map[int]bool)
	for _, p := range posts {
		userIDs[p.UserID] = true
		categoryIDs[p.CategoryID] = true
	}

	// 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername})
	if err != nil {
		s.log.Warn("批量查询用户信息失败", zap.Error(err))
	}
	userMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// 批量查询版块信息
	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id)
	}
	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, categoryIDList, []string{category.FieldID, category.FieldName})
	if err != nil {
		s.log.Warn("批量查询版块信息失败", zap.Error(err))
	}
	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	// Build response data | 构建响应数据
	items := make([]schema.RankingItem, 0, len(posts))
	for i, p := range posts {
		item := schema.RankingItem{
			Rank:      offset + i + 1,
			PostID:    &p.ID,
			PostTitle: &p.Title,
			Count:     p.ViewCount,
			CreatedAt: p.CreatedAt.Format(time_tools.DateTimeFormat),
		}
		items = append(items, item)
	}

	// Calculate total pages | 计算总页数
	totalPages := (total + req.PageSize - 1) / req.PageSize

	// Build response | 构建响应
	response := &schema.UserRankingListResponse{
		Type:       req.Type,
		TimeRange:  req.TimeRange,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		Items:      items,
	}

	s.log.Info("获取阅读排行榜成功",
		zap.Int("total", total),
		zap.Int("page", req.Page),
		zap.String("time_range", req.TimeRange),
		tracing.WithTraceIDField(ctx))

	return response, nil
}

// GetCommentRanking Get comment leaderboard | 获取评论排行榜
func (s *RankingService) GetCommentRanking(ctx context.Context, req schema.UserRankingListRequest) (*schema.UserRankingListResponse, error) {
	// Add tracing | 添加链路追踪
	s.log.Info("获取评论排行榜", tracing.WithTraceIDField(ctx))

	// Validate request parameters | 验证请求参数
	if req.Type != "comment" {
		return nil, errors.New("排行榜类型错误")
	}

	// Calculate time range | 计算时间范围
	startTime, err := s.calculateTimeRange(req.TimeRange)
	if err != nil {
		s.log.Error("计算时间范围失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("时间范围参数错误: %w", err)
	}

	// Calculate pagination parameters | 计算分页参数
	offset := (req.Page - 1) * req.PageSize

	// Build query conditions - Group by user and count comments | 构建查询条件 - 按用户分组统计评论数
	query := s.db.Comment.Query()

	// If not all-time leaderboard, add time range filter | 如果不是总榜，添加时间范围过滤
	if req.TimeRange != timeRangeAll {
		query = query.Where(
			comment.CreatedAtGTE(startTime),
		)
	}

	// Get all comments that meet the criteria | 获取所有符合条件的评论
	comments, err := query.All(ctx)
	if err != nil {
		s.log.Error("查询评论排行榜数据失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	// Collect user IDs | 收集用户ID
	userIDs := make(map[int]bool)
	for _, c := range comments {
		userIDs[c.UserID] = true
	}

	// 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar, user.FieldCreatedAt})
	if err != nil {
		s.log.Warn("批量查询用户信息失败", zap.Error(err))
	}
	userMap := make(map[int]*ent.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// Count comments by user | 按用户统计评论数
	userCommentCount := make(map[int]*schema.CommentRankingItem)
	for _, c := range comments {
		userID := c.UserID
		userInfo := userMap[userID]
		if userInfo == nil {
			continue
		}

		if item, exists := userCommentCount[userID]; exists {
			// Accumulate comment count and like count | 累加评论数和点赞数
			item.TotalComments++
			item.TotalLikes += c.LikeCount
		} else {
			// Create new statistics item | 创建新的统计项
			userCommentCount[userID] = &schema.CommentRankingItem{
				UserID:        userInfo.ID,
				Username:      userInfo.Username,
				Avatar:        userInfo.Avatar,
				TotalComments: 1,
				TotalLikes:    c.LikeCount,
				RegisteredAt:  userInfo.CreatedAt.Format(time_tools.DateTimeFormat),
			}
		}
	}

	// Convert to slice and sort (by comment count descending) | 转换为切片并排序（按评论数降序）
	commentItems := make([]schema.CommentRankingItem, 0, len(userCommentCount))
	for _, item := range userCommentCount {
		commentItems = append(commentItems, *item)
	}

	// Simple sort (by comment count descending, if same then by like count descending) | 简单排序（按评论数降序，如果评论数相同则按点赞数降序）
	for i := 0; i < len(commentItems)-1; i++ {
		for j := i + 1; j < len(commentItems); j++ {
			if commentItems[i].TotalComments < commentItems[j].TotalComments ||
				(commentItems[i].TotalComments == commentItems[j].TotalComments && commentItems[i].TotalLikes < commentItems[j].TotalLikes) {
				commentItems[i], commentItems[j] = commentItems[j], commentItems[i]
			}
		}
	}

	// Get total count | 获取总数
	total := len(commentItems)

	// Pagination handling | 分页处理
	start := offset
	end := offset + req.PageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	// Build response data | 构建响应数据
	items := make([]schema.RankingItem, 0, end-start)
	for i := start; i < end; i++ {
		item := schema.RankingItem{
			Rank:      i + 1,
			UserID:    &commentItems[i].UserID,
			Username:  &commentItems[i].Username,
			Count:     commentItems[i].TotalComments,
			CreatedAt: commentItems[i].RegisteredAt,
		}
		items = append(items, item)
	}

	// Calculate total pages | 计算总页数
	totalPages := (total + req.PageSize - 1) / req.PageSize

	// Build response | 构建响应
	response := &schema.UserRankingListResponse{
		Type:       req.Type,
		TimeRange:  req.TimeRange,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		Items:      items,
	}

	s.log.Info("获取评论排行榜成功",
		zap.Int("total", total),
		zap.Int("page", req.Page),
		zap.String("time_range", req.TimeRange),
		tracing.WithTraceIDField(ctx))

	return response, nil
}

// calculateTimeRange Calculate time range | 计算时间范围
func (s *RankingService) calculateTimeRange(timeRange string) (time.Time, error) {
	now := time.Now()

	switch timeRange {
	case timeRangeAll:
		// All-time leaderboard has no time limit, return an early time | 总榜不限制时间范围，返回一个很早的时间
		return time.Date(2020, 1, 1, 0, 0, 0, 0, now.Location()), nil
	case "week":
		// Weekly leaderboard: start from 7 days ago | 周榜：从7天前开始
		return now.AddDate(0, 0, -7), nil
	case "month":
		// Monthly leaderboard: start from 30 days ago | 月榜：从30天前开始
		return now.AddDate(0, 0, -30), nil
	default:
		return time.Time{}, errors.New("不支持的时间范围")
	}
}
