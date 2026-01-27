package service

import (
	"context"
	"encoding/json"
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
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// 排行榜常量
const (
	timeRangeAll = "all"

	// 固定分页参数
	rankingPage     = 1
	rankingPageSize = 100

	// 缓存过期时间（20分钟，比定时任务间隔稍长）
	rankingCacheTTL = 20 * 60

	// 缓存键前缀
	cacheKeyReadingRanking      = "ranking:reading:"
	cacheKeyPostCountRanking    = "ranking:post_count:"
	cacheKeyCommentCountRanking = "ranking:comment_count:"
	cacheKeyFollowerRanking     = "ranking:follower:"
	cacheKeyPointsRanking       = "ranking:points:"
	cacheKeyCurrencyRanking     = "ranking:currency:"
)

// IRankingService Ranking service interface | 排行榜服务接口
type IRankingService interface {
	// GetReadingRanking 获取阅读排行榜
	GetReadingRanking(ctx context.Context, req schema.RankingRequest) (*schema.ReadingRankingResponse, error)

	// GetCommentRanking 获取评论排行榜
	GetCommentRanking(ctx context.Context, req schema.RankingRequest) (*schema.CommentRankingResponse, error)

	// GetPostCountRanking 获取帖子数排行榜
	GetPostCountRanking(ctx context.Context, req schema.RankingRequest) (*schema.PostCountRankingResponse, error)

	// GetCommentCountRanking 获取评论数排行榜
	GetCommentCountRanking(ctx context.Context, req schema.RankingRequest) (*schema.CommentCountRankingResponse, error)

	// GetFollowerRanking 获取名人榜（被关注数）
	GetFollowerRanking(ctx context.Context, req schema.RankingRequest) (*schema.FollowerRankingResponse, error)

	// GetPointsRanking 获取积分榜
	GetPointsRanking(ctx context.Context, req schema.RankingRequest) (*schema.PointsRankingResponse, error)

	// GetCurrencyRanking 获取财富榜（货币）
	GetCurrencyRanking(ctx context.Context, req schema.RankingRequest) (*schema.CurrencyRankingResponse, error)
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

// GetReadingRanking 获取阅读排行榜（从缓存读取）
func (s *RankingService) GetReadingRanking(ctx context.Context, req schema.RankingRequest) (*schema.ReadingRankingResponse, error) {
	cacheKey := cacheKeyReadingRanking + req.TimeRange

	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var response schema.ReadingRankingResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}

	// 缓存未命中，返回空数据（等待定时任务更新）
	return &schema.ReadingRankingResponse{
		TimeRange:  req.TimeRange,
		Total:      0,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: 0,
		Items:      []schema.ReadingRankingItem{},
	}, nil
}

// refreshReadingRanking 刷新阅读排行榜缓存
func (s *RankingService) refreshReadingRanking(ctx context.Context, timeRange string) error {
	startTime, err := s.calculateTimeRange(timeRange)
	if err != nil {
		return fmt.Errorf("计算时间范围失败: %w", err)
	}

	offset := (rankingPage - 1) * rankingPageSize

	query := s.db.Post.Query().
		Where(post.StatusIn(post.StatusNormal, post.StatusLocked)).
		Order(ent.Desc(post.FieldViewCount))

	if timeRange != timeRangeAll {
		query = query.Where(post.CreatedAtGTE(startTime))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return fmt.Errorf("查询总数失败: %w", err)
	}

	posts, err := query.Offset(offset).Limit(rankingPageSize).All(ctx)
	if err != nil {
		return fmt.Errorf("查询数据失败: %w", err)
	}

	userIDs := make(map[int]bool)
	categoryIDs := make(map[int]bool)
	for _, p := range posts {
		userIDs[p.UserID] = true
		categoryIDs[p.CategoryID] = true
	}

	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername})
	if err != nil {
		s.log.Error("获取用户信息失败", zap.Error(err))
		return err
	}
	userMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id)
	}
	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, categoryIDList, []string{category.FieldID, category.FieldName})
	if err != nil {
		s.log.Error("获取分类信息失败", zap.Error(err))
		return err
	}
	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	items := make([]schema.ReadingRankingItem, 0, len(posts))
	for i, p := range posts {
		items = append(items, schema.ReadingRankingItem{
			Rank:           offset + i + 1,
			PostID:         p.ID,
			PostTitle:      p.Title,
			CategoryID:     p.CategoryID,
			CategoryName:   categoryMap[p.CategoryID],
			AuthorUsername: userMap[p.UserID],
			ViewCount:      p.ViewCount,
			LikeCount:      p.LikeCount,
			CreatedAt:      p.CreatedAt.Format(time_tools.DateTimeFormat),
		})
	}

	totalPages := (total + rankingPageSize - 1) / rankingPageSize

	response := &schema.ReadingRankingResponse{
		TimeRange:  timeRange,
		Total:      total,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: totalPages,
		Items:      items,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	cacheKey := cacheKeyReadingRanking + timeRange
	return s.cache.SetEx(ctx, cacheKey, string(data), rankingCacheTTL)
}

