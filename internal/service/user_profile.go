package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/postaction"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// IUserProfileService 用户个人中心服务接口
type IUserProfileService interface {
	// GetProfileOverview 获取用户个人中心概览
	GetProfileOverview(ctx context.Context, userID int) (*schema.UserProfileOverviewResponse, error)
	// GetUserPosts 获取用户主题帖列表
	GetUserPosts(ctx context.Context, userID int, req schema.UserProfilePostsRequest) (*schema.UserProfilePostsResponse, error)
	// GetUserComments 获取用户评论列表
	GetUserComments(ctx context.Context, userID int, req schema.UserProfileCommentsRequest) (*schema.UserProfileCommentsResponse, error)
	// GetUserFavorites 获取用户收藏列表
	GetUserFavorites(ctx context.Context, userID int, req schema.UserProfileFavoritesRequest) (*schema.UserProfileFavoritesResponse, error)
	// UpdatePassword 修改密码
	UpdatePassword(ctx context.Context, userID int, req schema.UserUpdatePasswordRequest) (*schema.UserUpdatePasswordResponse, error)
	// UpdateAvatar 修改头像
	UpdateAvatar(ctx context.Context, userID int, req schema.UserUpdateAvatarRequest) (*schema.UserUpdateAvatarResponse, error)
	// UpdateUsername 修改用户名(每七日可操作一次)
	UpdateUsername(ctx context.Context, userID int, req schema.UserUpdateUsernameRequest) (*schema.UserUpdateUsernameResponse, error)
	// UpdateEmail 修改邮箱
	UpdateEmail(ctx context.Context, userID int, req schema.UserUpdateEmailRequest) (*schema.UserUpdateEmailResponse, error)
	// CheckUsernameUpdatePermission 检查用户名修改权限(每七日可操作一次)
	CheckUsernameUpdatePermission(ctx context.Context, userID int) (bool, error)
	// TODO 粉丝列表
}

