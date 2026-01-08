package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/postaction"
	"github.com/PokeForum/PokeForum/ent/user"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/stats"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IPostService Post service interface | 帖子服务接口
type IPostService interface {
	// CreatePost Create a post | 创建帖子
	CreatePost(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error)
	// SaveDraft Save a draft | 保存草稿
	SaveDraft(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error)
	// UpdatePost Update a post | 更新帖子
	UpdatePost(ctx context.Context, userID int, req schema.UserPostUpdateRequest) (*schema.UserPostUpdateResponse, error)
	// SetPostPrivate Set post as private | 设置帖子私有
	SetPostPrivate(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error)
	// LikePost Like a post | 点赞帖子
	LikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error)
	// DislikePost Dislike a post | 点踩帖子
	DislikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error)
	// FavoritePost Favorite a post | 收藏帖子
	FavoritePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error)
	// GetPostList Get post list | 获取帖子列表
	GetPostList(ctx context.Context, req schema.UserPostListRequest) (*schema.UserPostListResponse, error)
	// GetPostDetail Get post detail | 获取帖子详情
	GetPostDetail(ctx context.Context, req schema.UserPostDetailRequest) (*schema.UserPostDetailResponse, error)
	// CheckEditPermission Check edit permission (can be operated once every three minutes) | 检查编辑权限（每三分钟可操作一次）
	CheckEditPermission(ctx context.Context, userID, postID int) (bool, error)
	// CheckPrivatePermission Check private permission (can be operated once every three days) | 检查私有权限（每三日可操作一次）
	CheckPrivatePermission(ctx context.Context, userID, postID int) (bool, error)
}

// PostService Post service implementation | 帖子服务实现
type PostService struct {
	db               *ent.Client
	postRepo         repository.IPostRepository
	categoryRepo     repository.ICategoryRepository
	userRepo         repository.IUserRepository
	postActionRepo   repository.IPostActionRepository
	cache            cache.ICacheService
	logger           *zap.Logger
	postStatsService IPostStatsService
	settingsService  ISettingsService
}

// NewPostService Create a post service instance | 创建帖子服务实例
func NewPostService(db *ent.Client, repos *repository.Repositories, cacheService cache.ICacheService, logger *zap.Logger, postStatsService IPostStatsService, settingsService ISettingsService) IPostService {
	return &PostService{
		db:               db,
		postRepo:         repos.Post,
		categoryRepo:     repos.Category,
		userRepo:         repos.User,
		postActionRepo:   repos.PostAction,
		cache:            cacheService,
		logger:           logger,
		postStatsService: postStatsService,
		settingsService:  settingsService,
	}
}