// GetCommentRanking 获取评论排行榜（已废弃，保留兼容）
func (s *RankingService) GetCommentRanking(ctx context.Context, req schema.RankingRequest) (*schema.CommentRankingResponse, error) {
	result, err := s.GetCommentCountRanking(ctx, req)
	if err != nil {
		return nil, err
	}

	items := make([]schema.CommentRankingItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = schema.CommentRankingItem(item)
	}

	response := &schema.CommentRankingResponse{
		TimeRange:  result.TimeRange,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
		Items:      items,
	}
	return response, nil
}

// GetPostCountRanking 获取帖子数排行榜（从缓存读取）
func (s *RankingService) GetPostCountRanking(ctx context.Context, req schema.RankingRequest) (*schema.PostCountRankingResponse, error) {
	cacheKey := cacheKeyPostCountRanking + req.TimeRange

	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var response schema.PostCountRankingResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}

	return &schema.PostCountRankingResponse{
		TimeRange:  req.TimeRange,
		Total:      0,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: 0,
		Items:      []schema.PostCountRankingItem{},
	}, nil
}

// refreshPostCountRanking 刷新帖子数排行榜缓存
func (s *RankingService) refreshPostCountRanking(ctx context.Context, timeRange string) error {
	startTime, err := s.calculateTimeRange(timeRange)
	if err != nil {
		return fmt.Errorf("计算时间范围失败: %w", err)
	}

	query := s.db.Post.Query().Where(post.StatusIn(post.StatusNormal, post.StatusLocked))
	if timeRange != timeRangeAll {
		query = query.Where(post.CreatedAtGTE(startTime))
	}

	posts, err := query.All(ctx)
	if err != nil {
		return fmt.Errorf("查询帖子数据失败: %w", err)
	}

	userStats := make(map[int]*schema.PostCountRankingItem)
	userIDs := make(map[int]bool)
	for _, p := range posts {
		userIDs[p.UserID] = true
		if item, exists := userStats[p.UserID]; exists {
			item.TotalPosts++
			item.TotalViews += p.ViewCount
		} else {
			userStats[p.UserID] = &schema.PostCountRankingItem{
				UserID:     p.UserID,
				TotalPosts: 1,
				TotalViews: p.ViewCount,
			}
		}
	}

	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar, user.FieldCreatedAt})
	if err != nil {
		s.log.Error("获取用户信息失败", zap.Error(err))
		return err
	}
	for _, u := range users {
		if item, exists := userStats[u.ID]; exists {
			item.Username = u.Username
			item.Avatar = u.Avatar
			item.RegisteredAt = u.CreatedAt.Format(time_tools.DateTimeFormat)
		}
	}

	items := make([]schema.PostCountRankingItem, 0, len(userStats))
	for _, item := range userStats {
		items = append(items, *item)
	}

	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].TotalPosts < items[j].TotalPosts ||
				(items[i].TotalPosts == items[j].TotalPosts && items[i].TotalViews < items[j].TotalViews) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	total := len(items)
	end := rankingPageSize
	if end > total {
		end = total
	}

	resultItems := make([]schema.PostCountRankingItem, 0, end)
	for i := 0; i < end; i++ {
		items[i].Rank = i + 1
		resultItems = append(resultItems, items[i])
	}

	totalPages := (total + rankingPageSize - 1) / rankingPageSize

	response := &schema.PostCountRankingResponse{
		TimeRange:  timeRange,
		Total:      total,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: totalPages,
		Items:      resultItems,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	cacheKey := cacheKeyPostCountRanking + timeRange
	return s.cache.SetEx(ctx, cacheKey, string(data), rankingCacheTTL)
}

