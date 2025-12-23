package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IPostManageService 帖子管理服务接口
type IPostManageService interface {
	// GetPostList 获取帖子列表
	GetPostList(ctx context.Context, req schema.PostListRequest) (*schema.PostListResponse, error)
	// CreatePost 创建帖子
	CreatePost(ctx context.Context, req schema.PostCreateRequest) (*ent.Post, error)
	// UpdatePost 更新帖子信息
	UpdatePost(ctx context.Context, req schema.PostUpdateRequest) (*ent.Post, error)
	// UpdatePostStatus 更新帖子状态
	UpdatePostStatus(ctx context.Context, req schema.PostStatusUpdateRequest) error
	// GetPostDetail 获取帖子详情
	GetPostDetail(ctx context.Context, id int) (*schema.PostDetailResponse, error)
	// SetPostEssence 设置帖子精华
	SetPostEssence(ctx context.Context, req schema.PostEssenceUpdateRequest) error
	// SetPostPin 设置帖子置顶
	SetPostPin(ctx context.Context, req schema.PostPinUpdateRequest) error
	// MovePost 移动帖子到其他版块
	MovePost(ctx context.Context, req schema.PostMoveRequest) error
	// DeletePost 删除帖子（软删除，状态设为Ban）
	DeletePost(ctx context.Context, id int) error
}