// CreatePost Create a post | 创建帖子
func (s *PostService) CreatePost(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error) {
	s.logger.Info("创建帖子", zap.Int("user_id", userID), zap.Int("category_id", req.CategoryID), zap.String("title", req.Title), tracing.WithTraceIDField(ctx))

	// Check user status | 检查用户状态
	if err := s.checkUserStatus(ctx, userID); err != nil {
		return nil, err
	}

	categoryData, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	newPost, err := s.postRepo.Create(ctx, userID, req.CategoryID, req.Title, req.Content, req.ReadPermission, post.StatusNormal)
	if err != nil {
		s.logger.Error("创建帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostCreateResponse{
		ID:             newPost.ID,
		CategoryID:     newPost.CategoryID,
		CategoryName:   categoryData.Name,
		Title:          newPost.Title,
		Content:        newPost.Content,
		Username:       userData.Username,
		ReadPermission: newPost.ReadPermission,
		ViewCount:      newPost.ViewCount,
		LikeCount:      newPost.LikeCount,
		DislikeCount:   newPost.DislikeCount,
		FavoriteCount:  newPost.FavoriteCount,
		IsEssence:      newPost.IsEssence,
		IsPinned:       newPost.IsPinned,
		Status:         string(newPost.Status),
		CreatedAt:      newPost.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      newPost.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("帖子创建成功", zap.Int("post_id", newPost.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// SaveDraft Save a draft | 保存草稿
func (s *PostService) SaveDraft(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error) {
	s.logger.Info("保存草稿", zap.Int("user_id", userID), zap.Int("category_id", req.CategoryID), zap.String("title", req.Title), tracing.WithTraceIDField(ctx))

	categoryData, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	newPost, err := s.postRepo.Create(ctx, userID, req.CategoryID, req.Title, req.Content, req.ReadPermission, post.StatusDraft)
	if err != nil {
		s.logger.Error("保存草稿失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostCreateResponse{
		ID:             newPost.ID,
		CategoryID:     newPost.CategoryID,
		CategoryName:   categoryData.Name,
		Title:          newPost.Title,
		Content:        newPost.Content,
		Username:       userData.Username,
		ReadPermission: newPost.ReadPermission,
		ViewCount:      newPost.ViewCount,
		LikeCount:      newPost.LikeCount,
		DislikeCount:   newPost.DislikeCount,
		FavoriteCount:  newPost.FavoriteCount,
		IsEssence:      newPost.IsEssence,
		IsPinned:       newPost.IsPinned,
		Status:         string(newPost.Status),
		CreatedAt:      newPost.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      newPost.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("草稿保存成功", zap.Int("post_id", newPost.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// UpdatePost Update a post | 更新帖子
func (s *PostService) UpdatePost(ctx context.Context, userID int, req schema.UserPostUpdateRequest) (*schema.UserPostUpdateResponse, error) {
	s.logger.Info("更新帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// Check user status | 检查用户状态
	if err := s.checkUserStatus(ctx, userID); err != nil {
		return nil, err
	}

	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	if postData.UserID != userID {
		return nil, errors.New("您不是该帖子的作者")
	}

	canEdit, err := s.CheckEditPermission(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, errors.New("帖子编辑过于频繁，请三分钟后再试")
	}

	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	updatedPost, err := s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetTitle(req.Title).
			SetContent(req.Content).
			SetReadPermission(req.ReadPermission)
	})
	if err != nil {
		s.logger.Error("更新帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	categoryName := ""
	categoryData, err := s.categoryRepo.GetByID(ctx, updatedPost.CategoryID)
	if err == nil {
		categoryName = categoryData.Name
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostUpdateResponse{
		ID:             updatedPost.ID,
		CategoryID:     updatedPost.CategoryID,
		CategoryName:   categoryName,
		Title:          updatedPost.Title,
		Content:        updatedPost.Content,
		Username:       userData.Username,
		ReadPermission: updatedPost.ReadPermission,
		ViewCount:      updatedPost.ViewCount,
		LikeCount:      updatedPost.LikeCount,
		DislikeCount:   updatedPost.DislikeCount,
		FavoriteCount:  updatedPost.FavoriteCount,
		IsEssence:      updatedPost.IsEssence,
		IsPinned:       updatedPost.IsPinned,
		Status:         string(updatedPost.Status),
		CreatedAt:      updatedPost.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      updatedPost.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("帖子更新成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// SetPostPrivate Set post as private | 设置帖子私有
func (s *PostService) SetPostPrivate(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("设置帖子私有", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	if postData.UserID != userID {
		return nil, errors.New("您不是该帖子的作者")
	}

	canSetPrivate, err := s.CheckPrivatePermission(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	if !canSetPrivate {
		return nil, errors.New("帖子私有设置过于频繁，请三天后再试")
	}

	targetStatus := post.StatusPrivate
	if postData.Status == post.StatusPrivate {
		targetStatus = post.StatusNormal
	}

	err = s.postRepo.UpdateStatus(ctx, req.ID, targetStatus)
	if err != nil {
		s.logger.Error("设置帖子私有失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostActionResponse{
		Success:       true,
		LikeCount:     postData.LikeCount,
		DislikeCount:  postData.DislikeCount,
		FavoriteCount: postData.FavoriteCount,
		ActionType:    "private",
	}

	s.logger.Info("帖子私有设置成功", zap.Int("post_id", req.ID), zap.String("status", string(targetStatus)), tracing.WithTraceIDField(ctx))
	return result, nil
}

// LikePost Like a post | 点赞帖子
func (s *PostService) LikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("点赞帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// Use stats service to perform like action | 使用统计服务执行点赞操作
	action, err := s.postStatsService.PerformAction(ctx, userID, req.ID, "Like")
	if err != nil {
		s.logger.Error("点赞帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostActionResponse{
		Success:       true,
		LikeCount:     action.LikeCount,
		DislikeCount:  action.DislikeCount,
		FavoriteCount: action.FavoriteCount,
		ActionType:    "like",
	}

	s.logger.Info("帖子点赞成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// DislikePost Dislike a post | 点踩帖子
func (s *PostService) DislikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("点踩帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// Use stats service to perform dislike action | 使用统计服务执行点踩操作
	action, err := s.postStatsService.PerformAction(ctx, userID, req.ID, "Dislike")
	if err != nil {
		s.logger.Error("点踩帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostActionResponse{
		Success:       true,
		LikeCount:     action.LikeCount,
		DislikeCount:  action.DislikeCount,
		FavoriteCount: action.FavoriteCount,
		ActionType:    "dislike",
	}

	s.logger.Info("帖子点踩成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// FavoritePost Favorite a post | 收藏帖子
func (s *PostService) FavoritePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("收藏帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// First check if user has already favorited | 先检查用户是否已经收藏
	userActionStatus, err := s.postStatsService.GetUserActionStatus(ctx, userID, req.ID)
	if err != nil {
		s.logger.Error("获取用户操作状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	var postStats *stats.Stats
	var actionType string

	if userActionStatus.HasFavorited {
		// Already favorited, cancel favorite | 已经收藏,执行取消收藏
		postStats, err = s.postStatsService.CancelAction(ctx, userID, req.ID, "Favorite")
		if err != nil {
			s.logger.Error("取消收藏失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, err
		}
		actionType = "unfavorite"
		s.logger.Info("取消收藏帖子成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	} else {
		// Not favorited, perform favorite | 未收藏,执行收藏
		postStats, err = s.postStatsService.PerformAction(ctx, userID, req.ID, "Favorite")
		if err != nil {
			s.logger.Error("收藏帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, err
		}
		actionType = "favorite"
		s.logger.Info("收藏帖子成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostActionResponse{
		Success:       true,
		LikeCount:     postStats.LikeCount,
		DislikeCount:  postStats.DislikeCount,
		FavoriteCount: postStats.FavoriteCount,
		ActionType:    actionType,
	}

	return result, nil
}

// GetPostList Get post list | 获取帖子列表
func (s *PostService) GetPostList(ctx context.Context, req schema.UserPostListRequest) (*schema.UserPostListResponse, error) {
	s.logger.Info("获取帖子列表", zap.Int("category_id", req.CategoryID), zap.Int("page", req.Page), zap.Int("page_size", req.PageSize), tracing.WithTraceIDField(ctx))

	// Set default values | 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Sort == "" {
		req.Sort = "latest"
	}

	// Use repository to query posts | 使用 Repository 查询帖子列表
	posts, total, err := s.postRepo.List(ctx, repository.ListPostOptions{
		CategoryID: req.CategoryID,
		Status:     post.StatusNormal,
		SortBy:     req.Sort,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		s.logger.Error("获取帖子列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Collect user IDs and category IDs | 收集用户ID和版块ID
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
		s.logger.Warn("批量查询用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
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
		s.logger.Warn("批量查询版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
	}
	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name
	}

	// Get current user ID, 0 if not logged in | 获取当前用户ID，如果未登录则为0
	currentUserID := tracing.GetUserID(ctx)

	// Get post ID list | 获取帖子ID列表
	postIDs := make([]int, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	var userLikeStatus map[int]map[string]bool
	if currentUserID != 0 {
		actions, err := s.postActionRepo.GetUserActionsForPosts(ctx, currentUserID, postIDs)
		if err != nil {
			s.logger.Warn("查询用户点赞状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			userLikeStatus = make(map[int]map[string]bool)
		} else {
			userLikeStatus = make(map[int]map[string]bool)
			for _, action := range actions {
				if _, exists := userLikeStatus[action.PostID]; !exists {
					userLikeStatus[action.PostID] = map[string]bool{"like": false, "dislike": false}
				}
				switch action.ActionType {
				case postaction.ActionTypeLike:
					userLikeStatus[action.PostID]["like"] = true
				case postaction.ActionTypeDislike:
					userLikeStatus[action.PostID]["dislike"] = true
				}
			}
		}
	} else {
		userLikeStatus = make(map[int]map[string]bool)
	}

	// Batch get real-time stats data | 批量获取实时统计数据
	statsMap, err := s.postStatsService.GetStatsMap(ctx, postIDs)
	if err != nil {
		s.logger.Warn("获取实时统计数据失败，将使用数据库中的旧数据", zap.Error(err), tracing.WithTraceIDField(ctx))
		// On failure, don't block the process, fallback to database data | 失败时不阻断流程，降级使用数据库数据
		statsMap = make(map[int]*stats.Stats)
	}

	// Convert to response format | 转换为响应格式
	result := make([]schema.UserPostCreateResponse, len(posts))
	for i, p := range posts {
		username := userMap[p.UserID]
		categoryName := categoryMap[p.CategoryID]

		// Prefer real-time stats data | 优先使用实时统计数据
		likeCount := p.LikeCount
		dislikeCount := p.DislikeCount
		favoriteCount := p.FavoriteCount
		viewCount := p.ViewCount
		if statsData, ok := statsMap[p.ID]; ok {
			likeCount = statsData.LikeCount
			dislikeCount = statsData.DislikeCount
			favoriteCount = statsData.FavoriteCount
			viewCount = statsData.ViewCount
		}

		// Get user like status | 获取用户点赞状态
		userLiked := false
		userDisliked := false
		if status, exists := userLikeStatus[p.ID]; exists {
			userLiked = status["like"]
			userDisliked = status["dislike"]
		}

		result[i] = schema.UserPostCreateResponse{
			ID:             p.ID,
			CategoryID:     p.CategoryID,
			CategoryName:   categoryName,
			Title:          p.Title,
			Content:        p.Content,
			Username:       username,
			ReadPermission: p.ReadPermission,
			ViewCount:      viewCount,
			LikeCount:      likeCount,
			DislikeCount:   dislikeCount,
			FavoriteCount:  favoriteCount,
			UserLiked:      userLiked,
			UserDisliked:   userDisliked,
			IsEssence:      p.IsEssence,
			IsPinned:       p.IsPinned,
			Status:         string(p.Status),
			CreatedAt:      p.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:      p.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	totalPages := (total + req.PageSize - 1) / req.PageSize

	return &schema.UserPostListResponse{
		Posts:      result,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetPostDetail Get post detail | 获取帖子详情
func (s *PostService) GetPostDetail(ctx context.Context, req schema.UserPostDetailRequest) (*schema.UserPostDetailResponse, error) {
	s.logger.Info("获取帖子详情", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Update view count (use stats service to reduce database pressure) | 更新浏览数(使用统计服务,减少数据库压力)
	if err = s.postStatsService.IncrViewCount(ctx, req.ID); err != nil {
		s.logger.Warn("增加帖子浏览数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// Don't affect main process | 不影响主要流程
	}

	// Get real-time stats data | 获取实时统计数据
	statsData, err := s.postStatsService.GetStats(ctx, req.ID)
	likeCount := postData.LikeCount
	dislikeCount := postData.DislikeCount
	favoriteCount := postData.FavoriteCount
	viewCount := postData.ViewCount
	if err != nil {
		s.logger.Warn("获取实时统计数据失败，将使用数据库中的旧数据", zap.Error(err), tracing.WithTraceIDField(ctx))
	} else {
		likeCount = statsData.LikeCount
		dislikeCount = statsData.DislikeCount
		favoriteCount = statsData.FavoriteCount
		viewCount = statsData.ViewCount
	}

	username := ""
	author, err := s.userRepo.GetByID(ctx, postData.UserID)
	if err == nil {
		username = author.Username
	}

	categoryName := ""
	categoryData, err := s.categoryRepo.GetByID(ctx, postData.CategoryID)
	if err == nil {
		categoryName = categoryData.Name
	}

	// Get current user ID, 0 if not logged in | 获取当前用户ID，如果未登录则为0
	currentUserID := tracing.GetUserID(ctx)

	userLiked := false
	userDisliked := false
	userFavorite := false
	if currentUserID != 0 {
		actions, err := s.postActionRepo.GetUserActionsForPosts(ctx, currentUserID, []int{req.ID})
		if err != nil {
			s.logger.Debug("查询用户操作状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		} else {
			for _, action := range actions {
				switch action.ActionType {
				case postaction.ActionTypeLike:
					userLiked = true
				case postaction.ActionTypeDislike:
					userDisliked = true
				case postaction.ActionTypeFavorite:
					userFavorite = true
				}
			}
		}
	}

	result := &schema.UserPostDetailResponse{
		ID:             postData.ID,
		CategoryID:     postData.CategoryID,
		CategoryName:   categoryName,
		Title:          postData.Title,
		Content:        postData.Content,
		UserID:         postData.UserID,
		Username:       username,
		ReadPermission: postData.ReadPermission,
		ViewCount:      viewCount,
		LikeCount:      likeCount,
		DislikeCount:   dislikeCount,
		FavoriteCount:  favoriteCount,
		UserLiked:      userLiked,
		UserDisliked:   userDisliked,
		UserFavorited:  userFavorite,
		IsEssence:      postData.IsEssence,
		IsPinned:       postData.IsPinned,
		Status:         string(postData.Status),
		CreatedAt:      postData.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      postData.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("获取帖子详情成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// CheckEditPermission Check edit permission (can be operated once every three minutes) | 检查编辑权限（每三分钟可操作一次）
func (s *PostService) CheckEditPermission(ctx context.Context, userID, postID int) (bool, error) {
	// Generate Redis key: post:edit:limit:{userID}:{postID} | 生成Redis键名：post:edit:limit:{userID}:{postID}
	redisKey := fmt.Sprintf("post:edit:limit:%d:%d", userID, postID)

	// Check if within limit period | 检查是否在限制期内
	lastEditTime, err := s.cache.Get(ctx, redisKey)
	if err != nil {
		s.logger.Error("获取编辑限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return false, fmt.Errorf("获取编辑限制失败: %w", err)
	}

	if lastEditTime != "" {
		// Parse last edit time | 解析最后编辑时间
		lastTime, err := time.Parse(time.RFC3339, lastEditTime)
		if err != nil {
			s.logger.Error("解析最后编辑时间失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			// Parse failed, allow operation | 解析失败，允许操作
			return true, nil
		}

		// Check if within three minutes | 检查是否在三分钟内
		if time.Since(lastTime) < 3*time.Minute {
			s.logger.Warn("编辑操作过于频繁", zap.Int("user_id", userID), zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
			return false, nil
		}
	}

	// Set new edit time limit | 设置新的编辑时间限制
	currentTime := time.Now().Format(time.RFC3339)
	err = s.cache.SetEx(ctx, redisKey, currentTime, 180) // 180 seconds = 3 minutes | 180秒 = 3分钟
	if err != nil {
		s.logger.Error("设置编辑限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// Set failed, but allow operation | 设置失败，但允许操作
		return true, nil
	}

	return true, nil
}

// CheckPrivatePermission Check private permission (can be operated once every three days) | 检查私有权限（每三日可操作一次）
func (s *PostService) CheckPrivatePermission(ctx context.Context, userID, postID int) (bool, error) {
	// Generate Redis key: post:private:limit:{userID}:{postID} | 生成Redis键名：post:private:limit:{userID}:{postID}
	redisKey := fmt.Sprintf("post:private:limit:%d:%d", userID, postID)

	// Check if within limit period | 检查是否在限制期内
	lastPrivateTime, err := s.cache.Get(ctx, redisKey)
	if err != nil {
		s.logger.Error("获取私有设置限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return false, fmt.Errorf("获取私有设置限制失败: %w", err)
	}

	if lastPrivateTime != "" {
		// Parse last private setting time | 解析最后私有设置时间
		lastTime, err := time.Parse(time.RFC3339, lastPrivateTime)
		if err != nil {
			s.logger.Error("解析最后私有设置时间失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			// Parse failed, allow operation | 解析失败，允许操作
			return true, nil
		}

		// Check if within three days | 检查是否在三日内
		if time.Since(lastTime) < 3*24*time.Hour {
			s.logger.Warn("私有设置操作过于频繁", zap.Int("user_id", userID), zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
			return false, nil
		}
	}

	// Set new private setting time limit | 设置新的私有设置时间限制
	currentTime := time.Now().Format(time.RFC3339)
	err = s.cache.SetEx(ctx, redisKey, currentTime, 259200) // 259200 seconds = 3 days | 259200秒 = 3天
	if err != nil {
		s.logger.Error("设置私有限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// Set failed, but allow operation | 设置失败，但允许操作
		return true, nil
	}

	return true, nil
}

// checkUserStatus Check if user status allows operation | 检查用户状态是否允许操作
func (s *PostService) checkUserStatus(ctx context.Context, userID int) error {
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	switch userData.Status {
	case user.StatusNormal:
		// Check if email verification is required | 检查是否需要验证邮箱
		verifyEmail, err := s.settingsService.GetSettingByKey(ctx, _const.SafeVerifyEmail, "false")
		if err != nil {
			s.logger.Warn("获取邮箱验证配置失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			verifyEmail = "false"
		}
		if verifyEmail == _const.SettingBoolTrue.String() && !userData.EmailVerified {
			return errors.New("您的邮箱尚未验证，请先完成验证")
		}
		return nil
	case user.StatusRiskControl:
		// TODO: RiskControl status requires admin review before publishing, temporarily allow | RiskControl状态需要管理员审核发布，暂时放行
		return nil
	case user.StatusMute:
		return errors.New("您已被禁言，无法进行此操作")
	case user.StatusBlocked:
		return errors.New("您的账号已被封禁，无法进行此操作")
	default:
		return errors.New("账号状态异常，无法进行此操作")
	}
}