// GetCommentCountRanking 获取评论数排行榜（从缓存读取）
func (s *RankingService) GetCommentCountRanking(ctx context.Context, req schema.RankingRequest) (*schema.CommentCountRankingResponse, error) {
	cacheKey := cacheKeyCommentCountRanking + req.TimeRange

	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var response schema.CommentCountRankingResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}

	return &schema.CommentCountRankingResponse{
		TimeRange:  req.TimeRange,
		Total:      0,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: 0,
		Items:      []schema.CommentCountRankingItem{},
	}, nil
}

// refreshCommentCountRanking 刷新评论数排行榜缓存
func (s *RankingService) refreshCommentCountRanking(ctx context.Context, timeRange string) error {
	startTime, err := s.calculateTimeRange(timeRange)
	if err != nil {
		return fmt.Errorf("计算时间范围失败: %w", err)
	}

	query := s.db.Comment.Query()
	if timeRange != timeRangeAll {
		query = query.Where(comment.CreatedAtGTE(startTime))
	}

	comments, err := query.All(ctx)
	if err != nil {
		return fmt.Errorf("查询评论数据失败: %w", err)
	}

	userStats := make(map[int]*schema.CommentCountRankingItem)
	userIDs := make(map[int]bool)
	for _, c := range comments {
		userIDs[c.UserID] = true
		if item, exists := userStats[c.UserID]; exists {
			item.TotalComments++
			item.TotalLikes += c.LikeCount
		} else {
			userStats[c.UserID] = &schema.CommentCountRankingItem{
				UserID:        c.UserID,
				TotalComments: 1,
				TotalLikes:    c.LikeCount,
			}
		}
	}

	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar, user.FieldCreatedAt})
	if err != nil {
		s.log.Error("获取用户信息失败", zap.Error(err))
		return err
	}
	for _, u := range users {
		if item, exists := userStats[u.ID]; exists {
			item.Username = u.Username
			item.Avatar = u.Avatar
			item.RegisteredAt = u.CreatedAt.Format(time_tools.DateTimeFormat)
		}
	}

	items := make([]schema.CommentCountRankingItem, 0, len(userStats))
	for _, item := range userStats {
		items = append(items, *item)
	}

	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].TotalComments < items[j].TotalComments ||
				(items[i].TotalComments == items[j].TotalComments && items[i].TotalLikes < items[j].TotalLikes) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	total := len(items)
	end := rankingPageSize
	if end > total {
		end = total
	}

	resultItems := make([]schema.CommentCountRankingItem, 0, end)
	for i := 0; i < end; i++ {
		items[i].Rank = i + 1
		resultItems = append(resultItems, items[i])
	}

	totalPages := (total + rankingPageSize - 1) / rankingPageSize

	response := &schema.CommentCountRankingResponse{
		TimeRange:  timeRange,
		Total:      total,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: totalPages,
		Items:      resultItems,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	cacheKey := cacheKeyCommentCountRanking + timeRange
	return s.cache.SetEx(ctx, cacheKey, string(data), rankingCacheTTL)
}

