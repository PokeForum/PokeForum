package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/postaction"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/stats"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// IPostService 帖子服务接口
type IPostService interface {
	// CreatePost 创建帖子
	CreatePost(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error)
	// SaveDraft 保存草稿
	SaveDraft(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error)
	// UpdatePost 更新帖子
	UpdatePost(ctx context.Context, userID int, req schema.UserPostUpdateRequest) (*schema.UserPostUpdateResponse, error)
	// SetPostPrivate 设置帖子私有
	SetPostPrivate(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error)
	// LikePost 点赞帖子
	LikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error)
	// DislikePost 点踩帖子
	DislikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error)
	// FavoritePost 收藏帖子
	FavoritePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error)
	// GetPostList 获取帖子列表
	GetPostList(ctx context.Context, req schema.UserPostListRequest) (*schema.UserPostListResponse, error)
	// GetPostDetail 获取帖子详情
	GetPostDetail(ctx context.Context, req schema.UserPostDetailRequest) (*schema.UserPostDetailResponse, error)
	// CheckEditPermission 检查编辑权限（每三分钟可操作一次）
	CheckEditPermission(ctx context.Context, userID, postID int) (bool, error)
	// CheckPrivatePermission 检查私有权限（每三日可操作一次）
	CheckPrivatePermission(ctx context.Context, userID, postID int) (bool, error)
}

// PostService 帖子服务实现
type PostService struct {
	db               *ent.Client
	cache            cache.ICacheService
	logger           *zap.Logger
	postStatsService IPostStatsService
}

// NewPostService 创建帖子服务实例
func NewPostService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IPostService {
	return &PostService{
		db:               db,
		cache:            cacheService,
		logger:           logger,
		postStatsService: NewPostStatsService(db, cacheService, logger),
	}
}

