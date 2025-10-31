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
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewPostService 创建帖子服务实例
func NewPostService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IPostService {
	return &PostService{
		db:     db,
		cache:  cacheService,
		logger: logger,
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
		WithCategory().
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

	// 构建响应数据
	result := &schema.UserPostUpdateResponse{
		ID:             updatedPost.ID,
		CategoryID:     updatedPost.CategoryID,
		CategoryName:   postData.Edges.Category.Name,
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

	// 检查是否已经点赞
	existingAction, err := s.db.PostAction.Query().
		Where(
			postaction.UserID(userID),
			postaction.PostID(req.ID),
			postaction.ActionTypeEQ(postaction.ActionTypeLike),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Error("查询点赞记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询点赞记录失败: %w", err)
	}

	// 如果已经点赞，不可取消（单向操作）
	if existingAction != nil {
		return nil, errors.New("您已经点赞过该帖子，不可取消点赞")
	}

	// 创建点赞记录
	_, err = s.db.PostAction.Create().
		SetUserID(userID).
		SetPostID(req.ID).
		SetActionType(postaction.ActionTypeLike).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建点赞记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建点赞记录失败: %w", err)
	}

	// 更新帖子点赞数
	updatedPost, err := s.db.Post.UpdateOneID(req.ID).
		SetLikeCount(postData.LikeCount + 1).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新帖子点赞数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新帖子点赞数失败: %w", err)
	}

	// 构建响应数据
	result := &schema.UserPostActionResponse{
		Success:       true,
		LikeCount:     updatedPost.LikeCount,
		DislikeCount:  updatedPost.DislikeCount,
		FavoriteCount: updatedPost.FavoriteCount,
		ActionType:    "like",
	}

	s.logger.Info("帖子点赞成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// DislikePost 点踩帖子
func (s *PostService) DislikePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("点踩帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

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

	// 检查是否已经点踩
	existingAction, err := s.db.PostAction.Query().
		Where(
			postaction.UserID(userID),
			postaction.PostID(req.ID),
			postaction.ActionTypeEQ(postaction.ActionTypeDislike),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Error("查询点踩记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询点踩记录失败: %w", err)
	}

	// 如果已经点踩，不可取消（单向操作）
	if existingAction != nil {
		return nil, errors.New("您已经点踩过该帖子，不可取消点踩")
	}

	// 创建点踩记录
	_, err = s.db.PostAction.Create().
		SetUserID(userID).
		SetPostID(req.ID).
		SetActionType(postaction.ActionTypeDislike).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建点踩记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建点踩记录失败: %w", err)
	}

	// 更新帖子点踩数
	updatedPost, err := s.db.Post.UpdateOneID(req.ID).
		SetDislikeCount(postData.DislikeCount + 1).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新帖子点踩数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新帖子点踩数失败: %w", err)
	}

	// 构建响应数据
	result := &schema.UserPostActionResponse{
		Success:       true,
		LikeCount:     updatedPost.LikeCount,
		DislikeCount:  updatedPost.DislikeCount,
		FavoriteCount: updatedPost.FavoriteCount,
		ActionType:    "dislike",
	}

	s.logger.Info("帖子点踩成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// FavoritePost 收藏帖子
func (s *PostService) FavoritePost(ctx context.Context, userID int, req schema.UserPostActionRequest) (*schema.UserPostActionResponse, error) {
	s.logger.Info("收藏帖子", zap.Int("user_id", userID), zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))

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

	// 检查是否已经收藏
	existingAction, err := s.db.PostAction.Query().
		Where(
			postaction.UserID(userID),
			postaction.PostID(req.ID),
			postaction.ActionTypeEQ(postaction.ActionTypeFavorite),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Error("查询收藏记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询收藏记录失败: %w", err)
	}

	// 如果已经收藏，取消收藏（双向操作）
	if existingAction != nil {
		// 删除收藏记录
		err = s.db.PostAction.DeleteOne(existingAction).Exec(ctx)
		if err != nil {
			s.logger.Error("删除收藏记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("删除收藏记录失败: %w", err)
		}

		// 更新帖子收藏数
		updatedPost, err := s.db.Post.UpdateOneID(req.ID).
			SetFavoriteCount(postData.FavoriteCount - 1).
			Save(ctx)
		if err != nil {
			s.logger.Error("更新帖子收藏数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("更新帖子收藏数失败: %w", err)
		}

		// 构建响应数据
		result := &schema.UserPostActionResponse{
			Success:       true,
			LikeCount:     updatedPost.LikeCount,
			DislikeCount:  updatedPost.DislikeCount,
			FavoriteCount: updatedPost.FavoriteCount,
			ActionType:    "unfavorite",
		}

		s.logger.Info("取消收藏帖子成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
		return result, nil
	}

	// 创建收藏记录
	_, err = s.db.PostAction.Create().
		SetUserID(userID).
		SetPostID(req.ID).
		SetActionType(postaction.ActionTypeFavorite).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建收藏记录失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建收藏记录失败: %w", err)
	}

	// 更新帖子收藏数
	updatedPost, err := s.db.Post.UpdateOneID(req.ID).
		SetFavoriteCount(postData.FavoriteCount + 1).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新帖子收藏数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新帖子收藏数失败: %w", err)
	}

	// 构建响应数据
	result := &schema.UserPostActionResponse{
		Success:       true,
		LikeCount:     updatedPost.LikeCount,
		DislikeCount:  updatedPost.DislikeCount,
		FavoriteCount: updatedPost.FavoriteCount,
		ActionType:    "favorite",
	}

	s.logger.Info("收藏帖子成功", zap.Int("post_id", req.ID), tracing.WithTraceIDField(ctx))
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
	query := s.db.Post.Query().
		WithAuthor().
		WithCategory()

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
		s.logger.Error("获取帖子列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子列表失败: %w", err)
	}

	// 转换为响应格式
	result := make([]schema.UserPostCreateResponse, len(posts))
	for i, p := range posts {
		username := ""
		if p.Edges.Author != nil {
			username = p.Edges.Author.Username
		}
		categoryName := ""
		if p.Edges.Category != nil {
			categoryName = p.Edges.Category.Name
		}

		result[i] = schema.UserPostCreateResponse{
			ID:             p.ID,
			CategoryID:     p.CategoryID,
			CategoryName:   categoryName,
			Title:          p.Title,
			Content:        p.Content,
			Username:       username,
			ReadPermission: p.ReadPermission,
			ViewCount:      p.ViewCount,
			LikeCount:      p.LikeCount,
			DislikeCount:   p.DislikeCount,
			FavoriteCount:  p.FavoriteCount,
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
		WithAuthor().
		WithCategory().
		Where(post.IDEQ(req.ID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子详情失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子详情失败: %w", err)
	}

	// 更新浏览数
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetViewCount(postData.ViewCount + 1).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新浏览数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 不影响主要流程，继续返回数据
	}

	// 构建响应数据
	username := ""
	if postData.Edges.Author != nil {
		username = postData.Edges.Author.Username
	}
	categoryName := ""
	if postData.Edges.Category != nil {
		categoryName = postData.Edges.Category.Name
	}

	result := &schema.UserPostDetailResponse{
		ID:             postData.ID,
		CategoryID:     postData.CategoryID,
		CategoryName:   categoryName,
		Title:          postData.Title,
		Content:        postData.Content,
		Username:       username,
		ReadPermission: postData.ReadPermission,
		ViewCount:      postData.ViewCount + 1, // 返回更新后的浏览数
		LikeCount:      postData.LikeCount,
		DislikeCount:   postData.DislikeCount,
		FavoriteCount:  postData.FavoriteCount,
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
	lastEditTime, err := s.cache.Get(redisKey)
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
	err = s.cache.SetEx(redisKey, currentTime, 180) // 180秒 = 3分钟
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
	lastPrivateTime, err := s.cache.Get(redisKey)
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
	err = s.cache.SetEx(redisKey, currentTime, 259200) // 259200秒 = 3天
	if err != nil {
		s.logger.Error("设置私有限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 设置失败，但允许操作
		return true, nil
	}

	return true, nil
}