// GetFollowerRanking 获取名人榜（从缓存读取）
func (s *RankingService) GetFollowerRanking(ctx context.Context, req schema.RankingRequest) (*schema.FollowerRankingResponse, error) {
	cacheKey := cacheKeyFollowerRanking + req.TimeRange

	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var response schema.FollowerRankingResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}

	return &schema.FollowerRankingResponse{
		TimeRange:  req.TimeRange,
		Total:      0,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: 0,
		Items:      []schema.FollowerRankingItem{},
	}, nil
}

// refreshFollowerRanking 刷新名人榜缓存
func (s *RankingService) refreshFollowerRanking(ctx context.Context, timeRange string) error {
	follows, err := s.db.UserFollow.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("查询关注数据失败: %w", err)
	}

	followerCount := make(map[int]int)
	userIDs := make(map[int]bool)
	for _, f := range follows {
		followerCount[f.FollowingID]++
		userIDs[f.FollowingID] = true
	}

	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar, user.FieldCreatedAt})
	if err != nil {
		s.log.Error("获取用户信息失败", zap.Error(err))
		return err
	}

	postCounts := make(map[int]int)
	for id := range userIDs {
		count, err := s.db.Post.Query().Where(post.UserIDEQ(id), post.StatusIn(post.StatusNormal, post.StatusLocked)).Count(ctx)
		if err != nil {
			s.log.Error("查询帖子数量失败", zap.Error(err))
			return err
		}
		postCounts[id] = count
	}

	items := make([]schema.FollowerRankingItem, 0, len(users))
	for _, u := range users {
		items = append(items, schema.FollowerRankingItem{
			UserID:         u.ID,
			Username:       u.Username,
			Avatar:         u.Avatar,
			TotalFollowers: followerCount[u.ID],
			TotalPosts:     postCounts[u.ID],
			RegisteredAt:   u.CreatedAt.Format(time_tools.DateTimeFormat),
		})
	}

	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].TotalFollowers < items[j].TotalFollowers {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	total := len(items)
	end := rankingPageSize
	if end > total {
		end = total
	}

	resultItems := make([]schema.FollowerRankingItem, 0, end)
	for i := 0; i < end; i++ {
		items[i].Rank = i + 1
		resultItems = append(resultItems, items[i])
	}

	totalPages := (total + rankingPageSize - 1) / rankingPageSize

	response := &schema.FollowerRankingResponse{
		TimeRange:  timeRange,
		Total:      total,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: totalPages,
		Items:      resultItems,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	cacheKey := cacheKeyFollowerRanking + timeRange
	return s.cache.SetEx(ctx, cacheKey, string(data), rankingCacheTTL)
}

// GetPointsRanking 获取积分榜（从缓存读取）
func (s *RankingService) GetPointsRanking(ctx context.Context, req schema.RankingRequest) (*schema.PointsRankingResponse, error) {
	cacheKey := cacheKeyPointsRanking + req.TimeRange

	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var response schema.PointsRankingResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}

	return &schema.PointsRankingResponse{
		TimeRange:  req.TimeRange,
		Total:      0,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: 0,
		Items:      []schema.PointsRankingItem{},
	}, nil
}