// PostManageService 帖子管理服务实现
type PostManageService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewPostManageService 创建帖子管理服务实例
func NewPostManageService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IPostManageService {
	return &PostManageService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// GetPostList 获取帖子列表
func (s *PostManageService) GetPostList(ctx context.Context, req schema.PostListRequest) (*schema.PostListResponse, error) {
	s.logger.Info("获取帖子列表", tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.Post.Query()

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where(
			post.Or(
				post.TitleContains(req.Keyword),
				post.ContentContains(req.Keyword),
			),
		)
	}

	// 状态筛选
	if req.Status != "" {
		query = query.Where(post.StatusEQ(post.Status(req.Status)))
	}

	// 版块筛选
	if req.CategoryID > 0 {
		query = query.Where(post.CategoryIDEQ(req.CategoryID))
	}

	// 用户筛选
	if req.UserID > 0 {
		query = query.Where(post.UserIDEQ(req.UserID))
	}

	// 精华帖筛选
	if req.IsEssence != nil {
		query = query.Where(post.IsEssenceEQ(*req.IsEssence))
	}

	// 置顶筛选
	if req.IsPinned != nil {
		query = query.Where(post.IsPinnedEQ(*req.IsPinned))
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取帖子总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子总数失败: %w", err)
	}

	// 分页查询，置顶帖在前，然后按创建时间倒序
	posts, err := query.
		Order(ent.Desc(post.FieldIsPinned), ent.Desc(post.FieldCreatedAt)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取帖子列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
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

	// 转换为响应格式
	list := make([]schema.PostListItem, len(posts))
	for i, p := range posts {
		// 截取内容前100字符
		content := p.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		// 获取关联信息
		username := userMap[p.UserID]
		categoryName := categoryMap[p.CategoryID]

		list[i] = schema.PostListItem{
			ID:            p.ID,
			UserID:        p.UserID,
			Username:      username,
			CategoryID:    p.CategoryID,
			CategoryName:  categoryName,
			Title:         p.Title,
			Content:       content,
			ViewCount:     p.ViewCount,
			LikeCount:     p.LikeCount,
			DislikeCount:  p.DislikeCount,
			FavoriteCount: p.FavoriteCount,
			IsEssence:     p.IsEssence,
			IsPinned:      p.IsPinned,
			Status:        p.Status.String(),
			PublishIP:     p.PublishIP,
			CreatedAt:     p.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:     p.UpdatedAt.Format(time_tools.DateTimeFormat),
		}
	}

	return &schema.PostListResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CreatePost 创建帖子
func (s *PostManageService) CreatePost(ctx context.Context, req schema.PostCreateRequest) (*ent.Post, error) {
	s.logger.Info("创建帖子", zap.String("title", req.Title), zap.Int("user_id", req.UserID), tracing.WithTraceIDField(ctx))

	// 检查用户是否存在
	userExists, err := s.db.User.Query().
		Where(user.IDEQ(req.UserID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查用户失败: %w", err)
	}
	if !userExists {
		return nil, errors.New("用户不存在")
	}

	// 检查版块是否存在
	categoryExists, err := s.db.Category.Query().
		Where(category.IDEQ(req.CategoryID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查版块失败: %w", err)
	}
	if !categoryExists {
		return nil, errors.New("版块不存在")
	}

	// 创建帖子
	p, err := s.db.Post.Create().
		SetUserID(req.UserID).
		SetCategoryID(req.CategoryID).
		SetTitle(req.Title).
		SetContent(req.Content).
		SetReadPermission(req.ReadPermission).
		SetPublishIP(req.PublishIP).
		SetStatus(post.Status(req.Status)).
		Save(ctx)
	if err != nil {
		s.logger.Error("创建帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建帖子失败: %w", err)
	}

	s.logger.Info("帖子创建成功", zap.Int("id", p.ID), tracing.WithTraceIDField(ctx))
	return p, nil
}

// UpdatePost 更新帖子信息
func (s *PostManageService) UpdatePost(ctx context.Context, req schema.PostUpdateRequest) (*ent.Post, error) {
	s.logger.Info("更新帖子信息", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	_, err := s.db.Post.Get(ctx, req.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子失败: %w", err)
	}

	// 构建更新操作
	update := s.db.Post.UpdateOneID(req.ID)

	if req.Title != "" {
		update = update.SetTitle(req.Title)
	}
	if req.Content != "" {
		update = update.SetContent(req.Content)
	}
	if req.ReadPermission != "" {
		update = update.SetReadPermission(req.ReadPermission)
	}
	if req.Status != "" {
		update = update.SetStatus(post.Status(req.Status))
	}

	// 执行更新
	updatedPost, err := update.Save(ctx)
	if err != nil {
		s.logger.Error("更新帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新帖子失败: %w", err)
	}

	s.logger.Info("帖子更新成功", zap.Int("id", updatedPost.ID), tracing.WithTraceIDField(ctx))
	return updatedPost, nil
}

// UpdatePostStatus 更新帖子状态
func (s *PostManageService) UpdatePostStatus(ctx context.Context, req schema.PostStatusUpdateRequest) error {
	s.logger.Info("更新帖子状态", zap.Int("id", req.ID), zap.String("status", req.Status), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	exists, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查帖子失败: %w", err)
	}
	if !exists {
		return errors.New("帖子不存在")
	}

	// 更新状态
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetStatus(post.Status(req.Status)).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新帖子状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("更新帖子状态失败: %w", err)
	}

	s.logger.Info("帖子状态更新成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// GetPostDetail 获取帖子详情
func (s *PostManageService) GetPostDetail(ctx context.Context, id int) (*schema.PostDetailResponse, error) {
	s.logger.Info("获取帖子详情", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// 获取帖子信息
	p, err := s.db.Post.Query().
		Where(post.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子失败: %w", err)
	}

	// 查询作者信息
	username := ""
	author, err := s.db.User.Query().
		Where(user.IDEQ(p.UserID)).
		Select(user.FieldUsername).
		Only(ctx)
	if err == nil {
		username = author.Username
	}

	// 查询版块信息
	categoryName := ""
	categoryData, err := s.db.Category.Query().
		Where(category.IDEQ(p.CategoryID)).
		Select(category.FieldName).
		Only(ctx)
	if err == nil {
		categoryName = categoryData.Name
	}

	// 转换为响应格式
	result := &schema.PostDetailResponse{
		ID:             p.ID,
		UserID:         p.UserID,
		Username:       username,
		CategoryID:     p.CategoryID,
		CategoryName:   categoryName,
		Title:          p.Title,
		Content:        p.Content,
		ReadPermission: p.ReadPermission,
		ViewCount:      p.ViewCount,
		LikeCount:      p.LikeCount,
		DislikeCount:   p.DislikeCount,
		FavoriteCount:  p.FavoriteCount,
		IsEssence:      p.IsEssence,
		IsPinned:       p.IsPinned,
		Status:         p.Status.String(),
		PublishIP:      p.PublishIP,
		CreatedAt:      p.CreatedAt.Format(time_tools.DateTimeFormat),
		UpdatedAt:      p.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	return result, nil
}

// SetPostEssence 设置帖子精华
func (s *PostManageService) SetPostEssence(ctx context.Context, req schema.PostEssenceUpdateRequest) error {
	s.logger.Info("设置帖子精华", zap.Int("id", req.ID), zap.Bool("is_essence", req.IsEssence), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	exists, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查帖子失败: %w", err)
	}
	if !exists {
		return errors.New("帖子不存在")
	}

	// 设置精华状态
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetIsEssence(req.IsEssence).
		Save(ctx)
	if err != nil {
		s.logger.Error("设置帖子精华失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("设置帖子精华失败: %w", err)
	}

	s.logger.Info("帖子精华设置成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// SetPostPin 设置帖子置顶
func (s *PostManageService) SetPostPin(ctx context.Context, req schema.PostPinUpdateRequest) error {
	s.logger.Info("设置帖子置顶", zap.Int("id", req.ID), zap.Bool("is_pinned", req.IsPinned), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	exists, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查帖子失败: %w", err)
	}
	if !exists {
		return errors.New("帖子不存在")
	}

	// 设置置顶状态
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetIsPinned(req.IsPinned).
		Save(ctx)
	if err != nil {
		s.logger.Error("设置帖子置顶失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("设置帖子置顶失败: %w", err)
	}

	s.logger.Info("帖子置顶设置成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// MovePost 移动帖子到其他版块
func (s *PostManageService) MovePost(ctx context.Context, req schema.PostMoveRequest) error {
	s.logger.Info("移动帖子", zap.Int("id", req.ID), zap.Int("category_id", req.CategoryID), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	postExists, err := s.db.Post.Query().
		Where(post.IDEQ(req.ID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查帖子失败: %w", err)
	}
	if !postExists {
		return errors.New("帖子不存在")
	}

	// 检查目标版块是否存在
	categoryExists, err := s.db.Category.Query().
		Where(category.IDEQ(req.CategoryID)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查目标版块失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查目标版块失败: %w", err)
	}
	if !categoryExists {
		return errors.New("目标版块不存在")
	}

	// 移动帖子
	_, err = s.db.Post.UpdateOneID(req.ID).
		SetCategoryID(req.CategoryID).
		Save(ctx)
	if err != nil {
		s.logger.Error("移动帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("移动帖子失败: %w", err)
	}

	s.logger.Info("帖子移动成功", zap.Int("id", req.ID), tracing.WithTraceIDField(ctx))
	return nil
}

// DeletePost 删除帖子（软删除，状态设为Ban）
func (s *PostManageService) DeletePost(ctx context.Context, id int) error {
	s.logger.Info("删除帖子", zap.Int("id", id), tracing.WithTraceIDField(ctx))

	// 检查帖子是否存在
	exists, err := s.db.Post.Query().
		Where(post.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		s.logger.Error("检查帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查帖子失败: %w", err)
	}
	if !exists {
		return errors.New("帖子不存在")
	}

	// 软删除：将状态设为Ban
	_, err = s.db.Post.UpdateOneID(id).
		SetStatus(post.StatusBan).
		Save(ctx)
	if err != nil {
		s.logger.Error("删除帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("删除帖子失败: %w", err)
	}

	s.logger.Info("帖子删除成功", zap.Int("id", id), tracing.WithTraceIDField(ctx))
	return nil
}
