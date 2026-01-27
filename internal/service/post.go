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

// NoPermissionError Custom error type for no read permission | 无阅读权限的自定义错误类型
type NoPermissionError struct {
	Reason string
}

// Error implements the error interface | 实现error接口
func (e *NoPermissionError) Error() string {
	return e.Reason
}

// IPostService Post service interface | 帖子服务接口
type IPostService interface {
	// CreatePost Create a post | 创建帖子
	CreatePost(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error)
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
	// GetDraftList Get draft list | 获取草稿列表
	GetDraftList(ctx context.Context, userID int, req schema.UserDraftListRequest) (*schema.UserPostListResponse, error)
	// SaveDraft Save a draft | 保存草稿
	SaveDraft(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error)
	// DeleteDraft Delete draft | 删除草稿
	DeleteDraft(ctx context.Context, userID int, req schema.UserDraftDeleteRequest) error
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

	// 检查版块是否为锁定状态，锁定状态不允许发帖
	if categoryData.Status == category.StatusLocked {
		s.logger.Warn("版块已锁定，不允许发帖", zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))
		return nil, errors.New("该版块已锁定，不允许发布新帖子")
	}

	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Convert read permission type to enum | 转换阅读权限类型为枚举
	readPermission := s.parseReadPermissionType(req.ReadPermissionType)

	// 如果版块是登录可见，且帖子阅读权限是 public，则自动变更为 login_required
	if categoryData.Status == category.StatusLoginRequired && readPermission == post.ReadPermissionPublic {
		readPermission = post.ReadPermissionLoginRequired
		s.logger.Info("版块为登录可见，自动将帖子阅读权限调整为login_required", zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))
	}

	newPost, err := s.postRepo.Create(ctx, userID, req.CategoryID, req.Title, req.Content, readPermission, req.ReadPermissionPoints, post.StatusNormal)
	if err != nil {
		s.logger.Error("创建帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostCreateResponse{
		ID:                   newPost.ID,
		CategoryID:           newPost.CategoryID,
		CategoryName:         categoryData.Name,
		Title:                newPost.Title,
		Content:              newPost.Content,
		Username:             userData.Username,
		ReadPermissionType:   string(newPost.ReadPermission),
		ReadPermissionPoints: newPost.ReadPermissionPoints,
		ViewCount:            newPost.ViewCount,
		LikeCount:            newPost.LikeCount,
		DislikeCount:         newPost.DislikeCount,
		FavoriteCount:        newPost.FavoriteCount,
		IsEssence:            newPost.IsEssence,
		IsPinned:             newPost.IsPinned,
		Status:               string(newPost.Status),
		CreatedAt:            newPost.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:            newPost.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("帖子创建成功", zap.Int("post_id", newPost.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// SaveDraft Save a draft | 保存草稿
func (s *PostService) SaveDraft(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error) {
	s.logger.Info("保存草稿", zap.Int("user_id", userID), zap.Int("draft_id", req.ID), zap.Int("category_id", req.CategoryID), zap.String("title", req.Title), tracing.WithTraceIDField(ctx))

	// 检查版块是否为锁定状态，锁定状态不允许保存草稿
	categoryData, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}
	if categoryData.Status == category.StatusLocked {
		s.logger.Warn("版块已锁定，不允许保存草稿", zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))
		return nil, errors.New("该版块已锁定，不允许保存草稿")
	}

	var resultPost *ent.Post

	// Convert read permission type to enum | 转换阅读权限类型为枚举
	readPermission := s.parseReadPermissionType(req.ReadPermissionType)

	// 如果版块是登录可见，且帖子阅读权限是 public，则自动变更为 login_required
	if categoryData.Status == category.StatusLoginRequired && readPermission == post.ReadPermissionPublic {
		readPermission = post.ReadPermissionLoginRequired
		s.logger.Info("版块为登录可见，自动将草稿阅读权限调整为login_required", zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))
	}

	// If ID exists, update existing draft | 如果ID存在，更新现有草稿
	if req.ID > 0 {
		// Get draft post | 获取草稿帖子
		draftPost, err := s.postRepo.GetByIDWithStatus(ctx, req.ID, post.StatusDraft)
		if err != nil {
			s.logger.Error("获取草稿失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, err
		}

		// Check ownership | 检查所有权
		if draftPost.UserID != userID {
			return nil, errors.New("您不是该草稿的作者")
		}

		// Update draft | 更新草稿
		resultPost, err = s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
			return u.SetCategoryID(req.CategoryID).
				SetTitle(req.Title).
				SetContent(req.Content).
				SetReadPermission(readPermission).
				SetReadPermissionPoints(req.ReadPermissionPoints)
		})
		if err != nil {
			s.logger.Error("更新草稿失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, err
		}
		s.logger.Info("草稿更新成功", zap.Int("draft_id", req.ID), tracing.WithTraceIDField(ctx))
	} else {
		// Create new draft | 创建新草稿
		// Check draft count limit (max 10 drafts per user) | 检查草稿数量限制（每个用户最多10篇草稿）
		draftCount, err := s.postRepo.CountByUserIDWithStatus(ctx, userID, post.StatusDraft)
		if err != nil {
			s.logger.Error("获取用户草稿数量失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, err
		}
		if draftCount >= 10 {
			return nil, errors.New("草稿数量已达上限（最多10篇），请删除或发布部分草稿后再试")
		}

		resultPost, err = s.postRepo.Create(ctx, userID, req.CategoryID, req.Title, req.Content, readPermission, req.ReadPermissionPoints, post.StatusDraft)
		if err != nil {
			s.logger.Error("保存草稿失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, err
		}
		s.logger.Info("草稿创建成功", zap.Int("draft_id", resultPost.ID), tracing.WithTraceIDField(ctx))
	}

	// Get user info | 获取用户信息
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Build response data | 构建响应数据
	result := &schema.UserPostCreateResponse{
		ID:                   resultPost.ID,
		CategoryID:           resultPost.CategoryID,
		CategoryName:         categoryData.Name,
		Title:                resultPost.Title,
		Content:              resultPost.Content,
		Username:             userData.Username,
		ReadPermissionType:   string(resultPost.ReadPermission),
		ReadPermissionPoints: resultPost.ReadPermissionPoints,
		ViewCount:            resultPost.ViewCount,
		LikeCount:            resultPost.LikeCount,
		DislikeCount:         resultPost.DislikeCount,
		FavoriteCount:        resultPost.FavoriteCount,
		IsEssence:            resultPost.IsEssence,
		IsPinned:             resultPost.IsPinned,
		Status:               string(resultPost.Status),
		CreatedAt:            resultPost.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:            resultPost.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("草稿保存成功", zap.Int("post_id", resultPost.ID), tracing.WithTraceIDField(ctx))
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

	// Check post status (Locked and Ban posts cannot be edited) | 检查帖子状态（锁定和封禁的帖子不允许编辑）
	if postData.Status == post.StatusLocked || postData.Status == post.StatusBan {
		s.logger.Warn("帖子状态不允许编辑", zap.Int("post_id", req.ID), zap.String("status", string(postData.Status)), tracing.WithTraceIDField(ctx))
		return nil, errors.New("该帖子已被锁定或封禁，无法编辑")
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

	// Convert read permission type to enum | 转换阅读权限类型为枚举
	readPermission := s.parseReadPermissionType(req.ReadPermissionType)

	// 获取版块信息，检查是否需要自动调整阅读权限
	categoryData, err := s.categoryRepo.GetByID(ctx, postData.CategoryID)
	if err != nil {
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 如果版块是登录可见，且帖子阅读权限是 public，则自动变更为 login_required
	if categoryData.Status == category.StatusLoginRequired && readPermission == post.ReadPermissionPublic {
		readPermission = post.ReadPermissionLoginRequired
		s.logger.Info("版块为登录可见，自动将帖子阅读权限调整为login_required", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	}

	updatedPost, err := s.postRepo.Update(ctx, req.ID, func(u *ent.PostUpdateOne) *ent.PostUpdateOne {
		return u.SetTitle(req.Title).
			SetContent(req.Content).
			SetReadPermission(readPermission).
			SetReadPermissionPoints(req.ReadPermissionPoints)
	})
	if err != nil {
		s.logger.Error("更新帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	categoryName := categoryData.Name

	// Build response data | 构建响应数据
	result := &schema.UserPostUpdateResponse{
		ID:                   updatedPost.ID,
		CategoryID:           updatedPost.CategoryID,
		CategoryName:         categoryName,
		Title:                updatedPost.Title,
		Content:              updatedPost.Content,
		Username:             userData.Username,
		ReadPermissionType:   string(updatedPost.ReadPermission),
		ReadPermissionPoints: updatedPost.ReadPermissionPoints,
		ViewCount:            updatedPost.ViewCount,
		LikeCount:            updatedPost.LikeCount,
		DislikeCount:         updatedPost.DislikeCount,
		FavoriteCount:        updatedPost.FavoriteCount,
		IsEssence:            updatedPost.IsEssence,
		IsPinned:             updatedPost.IsPinned,
		Status:               string(updatedPost.Status),
		CreatedAt:            updatedPost.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:            updatedPost.UpdatedAt.Format(time_tools.DateTimeFormat),
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
	s.logger.Info("获取帖子列表", zap.Int("category_id", req.CategoryID), zap.String("slug", req.Slug), zap.String("keyword", req.Keyword), zap.Int("page", req.Page), zap.Int("page_size", req.PageSize), tracing.WithTraceIDField(ctx))

	s.validateAndSetDefaults(&req) // Validate and set default values | 验证并设置默认值

	currentUserID := tracing.GetUserID(ctx) // Get current user ID | 获取当前用户ID
	isLoggedIn := currentUserID > 0         // Check if user is logged in | 检查用户是否登录

	excludeLoginRequiredCatIDs := s.getLoginRequiredCategoryIDs(ctx, isLoggedIn) // Get login required category IDs | 获取需要登录的版块ID列表

	categoryID, err := s.resolveCategoryID(ctx, req.Slug, req.CategoryID, isLoggedIn) // Resolve category ID | 解析版块ID
	if err != nil {
		return s.buildEmptyPostListResponse(req), nil // Return empty response on error | 错误时返回空响应
	}

	pinnedPosts := s.queryPinnedPosts(ctx, categoryID, req.Page, req.Keyword, excludeLoginRequiredCatIDs) // Query pinned posts | 查询置顶帖子

	posts, total, err := s.queryNormalPosts(ctx, categoryID, req, excludeLoginRequiredCatIDs) // Query normal posts | 查询普通帖子
	if err != nil {
		s.logger.Error("获取帖子列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	allPosts := append([]*ent.Post{}, pinnedPosts...) // Merge pinned and normal posts | 合并置顶帖和普通帖
	allPosts = append(allPosts, posts...)

	userMap := s.batchGetUserInfo(ctx, allPosts)         // Batch get user info | 批量获取用户信息
	categoryMap := s.batchGetCategoryInfo(ctx, allPosts) // Batch get category info | 批量获取版块信息

	postIDs := s.extractPostIDs(allPosts)                              // Extract post IDs | 提取帖子ID列表
	userLikeStatus := s.getUserLikeStatus(ctx, currentUserID, postIDs) // Get user like status | 获取用户点赞状态
	statsMap := s.batchGetStats(ctx, postIDs)                          // Get post stats | 获取帖子统计数据

	pinnedResult := s.convertPostsToResponse(pinnedPosts, userMap, categoryMap, userLikeStatus, statsMap) // Convert pinned posts to response | 转换置顶帖为响应格式
	result := s.convertPostsToResponse(posts, userMap, categoryMap, userLikeStatus, statsMap)             // Convert normal posts to response | 转换普通帖为响应格式

	totalPages := s.calculateTotalPages(total, req.PageSize) // Calculate total pages | 计算总页数

	return &schema.UserPostListResponse{
		PinnedPosts: pinnedResult,
		Posts:       result,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
	}, nil
}

// validateAndSetDefaults Validate and set default values for request | 验证并设置请求默认值
func (s *PostService) validateAndSetDefaults(req *schema.UserPostListRequest) {
	if req.Page <= 0 {
		req.Page = _const.DefaultPage // Set default page | 设置默认页码
	}
	if req.PageSize <= 0 {
		req.PageSize = _const.DefaultPageSize // Set default page size | 设置默认每页数量
	}
	if req.Sort == "" {
		req.Sort = _const.DefaultSort // Set default sort order | 设置默认排序方式
	}
}

// getLoginRequiredCategoryIDs Get login required category IDs | 获取登录可见版块ID列表
func (s *PostService) getLoginRequiredCategoryIDs(ctx context.Context, isLoggedIn bool) []int {
	if isLoggedIn {
		return nil // Return nil if user is logged in | 用户已登录返回nil
	}
	catIDs, err := s.categoryRepo.GetLoginRequiredCategoryIDs(ctx)
	if err != nil {
		s.logger.Warn("获取登录可见版块ID列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil
	}
	return catIDs
}

// resolveCategoryID Resolve category ID from slug or category ID | 从slug或category ID解析版块ID
func (s *PostService) resolveCategoryID(ctx context.Context, slug string, categoryID int, isLoggedIn bool) (int, error) {
	if slug != "" && categoryID == 0 {
		categoryData, err := s.categoryRepo.GetBySlug(ctx, slug) // Get category by slug | 通过slug获取版块
		if err != nil {
			s.logger.Warn("通过slug获取版块失败", zap.String("slug", slug), zap.Error(err), tracing.WithTraceIDField(ctx))
			return 0, err
		}
		categoryID = categoryData.ID

		if !isLoggedIn && categoryData.Status == category.StatusLoginRequired {
			s.logger.Warn("未登录用户尝试访问登录可见版块", zap.String("slug", slug), tracing.WithTraceIDField(ctx))
			return 0, errors.New("login required") // Return error if not logged in | 未登录返回错误
		}
	}

	if categoryID > 0 && !isLoggedIn {
		categoryData, err := s.categoryRepo.GetByID(ctx, categoryID) // Get category by ID | 通过ID获取版块
		if err == nil && categoryData.Status == category.StatusLoginRequired {
			s.logger.Warn("未登录用户尝试访问登录可见版块", zap.Int("category_id", categoryID), tracing.WithTraceIDField(ctx))
			return 0, errors.New("login required") // Return error if not logged in | 未登录返回错误
		}
	}

	return categoryID, nil
}

// queryPinnedPosts Query pinned posts | 查询置顶帖子
func (s *PostService) queryPinnedPosts(ctx context.Context, categoryID int, page int, keyword string, excludeLoginRequiredCatIDs []int) []*ent.Post {
	if page > 1 || keyword != "" {
		return nil // Return nil if not first page or has keyword | 非第一页或有搜索关键词时不查询置顶帖
	}

	var pinScopes []post.PinScope
	if categoryID == 0 {
		pinScopes = []post.PinScope{post.PinScopeHome, post.PinScopeGlobal} // Query home and global pinned | 查询首页和全局置顶
	} else {
		pinScopes = []post.PinScope{post.PinScopeCategory, post.PinScopeGlobal} // Query category and global pinned | 查询版块和全局置顶
	}

	pinnedPosts, _, err := s.postRepo.List(ctx, repository.ListPostOptions{
		CategoryID:                 categoryID,
		Keyword:                    keyword,
		Statuses:                   []post.Status{post.StatusNormal, post.StatusLocked},
		PinScopes:                  pinScopes,
		SortBy:                     "latest",
		ExcludeLoginRequiredCatIDs: excludeLoginRequiredCatIDs,
	})
	if err != nil {
		s.logger.Error("获取置顶帖子列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil
	}

	return pinnedPosts
}

// queryNormalPosts Query normal posts | 查询普通帖子
func (s *PostService) queryNormalPosts(ctx context.Context, categoryID int, req schema.UserPostListRequest, excludeLoginRequiredCatIDs []int) ([]*ent.Post, int, error) {
	return s.postRepo.List(ctx, repository.ListPostOptions{
		CategoryID:                 categoryID,
		Keyword:                    req.Keyword,
		Statuses:                   []post.Status{post.StatusNormal, post.StatusLocked},
		SortBy:                     req.Sort,
		Page:                       req.Page,
		PageSize:                   req.PageSize,
		ExcludePinned:              true, // Exclude pinned posts | 排除置顶帖子
		ExcludeLoginRequiredCatIDs: excludeLoginRequiredCatIDs,
	})
}

// batchGetUserInfo Batch get user information | 批量获取用户信息
func (s *PostService) batchGetUserInfo(ctx context.Context, posts []*ent.Post) map[int]struct {
	ID       int
	Username string
	Avatar   string
} {
	userIDs := make(map[int]bool)
	for _, p := range posts {
		userIDs[p.UserID] = true // Collect user IDs | 收集用户ID
	}

	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id) // Convert to slice | 转换为切片
	}

	users, err := s.userRepo.GetByIDsWithFields(ctx, userIDList, []string{user.FieldID, user.FieldUsername, user.FieldAvatar}) // Batch query users | 批量查询用户
	if err != nil {
		s.logger.Warn("批量查询用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return make(map[int]struct {
			ID       int
			Username string
			Avatar   string
		})
	}

	userMap := make(map[int]struct {
		ID       int
		Username string
		Avatar   string
	})
	for _, u := range users {
		userMap[u.ID] = struct {
			ID       int
			Username string
			Avatar   string
		}{
			ID:       u.ID,
			Username: u.Username,
			Avatar:   u.Avatar,
		} // Build user map | 构建用户映射
	}

	return userMap
}

// batchGetCategoryInfo Batch get category information | 批量获取版块信息
func (s *PostService) batchGetCategoryInfo(ctx context.Context, posts []*ent.Post) map[int]string {
	categoryIDs := make(map[int]bool)
	for _, p := range posts {
		categoryIDs[p.CategoryID] = true // Collect category IDs | 收集版块ID
	}

	categoryIDList := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		categoryIDList = append(categoryIDList, id) // Convert to slice | 转换为切片
	}

	categories, err := s.categoryRepo.GetByIDsWithFields(ctx, categoryIDList, []string{category.FieldID, category.FieldName}) // Batch query categories | 批量查询版块
	if err != nil {
		s.logger.Warn("批量查询版块信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return make(map[int]string)
	}

	categoryMap := make(map[int]string)
	for _, c := range categories {
		categoryMap[c.ID] = c.Name // Build category map | 构建版块映射
	}

	return categoryMap
}

// extractPostIDs Extract post IDs from posts | 从帖子中提取ID列表
func (s *PostService) extractPostIDs(posts []*ent.Post) []int {
	postIDs := make([]int, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID // Extract post ID | 提取帖子ID
	}
	return postIDs
}

// getUserLikeStatus Get user like status for posts | 获取用户对帖子的点赞状态
func (s *PostService) getUserLikeStatus(ctx context.Context, currentUserID int, postIDs []int) map[int]map[string]bool {
	if currentUserID == 0 {
		return make(map[int]map[string]bool) // Return empty map if not logged in | 未登录返回空映射
	}

	actions, err := s.postActionRepo.GetUserActionsForPosts(ctx, currentUserID, postIDs) // Get user actions | 获取用户操作记录
	if err != nil {
		s.logger.Warn("查询用户点赞状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return make(map[int]map[string]bool)
	}

	userLikeStatus := make(map[int]map[string]bool)
	for _, action := range actions {
		if _, exists := userLikeStatus[action.PostID]; !exists {
			userLikeStatus[action.PostID] = map[string]bool{"like": false, "dislike": false} // Initialize status | 初始化状态
		}
		switch action.ActionType {
		case postaction.ActionTypeLike:
			userLikeStatus[action.PostID]["like"] = true
		case postaction.ActionTypeDislike:
			userLikeStatus[action.PostID]["dislike"] = true
		}
	}

	return userLikeStatus
}

// batchGetStats Batch get stats data | 批量获取统计数据
func (s *PostService) batchGetStats(ctx context.Context, postIDs []int) map[int]*stats.Stats {
	statsMap, err := s.postStatsService.GetStatsMap(ctx, postIDs) // Get real-time stats | 获取实时统计数据
	if err != nil {
		s.logger.Warn("获取实时统计数据失败，将使用数据库中的旧数据", zap.Error(err), tracing.WithTraceIDField(ctx))
		return make(map[int]*stats.Stats)
	}
	return statsMap
}

// convertPostsToResponse Convert posts to response format | 将帖子转换为响应格式
func (s *PostService) convertPostsToResponse(posts []*ent.Post, userMap map[int]struct {
	ID       int
	Username string
	Avatar   string
}, categoryMap map[int]string, userLikeStatus map[int]map[string]bool, statsMap map[int]*stats.Stats) []schema.UserPostCreateResponse {
	result := make([]schema.UserPostCreateResponse, len(posts))
	for i, p := range posts {
		result[i] = s.convertPostToResponse(p, userMap, categoryMap, userLikeStatus, statsMap) // Convert each post | 转换每个帖子
	}
	return result
}

// convertPostToResponse Convert single post to response format | 将单个帖子转换为响应格式
func (s *PostService) convertPostToResponse(p *ent.Post, userMap map[int]struct {
	ID       int
	Username string
	Avatar   string
}, categoryMap map[int]string, userLikeStatus map[int]map[string]bool, statsMap map[int]*stats.Stats) schema.UserPostCreateResponse {
	userInfo := userMap[p.UserID]             // Get user info | 获取用户信息
	categoryName := categoryMap[p.CategoryID] // Get category name | 获取版块名称

	likeCount := p.LikeCount
	dislikeCount := p.DislikeCount
	favoriteCount := p.FavoriteCount
	viewCount := p.ViewCount
	if statsData, ok := statsMap[p.ID]; ok { // Prefer real-time stats | 优先使用实时统计数据
		likeCount = statsData.LikeCount
		dislikeCount = statsData.DislikeCount
		favoriteCount = statsData.FavoriteCount
		viewCount = statsData.ViewCount
	}

	userLiked := false
	userDisliked := false
	if status, exists := userLikeStatus[p.ID]; exists { // Get user like status | 获取用户点赞状态
		userLiked = status["like"]
		userDisliked = status["dislike"]
	}

	return schema.UserPostCreateResponse{
		ID:                   p.ID,
		CategoryID:           p.CategoryID,
		CategoryName:         categoryName,
		Title:                p.Title,
		Content:              "[内容已隐藏]", // Hide content in list | 列表中隐藏内容
		UserID:               userInfo.ID,
		Username:             userInfo.Username,
		Avatar:               userInfo.Avatar,
		ReadPermissionType:   string(p.ReadPermission),
		ReadPermissionPoints: p.ReadPermissionPoints,
		ViewCount:            viewCount,
		LikeCount:            likeCount,
		DislikeCount:         dislikeCount,
		FavoriteCount:        favoriteCount,
		UserLiked:            userLiked,
		UserDisliked:         userDisliked,
		IsEssence:            p.IsEssence,
		IsPinned:             p.IsPinned,
		Status:               string(p.Status),
		CreatedAt:            p.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:            p.UpdatedAt.Format(time_tools.DateTimeFormat),
	}
}

// calculateTotalPages Calculate total pages | 计算总页数
func (s *PostService) calculateTotalPages(total int, pageSize int) int {
	if pageSize <= 0 {
		return 0 // Return 0 if page size is invalid | 页面大小无效时返回0
	}
	return (total + pageSize - 1) / pageSize // Calculate total pages | 计算总页数
}

// buildEmptyPostListResponse Build empty post list response | 构建空的帖子列表响应
func (s *PostService) buildEmptyPostListResponse(req schema.UserPostListRequest) *schema.UserPostListResponse {
	return &schema.UserPostListResponse{
		PinnedPosts: []schema.UserPostCreateResponse{},
		Posts:       []schema.UserPostCreateResponse{},
		Total:       0,
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  0,
	}
}

// GetPostDetail Get post detail | 获取帖子详情
func (s *PostService) GetPostDetail(ctx context.Context, req schema.UserPostDetailRequest) (*schema.UserPostDetailResponse, error) {
	s.logger.Info("获取帖子详情", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	postData, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("获取帖子详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Get current user ID, 0 if not logged in | 获取当前用户ID，如果未登录则为0
	currentUserID := tracing.GetUserID(ctx)

	// Check if current user is the author | 检查当前用户是否为作者
	isAuthor := currentUserID > 0 && postData.UserID == currentUserID

	// Check post status (only allow Normal and Locked for non-authors) | 检查帖子状态（非作者只允许Normal和Locked）
	if !isAuthor && postData.Status != post.StatusNormal && postData.Status != post.StatusLocked {
		s.logger.Warn("帖子状态不允许访问", zap.Int("post_id", req.ID), zap.String("status", string(postData.Status)), tracing.WithTraceIDField(ctx))
		return nil, errors.New("帖子不存在或已删除")
	}

	// Check read permission (skip for author viewing own post) | 检查阅读权限（作者查看自己的帖子时跳过）
	if !isAuthor {
		hasPermission, reason := s.checkReadPermission(ctx, postData.ReadPermission, postData.ReadPermissionPoints, currentUserID)
		if !hasPermission {
			s.logger.Warn("用户无阅读权限", zap.Int("post_id", req.ID), zap.Int("user_id", currentUserID), zap.String("reason", reason), tracing.WithTraceIDField(ctx))
			return nil, &NoPermissionError{Reason: reason}
		}
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
		ID:                   postData.ID,
		CategoryID:           postData.CategoryID,
		CategoryName:         categoryName,
		Title:                postData.Title,
		Content:              postData.Content,
		UserID:               postData.UserID,
		Username:             username,
		Avatar:               author.Avatar,
		ReadPermissionType:   string(postData.ReadPermission),
		ReadPermissionPoints: postData.ReadPermissionPoints,
		ViewCount:            viewCount,
		LikeCount:            likeCount,
		DislikeCount:         dislikeCount,
		FavoriteCount:        favoriteCount,
		UserLiked:            userLiked,
		UserDisliked:         userDisliked,
		UserFavorited:        userFavorite,
		IsEssence:            postData.IsEssence,
		IsPinned:             postData.IsPinned,
		Status:               string(postData.Status),
		CreatedAt:            postData.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:            postData.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("获取帖子详情成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetDraftList Get draft list | 获取草稿列表
func (s *PostService) GetDraftList(ctx context.Context, userID int, req schema.UserDraftListRequest) (*schema.UserPostListResponse, error) {
	s.logger.Info("获取草稿列表", zap.Int("user_id", userID), zap.Int("page", req.Page), zap.Int("page_size", req.PageSize), tracing.WithTraceIDField(ctx))

	// Set default values | 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// Use repository to query user's draft posts | 使用 Repository 查询用户的草稿帖子
	posts, total, err := s.postRepo.GetByUserID(ctx, userID, repository.ListPostOptions{
		Status:   post.StatusDraft,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		s.logger.Error("获取草稿列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Collect category IDs | 收集版块ID
	categoryIDs := make(map[int]bool)
	for _, p := range posts {
		categoryIDs[p.CategoryID] = true
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

	// Get user info | 获取用户信息
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// Convert to response format | 转换为响应格式
	result := make([]schema.UserPostCreateResponse, len(posts))
	for i, p := range posts {
		categoryName := categoryMap[p.CategoryID]

		result[i] = schema.UserPostCreateResponse{
			ID:                   p.ID,
			CategoryID:           p.CategoryID,
			CategoryName:         categoryName,
			Title:                p.Title,
			Content:              p.Content,
			Username:             userData.Username,
			ReadPermissionType:   string(p.ReadPermission),
			ReadPermissionPoints: p.ReadPermissionPoints,
			ViewCount:            p.ViewCount,
			LikeCount:            p.LikeCount,
			DislikeCount:         p.DislikeCount,
			FavoriteCount:        p.FavoriteCount,
			IsEssence:            p.IsEssence,
			IsPinned:             p.IsPinned,
			Status:               string(p.Status),
			CreatedAt:            p.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:            p.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	totalPages := (total + req.PageSize - 1) / req.PageSize

	s.logger.Info("草稿列表获取成功", zap.Int("total", total), tracing.WithTraceIDField(ctx))
	return &schema.UserPostListResponse{
		Posts:      result,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// DeleteDraft Delete draft | 删除草稿
func (s *PostService) DeleteDraft(ctx context.Context, userID int, req schema.UserDraftDeleteRequest) error {
	s.logger.Info("删除草稿", zap.Int("user_id", userID), zap.Int("draft_id", req.ID), tracing.WithTraceIDField(ctx))

	// Get draft post | 获取草稿帖子
	draftPost, err := s.postRepo.GetByIDWithStatus(ctx, req.ID, post.StatusDraft)
	if err != nil {
		s.logger.Error("获取草稿失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	// Check ownership | 检查所有权
	if draftPost.UserID != userID {
		return errors.New("您不是该草稿的作者")
	}

	// Delete draft | 删除草稿
	err = s.db.Post.DeleteOneID(req.ID).Exec(ctx)
	if err != nil {
		s.logger.Error("删除草稿失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return err
	}

	s.logger.Info("草稿删除成功", zap.Int("draft_id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
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
	case user.StatusMute:
		return errors.New("您已被禁言，无法进行此操作")
	case user.StatusBlocked:
		return errors.New("您的账号已被封禁，无法进行此操作")
	default:
		return errors.New("账号状态异常，无法进行此操作")
	}
}

// checkReadPermission Check if user has permission to read the post | 检查用户是否有阅读权限
// Returns (hasPermission, reason) | 返回 (是否有权限, 原因)
func (s *PostService) checkReadPermission(ctx context.Context, readPermission post.ReadPermission, readPermissionPoints int, currentUserID int) (bool, string) {
	// "public" means public access | "public"表示公开访问
	if readPermission == post.ReadPermissionPublic || readPermission == "" {
		return true, ""
	}

	// "login_required" means login is required | "login_required"表示需要登录
	if readPermission == post.ReadPermissionLoginRequired {
		if currentUserID <= 0 {
			return false, "该帖子需要登录后查看"
		}
		return true, ""
	}

	// "points_required" means points greater than threshold is required | "points_required"表示需要积分大于阈值
	if readPermission == post.ReadPermissionPointsRequired {
		if currentUserID <= 0 {
			return false, "该帖子需要登录并满足积分要求后查看"
		}

		// Get user points | 获取用户积分
		userData, err := s.userRepo.GetByID(ctx, currentUserID)
		if err != nil {
			s.logger.Error("获取用户信息失败，拒绝访问", zap.Error(err), tracing.WithTraceIDField(ctx))
			return false, "获取用户信息失败，无法验证阅读权限"
		}

		if userData.Points < readPermissionPoints {
			return false, fmt.Sprintf("该帖子需要积分达到 %d 才能查看，您当前积分为 %d", readPermissionPoints, userData.Points)
		}
		return true, ""
	}

	// Unknown permission type, deny access and log error | 未知权限类型，拒绝访问并记录错误
	s.logger.Error("未知的阅读权限类型，拒绝访问", zap.String("read_permission", string(readPermission)), tracing.WithTraceIDField(ctx))
	return false, "阅读权限配置异常，暂时无法访问"
}

// parseReadPermissionType Parse read permission type from string to enum | 从字符串解析阅读权限类型为枚举
func (s *PostService) parseReadPermissionType(permissionType string) post.ReadPermission {
	switch permissionType {
	case "login_required":
		return post.ReadPermissionLoginRequired
	case "points":
		return post.ReadPermissionPointsRequired
	default:
		return post.ReadPermissionPublic
	}
}