// refreshPointsRanking 刷新积分榜缓存
func (s *RankingService) refreshPointsRanking(ctx context.Context, timeRange string) error {
	query := s.db.User.Query().Order(ent.Desc(user.FieldPoints))

	total, err := query.Count(ctx)
	if err != nil {
		return fmt.Errorf("查询积分榜总数失败: %w", err)
	}

	users, err := query.Limit(rankingPageSize).All(ctx)
	if err != nil {
		return fmt.Errorf("查询积分榜数据失败: %w", err)
	}

	items := make([]schema.PointsRankingItem, 0, len(users))
	for i, u := range users {
		items = append(items, schema.PointsRankingItem{
			Rank:         i + 1,
			UserID:       u.ID,
			Username:     u.Username,
			Avatar:       u.Avatar,
			Points:       u.Points,
			Experience:   u.Experience,
			RegisteredAt: u.CreatedAt.Format(time_tools.DateTimeFormat),
		})
	}

	totalPages := (total + rankingPageSize - 1) / rankingPageSize

	response := &schema.PointsRankingResponse{
		TimeRange:  timeRange,
		Total:      total,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: totalPages,
		Items:      items,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	cacheKey := cacheKeyPointsRanking + timeRange
	return s.cache.SetEx(ctx, cacheKey, string(data), rankingCacheTTL)
}

// GetCurrencyRanking 获取财富榜（从缓存读取）
func (s *RankingService) GetCurrencyRanking(ctx context.Context, req schema.RankingRequest) (*schema.CurrencyRankingResponse, error) {
	cacheKey := cacheKeyCurrencyRanking + req.TimeRange

	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var response schema.CurrencyRankingResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}

	return &schema.CurrencyRankingResponse{
		TimeRange:  req.TimeRange,
		Total:      0,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: 0,
		Items:      []schema.CurrencyRankingItem{},
	}, nil
}

// refreshCurrencyRanking 刷新财富榜缓存
func (s *RankingService) refreshCurrencyRanking(ctx context.Context, timeRange string) error {
	query := s.db.User.Query().Order(ent.Desc(user.FieldCurrency))

	total, err := query.Count(ctx)
	if err != nil {
		return fmt.Errorf("查询财富榜总数失败: %w", err)
	}

	users, err := query.Limit(rankingPageSize).All(ctx)
	if err != nil {
		return fmt.Errorf("查询财富榜数据失败: %w", err)
	}

	items := make([]schema.CurrencyRankingItem, 0, len(users))
	for i, u := range users {
		items = append(items, schema.CurrencyRankingItem{
			Rank:         i + 1,
			UserID:       u.ID,
			Username:     u.Username,
			Avatar:       u.Avatar,
			Currency:     u.Currency,
			Points:       u.Points,
			RegisteredAt: u.CreatedAt.Format(time_tools.DateTimeFormat),
		})
	}

	totalPages := (total + rankingPageSize - 1) / rankingPageSize

	response := &schema.CurrencyRankingResponse{
		TimeRange:  timeRange,
		Total:      total,
		Page:       rankingPage,
		PageSize:   rankingPageSize,
		TotalPages: totalPages,
		Items:      items,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	cacheKey := cacheKeyCurrencyRanking + timeRange
	return s.cache.SetEx(ctx, cacheKey, string(data), rankingCacheTTL)
}

// RefreshAllRankings 刷新所有排行榜缓存（供定时任务调用）
func (s *RankingService) RefreshAllRankings(ctx context.Context) error {
	timeRanges := []string{"all", "month", "week"}

	for _, tr := range timeRanges {
		if err := s.refreshReadingRanking(ctx, tr); err != nil {
			s.log.Error("刷新阅读排行榜失败", zap.String("time_range", tr), zap.Error(err))
		}
		if err := s.refreshPostCountRanking(ctx, tr); err != nil {
			s.log.Error("刷新帖子数排行榜失败", zap.String("time_range", tr), zap.Error(err))
		}
		if err := s.refreshCommentCountRanking(ctx, tr); err != nil {
			s.log.Error("刷新评论数排行榜失败", zap.String("time_range", tr), zap.Error(err))
		}
		if err := s.refreshFollowerRanking(ctx, tr); err != nil {
			s.log.Error("刷新名人榜失败", zap.String("time_range", tr), zap.Error(err))
		}
		if err := s.refreshPointsRanking(ctx, tr); err != nil {
			s.log.Error("刷新积分榜失败", zap.String("time_range", tr), zap.Error(err))
		}
		if err := s.refreshCurrencyRanking(ctx, tr); err != nil {
			s.log.Error("刷新财富榜失败", zap.String("time_range", tr), zap.Error(err))
		}
	}

	s.log.Info("所有排行榜缓存刷新完成")
	return nil
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