// CreatePost 创建帖子
func (s *PostService) CreatePost(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error) {
	s.logger.Info("创建帖子", zap.Int("user_id", userID), zap.Int("category_id", req.CategoryID), zap.String("title", req.Title), tracing.WithTraceIDField(ctx))

	// 检查版块是否存在
	categoryData, err := s.db.Category.Query().
		Where(category.IDEQ(req.CategoryID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("版块不存在")
		}
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块失败: %w", err)
	}

	// 获取用户信息
	userData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 创建帖子
	newPost, err := s.db.Post.Create().
		SetUserID(userID).
		SetCategoryID(req.CategoryID).
		SetTitle(req.Title).
		SetContent(req.Content).
		SetReadPermission(req.ReadPermission).
		SetStatus(post.StatusNormal).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建帖子失败: %w", err)
	}

	// 构建响应数据
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
		CreatedAt:      newPost.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      newPost.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("帖子创建成功", zap.Int("post_id", newPost.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// SaveDraft 保存草稿
func (s *PostService) SaveDraft(ctx context.Context, userID int, req schema.UserPostCreateRequest) (*schema.UserPostCreateResponse, error) {
	s.logger.Info("保存草稿", zap.Int("user_id", userID), zap.Int("category_id", req.CategoryID), zap.String("title", req.Title), tracing.WithTraceIDField(ctx))

	// 检查版块是否存在
	categoryData, err := s.db.Category.Query().
		Where(category.IDEQ(req.CategoryID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("版块不存在")
		}
		s.logger.Error("获取版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取版块失败: %w", err)
	}

	// 获取用户信息
	userData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 创建草稿
	newPost, err := s.db.Post.Create().
		SetUserID(userID).
		SetCategoryID(req.CategoryID).
		SetTitle(req.Title).
		SetContent(req.Content).
		SetReadPermission(req.ReadPermission).
		SetStatus(post.StatusDraft).
		Save(ctx)
	if err != nil {
		s.logger.Error("保存草稿失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("保存草稿失败: %w", err)
	}

	// 构建响应数据
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
		CreatedAt:      newPost.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      newPost.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("草稿保存成功", zap.Int("post_id", newPost.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// UpdatePost 更新帖子
func (s *PostService) UpdatePost(ctx context.Context, userID int, req schema.UserPostUpdateRequest) (*schema.UserPostUpdateResponse, error) {
	s.logger.Info("更新帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查是否为帖子作者
	if postData.UserID != userID {
		return nil, errors.New("您不是该帖子的作者")
	}

	// 检查编辑权限（每三分钟可操作一次）
	canEdit, err := s.CheckEditPermission(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, errors.New("帖子编辑过于频繁，请三分钟后再试")
	}

	// 获取用户信息
	userData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 更新帖子
	updatedPost, err := s.db.Post.UpdateOneID(req.ID).
		SetTitle(req.Title).
		SetContent(req.Content).
		SetReadPermission(req.ReadPermission).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新帖子失败: %w", err)
	}

	// 查询版块信息
	categoryName := ""
	categoryData, err := s.db.Category.Query().
		Where(category.IDEQ(updatedPost.CategoryID)).
		Select(category.FieldName).
		Only(ctx)
	if err == nil {
		categoryName = categoryData.Name
	}

	// 构建响应数据
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
		CreatedAt:      updatedPost.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      updatedPost.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("帖子更新成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// SetPostPrivate 设置帖子私有
func (s *PostService) SetPostPrivate(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("设置帖子私有", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查是否为帖子作者
	if postData.UserID != userID {
		return nil, errors.New("您不是该帖子的作者")
	}

	// 检查私有权限（每三日可操作一次）
	canSetPrivate, err := s.CheckPrivatePermission(ctx, userID, req.ID)
	if err != nil {
		return nil, err
	}
	if !canSetPrivate {
		return nil, errors.New("帖子私有设置过于频繁，请三天后再试")
	}

	// 切换私有状态
	targetStatus := post.StatusPrivate
	if postData.Status == post.StatusPrivate {
		targetStatus = post.StatusNormal
	}

	// 更新帖子状态
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetStatus(targetStatus).
		Save(ctx)
	if err != nil {
		s.logger.Error("设置帖子私有失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("设置帖子私有失败: %w", err)
	}

	// 构建响应数据
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

// LikePost 点赞帖子
func (s *PostService) LikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("点赞帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// 使用统计服务执行点赞操作
	action, err := s.postStatsService.PerformAction(ctx, userID, req.ID, "Like")
	if err != nil {
		s.logger.Error("点赞帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 构建响应数据
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

// DislikePost 点踩帖子
func (s *PostService) DislikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("点踩帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// 使用统计服务执行点踩操作
	action, err := s.postStatsService.PerformAction(ctx, userID, req.ID, "Dislike")
	if err != nil {
		s.logger.Error("点踩帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 构建响应数据
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

// FavoritePost 收藏帖子
func (s *PostService) FavoritePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("收藏帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// 先检查用户是否已经收藏
	userActionStatus, err := s.postStatsService.GetUserActionStatus(ctx, userID, req.ID)
	if err != nil {
		s.logger.Error("获取用户操作状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	var postStats *stats.Stats
	var actionType string

	if userActionStatus.HasFavorited {
		// 已经收藏,执行取消收藏
		postStats, err = s.postStatsService.CancelAction(ctx, userID, req.ID, "Favorite")
		if err != nil {
			s.logger.Error("取消收藏失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, err
		}
		actionType = "unfavorite"
		s.logger.Info("取消收藏帖子成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	} else {
		// 未收藏,执行收藏
		postStats, err = s.postStatsService.PerformAction(ctx, userID, req.ID, "Favorite")
		if err != nil {
			s.logger.Error("收藏帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, err
		}
		actionType = "favorite"
		s.logger.Info("收藏帖子成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	}

	// 构建响应数据
	result := &schema.UserPostActionResponse{
		Success:       true,
		LikeCount:     postStats.LikeCount,
		DislikeCount:  postStats.DislikeCount,
		FavoriteCount: postStats.FavoriteCount,
		ActionType:    actionType,
	}

	return result, nil
}

// GetPostList 获取帖子列表
func (s *PostService) GetPostList(ctx context.Context, req schema.UserPostListRequest) (*schema.UserPostListResponse, error) {
	s.logger.Info("获取帖子列表", zap.Int("category_id", req.CategoryID), zap.Int("page", req.Page), zap.Int("page_size", req.PageSize), tracing.WithTraceIDField(ctx))

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Sort == "" {
		req.Sort = "latest"
	}

	// 构建查询条件
	query := s.db.Post.Query()

	// 如果指定了版块ID
	if req.CategoryID > 0 {
		query = query.Where(post.CategoryID(req.CategoryID))
	}

	// 只显示正常状态的帖子
	query = query.Where(post.StatusEQ(post.StatusNormal))

	// 根据排序方式设置排序
	switch req.Sort {
	case "hot":
		query = query.Order(ent.Desc(post.FieldViewCount), ent.Desc(post.FieldLikeCount))
	case "essence":
		query = query.Where(post.IsEssence(true)).Order(ent.Desc(post.FieldCreatedAt))
	default: // latest
		query = query.Order(ent.Desc(post.FieldCreatedAt))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取帖子总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子总数失败: %w", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	posts, err := query.
		Offset(offset).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取帖子列表失败", zap.Error(err))
		return nil, fmt.Errorf("获取帖子列表失败: %w", err)
	}

	// 收集用户ID和版块ID
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

	// 获取当前用户ID，如果未登录则为0
	currentUserID := tracing.GetUserID(ctx)

	// 获取帖子ID列表
	postIDs := make([]int, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	// 批量查询用户点赞状态（仅当用户已登录时）
	var userLikeStatus map[int]map[string]bool // postID -> {like: bool, dislike: bool}
	if currentUserID != 0 {
		// 批量查询用户对这些帖子的点赞记录
		actions, err := s.db.PostAction.Query().
			Where(
				postaction.UserIDEQ(currentUserID),
				postaction.PostIDIn(postIDs...),
			).
			Select(postaction.FieldPostID, postaction.FieldActionType).
			All(ctx)
		if err != nil {
			s.logger.Warn("查询用户点赞状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			// 失败时不阻断流程，使用默认状态
			userLikeStatus = make(map[int]map[string]bool)
		} else {
			// 构建点赞状态映射
			userLikeStatus = make(map[int]map[string]bool)
			for _, action := range actions {
				if _, exists := userLikeStatus[action.PostID]; !exists {
					userLikeStatus[action.PostID] = map[string]bool{"like": false, "dislike": false}
				}
				if action.ActionType == postaction.ActionTypeLike {
					userLikeStatus[action.PostID]["like"] = true
				} else if action.ActionType == postaction.ActionTypeDislike {
					userLikeStatus[action.PostID]["dislike"] = true
				}
			}
		}
	} else {
		// 未登录用户，所有状态为false
		userLikeStatus = make(map[int]map[string]bool)
	}

	// 批量获取实时统计数据
	statsMap, err := s.postStatsService.GetStatsMap(ctx, postIDs)
	if err != nil {
		s.logger.Warn("获取实时统计数据失败，将使用数据库中的旧数据", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 失败时不阻断流程，降级使用数据库数据
		statsMap = make(map[int]*stats.Stats)
	}

	// 转换为响应格式
	result := make([]schema.UserPostCreateResponse, len(posts))
	for i, p := range posts {
		username := userMap[p.UserID]
		categoryName := categoryMap[p.CategoryID]

		// 优先使用实时统计数据
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

		// 获取用户点赞状态
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
			CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

// GetPostDetail 获取帖子详情
func (s *PostService) GetPostDetail(ctx context.Context, req schema.UserPostDetailRequest) (*schema.UserPostDetailResponse, error) {
	s.logger.Info("获取帖子详情", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

	// 查询帖子详情
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子详情失败: %w", err)
	}

	// 更新浏览数(使用统计服务,减少数据库压力)
	if err = s.postStatsService.IncrViewCount(ctx, req.ID); err != nil {
		s.logger.Warn("增加帖子浏览数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 不影响主要流程
	}

	// 获取实时统计数据
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

	// 查询作者信息
	username := ""
	author, err := s.db.User.Query().
		Where(user.IDEQ(postData.UserID)).
		Select(user.FieldUsername).
		Only(ctx)
	if err == nil {
		username = author.Username
	}

	// 查询版块信息
	categoryName := ""
	categoryData, err := s.db.Category.Query().
		Where(category.IDEQ(postData.CategoryID)).
		Select(category.FieldName).
		Only(ctx)
	if err == nil {
		categoryName = categoryData.Name
	}

	// 获取当前用户ID，如果未登录则为0
	currentUserID := tracing.GetUserID(ctx)

	// 查询用户点赞状态（仅当用户已登录时）
	userLiked := false
	userDisliked := false
	if currentUserID != 0 {
		action, err := s.db.PostAction.Query().
			Where(
				postaction.UserIDEQ(currentUserID),
				postaction.PostIDEQ(req.ID),
			).
			Select(postaction.FieldActionType).
			Only(ctx)
		if err != nil {
			// 没有记录或查询失败，保持默认状态
			s.logger.Debug("查询用户点赞状态失败或无记录", zap.Error(err), tracing.WithTraceIDField(ctx))
		} else {
			if action.ActionType == postaction.ActionTypeLike {
				userLiked = true
			} else if action.ActionType == postaction.ActionTypeDislike {
				userDisliked = true
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
		IsEssence:      postData.IsEssence,
		IsPinned:       postData.IsPinned,
		Status:         string(postData.Status),
		CreatedAt:      postData.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      postData.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("获取帖子详情成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// CheckEditPermission 检查编辑权限（每三分钟可操作一次）
func (s *PostService) CheckEditPermission(ctx context.Context, userID, postID int) (bool, error) {
	// 生成Redis键名：post:edit:limit:{userID}:{postID}
	redisKey := fmt.Sprintf("post:edit:limit:%d:%d", userID, postID)

	// 检查是否在限制期内
	lastEditTime, err := s.cache.Get(ctx, redisKey)
	if err != nil {
		s.logger.Error("获取编辑限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return false, fmt.Errorf("获取编辑限制失败: %w", err)
	}

	if lastEditTime != "" {
		// 解析最后编辑时间
		lastTime, err := time.Parse(time.RFC3339, lastEditTime)
		if err != nil {
			s.logger.Error("解析最后编辑时间失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			// 解析失败，允许操作
			return true, nil
		}

		// 检查是否在三分钟内
		if time.Since(lastTime) < 3*time.Minute {
			s.logger.Warn("编辑操作过于频繁", zap.Int("user_id", userID), zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
			return false, nil
		}
	}

	// 设置新的编辑时间限制
	currentTime := time.Now().Format(time.RFC3339)
	err = s.cache.SetEx(ctx, redisKey, currentTime, 180) // 180秒 = 3分钟
	if err != nil {
		s.logger.Error("设置编辑限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 设置失败，但允许操作
		return true, nil
	}

	return true, nil
}

// CheckPrivatePermission 检查私有权限（每三日可操作一次）
func (s *PostService) CheckPrivatePermission(ctx context.Context, userID, postID int) (bool, error) {
	// 生成Redis键名：post:private:limit:{userID}:{postID}
	redisKey := fmt.Sprintf("post:private:limit:%d:%d", userID, postID)

	// 检查是否在限制期内
	lastPrivateTime, err := s.cache.Get(ctx, redisKey)
	if err != nil {
		s.logger.Error("获取私有设置限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return false, fmt.Errorf("获取私有设置限制失败: %w", err)
	}

	if lastPrivateTime != "" {
		// 解析最后私有设置时间
		lastTime, err := time.Parse(time.RFC3339, lastPrivateTime)
		if err != nil {
			s.logger.Error("解析最后私有设置时间失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			// 解析失败，允许操作
			return true, nil
		}

		// 检查是否在三日内
		if time.Since(lastTime) < 3*24*time.Hour {
			s.logger.Warn("私有设置操作过于频繁", zap.Int("user_id", userID), zap.Int("post_id", postID), tracing.WithTraceIDField(ctx))
			return false, nil
		}
	}

	// 设置新的私有设置时间限制
	currentTime := time.Now().Format(time.RFC3339)
	err = s.cache.SetEx(ctx, redisKey, currentTime, 259200) // 259200秒 = 3天
	if err != nil {
		s.logger.Error("设置私有限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 设置失败，但允许操作
		return true, nil
	}

	return true, nil
}