// UserProfileService 用户个人中心服务实现
type UserProfileService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewUserProfileService 创建用户个人中心服务实例
func NewUserProfileService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) IUserProfileService {
	return &UserProfileService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// GetProfileOverview 获取用户个人中心概览
func (s *UserProfileService) GetProfileOverview(ctx context.Context, userID int) (*schema.UserProfileOverviewResponse, error) {
	s.logger.Info("获取用户个人中心概览", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 查询用户信息
	userData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 构建响应数据
	result := &schema.UserProfileOverviewResponse{
		ID:            userData.ID,
		Username:      userData.Username,
		Email:         userData.Email,
		Avatar:        userData.Avatar,
		Signature:     userData.Signature,
		Readme:        userData.Readme,
		EmailVerified: userData.EmailVerified,
		Points:        userData.Points,
		Currency:      userData.Currency,
		PostCount:     userData.PostCount,
		CommentCount:  userData.CommentCount,
		Status:        string(userData.Status),
		Role:          string(userData.Role),
		CreatedAt:     userData.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	s.logger.Info("获取用户个人中心概览成功", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetUserPosts 获取用户主题帖列表
func (s *UserProfileService) GetUserPosts(ctx context.Context, userID int, req schema.UserProfilePostsRequest) (*schema.UserProfilePostsResponse, error) {
	s.logger.Info("获取用户主题帖列表", zap.Int("user_id", userID), zap.Int("page", req.Page), tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.Post.Query().
		WithCategory().
		Where(post.UserIDEQ(userID))

	// 如果指定了状态，则筛选状态
	if req.Status != "" {
		query = query.Where(post.StatusEQ(post.Status(req.Status)))
	}

	// 按创建时间倒序排序
	query = query.Order(ent.Desc(post.FieldCreatedAt))

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取用户主题帖总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户主题帖总数失败: %w", err)
	}

	// 分页查询
	posts, err := query.
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取用户主题帖列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户主题帖列表失败: %w", err)
	}

	// 构建响应数据
	list := make([]schema.UserProfilePostItem, len(posts))
	for i, postData := range posts {
		categoryName := ""
		if postData.Edges.Category != nil {
			categoryName = postData.Edges.Category.Name
		}

		list[i] = schema.UserProfilePostItem{
			ID:            postData.ID,
			CategoryID:    postData.CategoryID,
			CategoryName:  categoryName,
			Title:         postData.Title,
			ViewCount:     postData.ViewCount,
			LikeCount:     postData.LikeCount,
			DislikeCount:  postData.DislikeCount,
			FavoriteCount: postData.FavoriteCount,
			IsEssence:     postData.IsEssence,
			IsPinned:      postData.IsPinned,
			Status:        string(postData.Status),
			CreatedAt:     postData.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	result := &schema.UserProfilePostsResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	s.logger.Info("获取用户主题帖列表成功", zap.Int("user_id", userID), zap.Int("total", total), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetUserComments 获取用户评论列表
func (s *UserProfileService) GetUserComments(ctx context.Context, userID int, req schema.UserProfileCommentsRequest) (*schema.UserProfileCommentsResponse, error) {
	s.logger.Info("获取用户评论列表", zap.Int("user_id", userID), zap.Int("page", req.Page), tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.Comment.Query().
		Where(comment.UserIDEQ(userID)).
		Order(ent.Desc(comment.FieldCreatedAt))

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取用户评论总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户评论总数失败: %w", err)
	}

	// 分页查询
	comments, err := query.
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取用户评论列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户评论列表失败: %w", err)
	}

	// 获取评论对应的帖子信息
	postIDs := make([]int, len(comments))
	for i, commentData := range comments {
		postIDs[i] = commentData.PostID
	}

	// 批量查询帖子信息
	posts, err := s.db.Post.Query().
		Where(post.IDIn(postIDs...)).
		All(ctx)
	if err != nil {
		s.logger.Error("获取帖子信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子信息失败: %w", err)
	}

	// 构建帖子ID到帖子标题的映射
	postTitleMap := make(map[int]string)
	for _, postData := range posts {
		postTitleMap[postData.ID] = postData.Title
	}

	// 构建响应数据
	list := make([]schema.UserProfileCommentItem, len(comments))
	for i, commentData := range comments {
		postTitle := ""
		if title, ok := postTitleMap[commentData.PostID]; ok {
			postTitle = title
		}

		list[i] = schema.UserProfileCommentItem{
			ID:           commentData.ID,
			PostID:       commentData.PostID,
			PostTitle:    postTitle,
			Content:      commentData.Content,
			LikeCount:    commentData.LikeCount,
			DislikeCount: commentData.DislikeCount,
			CreatedAt:    commentData.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	result := &schema.UserProfileCommentsResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	s.logger.Info("获取用户评论列表成功", zap.Int("user_id", userID), zap.Int("total", total), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetUserFavorites 获取用户收藏列表
func (s *UserProfileService) GetUserFavorites(ctx context.Context, userID int, req schema.UserProfileFavoritesRequest) (*schema.UserProfileFavoritesResponse, error) {
	s.logger.Info("获取用户收藏列表", zap.Int("user_id", userID), zap.Int("page", req.Page), tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.ActionTypeEQ(postaction.ActionTypeFavorite),
		).
		WithPost(func(q *ent.PostQuery) {
			q.WithCategory().WithAuthor()
		}).
		Order(ent.Desc(postaction.FieldCreatedAt))

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取用户收藏总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户收藏总数失败: %w", err)
	}

	// 分页查询
	favorites, err := query.
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取用户收藏列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户收藏列表失败: %w", err)
	}

	// 构建响应数据
	list := make([]schema.UserProfileFavoriteItem, len(favorites))
	for i, favorite := range favorites {
		postData := favorite.Edges.Post
		if postData == nil {
			continue
		}

		categoryName := ""
		if postData.Edges.Category != nil {
			categoryName = postData.Edges.Category.Name
		}

		username := ""
		if postData.Edges.Author != nil {
			username = postData.Edges.Author.Username
		}

		list[i] = schema.UserProfileFavoriteItem{
			ID:            postData.ID,
			CategoryID:    postData.CategoryID,
			CategoryName:  categoryName,
			Title:         postData.Title,
			Username:      username,
			ViewCount:     postData.ViewCount,
			LikeCount:     postData.LikeCount,
			FavoriteCount: postData.FavoriteCount,
			CreatedAt:     postData.CreatedAt.Format("2006-01-02T15:04:05Z"),
			FavoritedAt:   favorite.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	result := &schema.UserProfileFavoritesResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	s.logger.Info("获取用户收藏列表成功", zap.Int("user_id", userID), zap.Int("total", total), tracing.WithTraceIDField(ctx))
	return result, nil
}

// UpdatePassword 修改密码
func (s *UserProfileService) UpdatePassword(ctx context.Context, userID int, req schema.UserUpdatePasswordRequest) (*schema.UserUpdatePasswordResponse, error) {
	s.logger.Info("修改密码", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 查询用户信息
	userData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 验证旧密码
	oldPasswordHash := hashPassword(req.OldPassword, userData.PasswordSalt)
	if oldPasswordHash != userData.Password {
		return nil, errors.New("旧密码错误")
	}

	// 生成新的密码哈希
	newPasswordHash := hashPassword(req.NewPassword, userData.PasswordSalt)

	// 更新密码
	_, err = s.db.User.UpdateOneID(userID).
		SetPassword(newPasswordHash).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新密码失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新密码失败: %w", err)
	}

	result := &schema.UserUpdatePasswordResponse{
		Success: true,
		Message: "密码修改成功，请重新登录",
	}

	s.logger.Info("密码修改成功", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// UpdateAvatar 修改头像
func (s *UserProfileService) UpdateAvatar(ctx context.Context, userID int, req schema.UserUpdateAvatarRequest) (*schema.UserUpdateAvatarResponse, error) {
	s.logger.Info("修改头像", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 更新头像
	_, err := s.db.User.UpdateOneID(userID).
		SetAvatar(req.AvatarURL).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新头像失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新头像失败: %w", err)
	}

	result := &schema.UserUpdateAvatarResponse{
		Success:   true,
		AvatarURL: req.AvatarURL,
	}

	s.logger.Info("头像修改成功", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// UpdateUsername 修改用户名(每七日可操作一次)
func (s *UserProfileService) UpdateUsername(ctx context.Context, userID int, req schema.UserUpdateUsernameRequest) (*schema.UserUpdateUsernameResponse, error) {
	s.logger.Info("修改用户名", zap.Int("user_id", userID), zap.String("new_username", req.Username), tracing.WithTraceIDField(ctx))

	// 检查修改权限（每七日可操作一次）
	canUpdate, err := s.CheckUsernameUpdatePermission(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !canUpdate {
		return nil, errors.New("用户名修改过于频繁，请七天后再试")
	}

	// 检查用户名是否已存在
	existingUser, err := s.db.User.Query().
		Where(user.UsernameEQ(req.Username)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Error("检查用户名失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if existingUser != nil && existingUser.ID != userID {
		return nil, errors.New("用户名已被使用")
	}

	// 更新用户名
	_, err = s.db.User.UpdateOneID(userID).
		SetUsername(req.Username).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新用户名失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新用户名失败: %w", err)
	}

	result := &schema.UserUpdateUsernameResponse{
		Success:  true,
		Username: req.Username,
	}

	s.logger.Info("用户名修改成功", zap.Int("user_id", userID), zap.String("new_username", req.Username), tracing.WithTraceIDField(ctx))
	return result, nil
}

// UpdateEmail 修改邮箱
func (s *UserProfileService) UpdateEmail(ctx context.Context, userID int, req schema.UserUpdateEmailRequest) (*schema.UserUpdateEmailResponse, error) {
	s.logger.Info("修改邮箱", zap.Int("user_id", userID), zap.String("new_email", req.NewEmail), tracing.WithTraceIDField(ctx))

	// 查询用户信息
	userData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("获取用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 验证密码
	passwordHash := hashPassword(req.Password, userData.PasswordSalt)
	if passwordHash != userData.Password {
		return nil, errors.New("密码错误")
	}

	// TODO: 验证邮箱验证码
	// 这里需要集成邮箱验证码服务，暂时省略验证码校验逻辑

	// 检查邮箱是否已存在
	existingUser, err := s.db.User.Query().
		Where(user.EmailEQ(req.NewEmail)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Error("检查邮箱失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existingUser != nil && existingUser.ID != userID {
		return nil, errors.New("邮箱已被使用")
	}

	// 更新邮箱，并设置邮箱未验证
	_, err = s.db.User.UpdateOneID(userID).
		SetEmail(req.NewEmail).
		SetEmailVerified(false).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新邮箱失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新邮箱失败: %w", err)
	}

	result := &schema.UserUpdateEmailResponse{
		Success: true,
		Email:   req.NewEmail,
		Message: "邮箱修改成功，请重新验证邮箱",
	}

	s.logger.Info("邮箱修改成功", zap.Int("user_id", userID), zap.String("new_email", req.NewEmail), tracing.WithTraceIDField(ctx))
	return result, nil
}

// CheckUsernameUpdatePermission 检查用户名修改权限(每七日可操作一次)
func (s *UserProfileService) CheckUsernameUpdatePermission(ctx context.Context, userID int) (bool, error) {
	// 生成Redis键名：user:username:update:limit:{userID}
	redisKey := fmt.Sprintf("user:username:update:limit:%d", userID)

	// 检查是否在限制期内
	lastUpdateTime, err := s.cache.Get(redisKey)
	if err != nil {
		s.logger.Error("获取用户名修改限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return false, fmt.Errorf("获取用户名修改限制失败: %w", err)
	}

	if lastUpdateTime != "" {
		// 解析最后修改时间
		lastTime, err := time.Parse(time.RFC3339, lastUpdateTime)
		if err != nil {
			s.logger.Error("解析最后修改时间失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			// 解析失败，允许操作
			return true, nil
		}

		// 检查是否在七日内
		if time.Since(lastTime) < 7*24*time.Hour {
			s.logger.Warn("用户名修改操作过于频繁", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))
			return false, nil
		}
	}

	// 设置新的修改时间限制
	currentTime := time.Now().Format(time.RFC3339)
	err = s.cache.SetEx(redisKey, currentTime, 604800) // 604800秒 = 7天
	if err != nil {
		s.logger.Error("设置用户名修改限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 设置失败，但允许操作
		return true, nil
	}

	return true, nil
}

// hashPassword 生成密码哈希值
func hashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}
