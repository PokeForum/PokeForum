package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/category"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/postaction"
	"github.com/PokeForum/PokeForum/ent/user"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	smtp "github.com/PokeForum/PokeForum/internal/pkg/email"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// IUserProfileService 用户个人中心服务接口
type IUserProfileService interface {
	// GetProfileOverview 获取用户个人中心概览
	GetProfileOverview(ctx context.Context, userID int, isOwner bool) (*schema.UserProfileOverviewResponse, error)
	// GetUserPosts 获取用户主题帖列表
	GetUserPosts(ctx context.Context, userID int, req schema.UserProfilePostsRequest, isOwner bool) (*schema.UserProfilePostsResponse, error)
	// GetUserComments 获取用户评论列表
	GetUserComments(ctx context.Context, userID int, req schema.UserProfileCommentsRequest, isOwner bool) (*schema.UserProfileCommentsResponse, error)
	// GetUserFavorites 获取用户收藏列表
	GetUserFavorites(ctx context.Context, userID int, req schema.UserProfileFavoritesRequest, isOwner bool) (*schema.UserProfileFavoritesResponse, error)
	// UpdatePassword 修改密码
	UpdatePassword(ctx context.Context, userID int, req schema.UserUpdatePasswordRequest) (*schema.UserUpdatePasswordResponse, error)
	// UpdateAvatar 修改头像
	UpdateAvatar(ctx context.Context, userID int, req schema.UserUpdateAvatarRequest) (*schema.UserUpdateAvatarResponse, error)
	// UpdateUsername 修改用户名(每七日可操作一次)
	UpdateUsername(ctx context.Context, userID int, req schema.UserUpdateUsernameRequest) (*schema.UserUpdateUsernameResponse, error)
	// SendEmailVerifyCode 发送邮箱验证码（直接发送到用户注册邮箱）
	SendEmailVerifyCode(ctx context.Context, userID int) (*schema.EmailVerifyCodeResponse, error)
	// VerifyEmail 验证邮箱
	VerifyEmail(ctx context.Context, userID int, req schema.EmailVerifyRequest) (*schema.EmailVerifyResponse, error)
	// CheckUsernameUpdatePermission 检查用户名修改权限(每七日可操作一次)
	CheckUsernameUpdatePermission(ctx context.Context, userID int) (bool, error)
	// TODO 粉丝列表
}

// UserProfileService 用户个人中心服务实现
type UserProfileService struct {
	db                *ent.Client
	cache             cache.ICacheService
	logger            *zap.Logger
	settings          ISettingsService
	userManageService IUserManageService
}

// NewUserProfileService 创建用户个人中心服务实例
func NewUserProfileService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger, settingsService ISettingsService, userManageService IUserManageService) IUserProfileService {
	return &UserProfileService{
		db:                db,
		cache:             cacheService,
		logger:            logger,
		settings:          settingsService,
		userManageService: userManageService,
	}
}

// GetProfileOverview 获取用户个人中心概览
func (s *UserProfileService) GetProfileOverview(ctx context.Context, userID int, isOwner bool) (*schema.UserProfileOverviewResponse, error) {
	s.logger.Info("获取用户个人中心概览", zap.Int("user_id", userID), zap.Bool("is_owner", isOwner), tracing.WithTraceIDField(ctx))

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

	// 实时查询用户的发帖数和评论数
	postCount, err := s.userManageService.GetUserPostCount(ctx, userData.ID)
	if err != nil {
		s.logger.Error("查询用户发帖数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		postCount = 0
	}

	commentCount, err := s.userManageService.GetUserCommentCount(ctx, userData.ID)
	if err != nil {
		s.logger.Error("查询用户评论数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		commentCount = 0
	}

	// 构建响应数据
	result := &schema.UserProfileOverviewResponse{
		ID:           userData.ID,
		Username:     userData.Username,
		Avatar:       userData.Avatar,
		Signature:    userData.Signature,
		Readme:       userData.Readme,
		PostCount:    postCount,
		CommentCount: commentCount,
		Status:       string(userData.Status),
		Role:         string(userData.Role),
		CreatedAt:    userData.CreatedAt.Format(time_tools.DateTimeFormat),
	}

	// 只有本人才能看到敏感数据
	if isOwner {
		result.Email = userData.Email
		result.EmailVerified = userData.EmailVerified
		result.Points = userData.Points
		result.Currency = userData.Currency
	} else {
		// 他人查看时屏蔽敏感数据
		result.Email = ""
		result.EmailVerified = false
		result.Points = 0
		result.Currency = 0
	}

	s.logger.Info("获取用户个人中心概览成功", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetUserPosts 获取用户主题帖列表
func (s *UserProfileService) GetUserPosts(ctx context.Context, userID int, req schema.UserProfilePostsRequest, isOwner bool) (*schema.UserProfilePostsResponse, error) {
	s.logger.Info("获取用户主题帖列表", zap.Int("user_id", userID), zap.Int("page", req.Page), zap.Bool("is_owner", isOwner), tracing.WithTraceIDField(ctx))

	// 构建查询条件
	query := s.db.Post.Query().
		Where(post.UserIDEQ(userID))

	// 如果指定了状态，则筛选状态
	if req.Status != "" {
		query = query.Where(post.StatusEQ(post.Status(req.Status)))
	} else if !isOwner {
		// 他人查看时，只显示正常状态的帖子
		query = query.Where(post.StatusEQ(post.StatusNormal))
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

	// 收集版块ID
	categoryIDs := make(map[int]bool)
	for _, p := range posts {
		categoryIDs[p.CategoryID] = true
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

	// 构建响应数据
	list := make([]schema.UserProfilePostItem, len(posts))
	for i, p := range posts {
		categoryName := categoryMap[p.CategoryID]

		list[i] = schema.UserProfilePostItem{
			ID:            p.ID,
			CategoryID:    p.CategoryID,
			CategoryName:  categoryName,
			Title:         p.Title,
			ViewCount:     p.ViewCount,
			LikeCount:     p.LikeCount,
			DislikeCount:  p.DislikeCount,
			FavoriteCount: p.FavoriteCount,
			IsEssence:     p.IsEssence,
			IsPinned:      p.IsPinned,
			Status:        string(p.Status),
			CreatedAt:     p.CreatedAt.Format(time_tools.DateTimeFormat),
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
func (s *UserProfileService) GetUserComments(ctx context.Context, userID int, req schema.UserProfileCommentsRequest, isOwner bool) (*schema.UserProfileCommentsResponse, error) {
	s.logger.Info("获取用户评论列表", zap.Int("user_id", userID), zap.Int("page", req.Page), zap.Bool("is_owner", isOwner), tracing.WithTraceIDField(ctx))

	// 构建基础查询条件
	baseQuery := s.db.Comment.Query().
		Where(comment.UserIDEQ(userID))

	// 获取过滤后的总数
	totalQuery := s.db.Comment.Query().Where(comment.UserIDEQ(userID))
	if !isOwner {
		// 他人查看时，需要过滤掉私有/封禁帖子的评论
		// 先获取该用户所有评论关联的帖子ID
		allCommentPostIDs, err := baseQuery.Select(comment.FieldPostID).Strings(ctx)
		if err != nil {
			s.logger.Error("获取用户评论帖子ID失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("获取用户评论帖子ID失败: %w", err)
		}

		// 转换为int数组
		postIDs := make([]int, 0, len(allCommentPostIDs))
		for _, idStr := range allCommentPostIDs {
			if id, err := strconv.Atoi(idStr); err == nil {
				postIDs = append(postIDs, id)
			}
		}

		// 批量查询帖子状态
		if len(postIDs) > 0 {
			posts, err := s.db.Post.Query().
				Where(post.IDIn(postIDs...)).
				Select(post.FieldID, post.FieldStatus).
				All(ctx)
			if err != nil {
				s.logger.Error("获取帖子状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
				return nil, fmt.Errorf("获取帖子状态失败: %w", err)
			}

			// 筛选公开帖子的ID
			publicPostIDs := make([]int, 0)
			for _, postData := range posts {
				if postData.Status != post.StatusPrivate && postData.Status != post.StatusBan {
					publicPostIDs = append(publicPostIDs, postData.ID)
				}
			}

			// 只统计公开帖子的评论
			if len(publicPostIDs) > 0 {
				totalQuery = totalQuery.Where(comment.PostIDIn(publicPostIDs...))
			} else {
				// 没有公开帖子，总数为0
				totalQuery = totalQuery.Where(comment.PostIDIn(-1))
			}
		} else {
			// 没有评论，总数为0
			totalQuery = totalQuery.Where(comment.PostIDIn(-1))
		}
	}
	total, err := totalQuery.Count(ctx)
	if err != nil {
		s.logger.Error("获取用户评论总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户评论总数失败: %w", err)
	}

	// 分页查询
	comments, err := baseQuery.
		Order(ent.Desc(comment.FieldCreatedAt)).
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
	postTitleMap := make(map[int]*ent.Post)
	for _, postData := range posts {
		postTitleMap[postData.ID] = postData
	}

	// 构建响应数据
	list := make([]schema.UserProfileCommentItem, 0, len(comments))
	for _, commentData := range comments {
		postData, ok := postTitleMap[commentData.PostID]
		if !ok {
			// 帖子不存在，跳过
			continue
		}

		// 他人查看时，过滤掉私有或封禁帖子的评论
		if !isOwner && (postData.Status == post.StatusPrivate || postData.Status == post.StatusBan) {
			continue
		}

		list = append(list, schema.UserProfileCommentItem{
			ID:           commentData.ID,
			PostID:       commentData.PostID,
			PostTitle:    postData.Title,
			Content:      commentData.Content,
			LikeCount:    commentData.LikeCount,
			DislikeCount: commentData.DislikeCount,
			CreatedAt:    commentData.CreatedAt.Format(time_tools.DateTimeFormat),
		})
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
func (s *UserProfileService) GetUserFavorites(ctx context.Context, userID int, req schema.UserProfileFavoritesRequest, isOwner bool) (*schema.UserProfileFavoritesResponse, error) {
	s.logger.Info("获取用户收藏列表", zap.Int("user_id", userID), zap.Int("page", req.Page), zap.Bool("is_owner", isOwner), tracing.WithTraceIDField(ctx))

	// 构建基础查询条件
	baseQuery := s.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.ActionTypeEQ(postaction.ActionTypeFavorite),
		)

	// 获取过滤后的总数
	totalQuery := s.db.PostAction.Query().
		Where(
			postaction.UserIDEQ(userID),
			postaction.ActionTypeEQ(postaction.ActionTypeFavorite),
		)
	if !isOwner {
		// 他人查看时，需要过滤掉私有/封禁帖子的收藏
		// 先获取该用户所有收藏关联的帖子ID
		allFavoritePostIDs, err := baseQuery.Select(postaction.FieldPostID).Strings(ctx)
		if err != nil {
			s.logger.Error("获取用户收藏帖子ID失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("获取用户收藏帖子ID失败: %w", err)
		}

		// 转换为int数组
		postIDs := make([]int, 0, len(allFavoritePostIDs))
		for _, idStr := range allFavoritePostIDs {
			if id, err := strconv.Atoi(idStr); err == nil {
				postIDs = append(postIDs, id)
			}
		}

		// 批量查询帖子状态
		if len(postIDs) > 0 {
			posts, err := s.db.Post.Query().
				Where(post.IDIn(postIDs...)).
				Select(post.FieldID, post.FieldStatus).
				All(ctx)
			if err != nil {
				s.logger.Error("获取帖子状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
				return nil, fmt.Errorf("获取帖子状态失败: %w", err)
			}

			// 筛选公开帖子的ID
			publicPostIDs := make([]int, 0)
			for _, postData := range posts {
				if postData.Status != post.StatusPrivate && postData.Status != post.StatusBan {
					publicPostIDs = append(publicPostIDs, postData.ID)
				}
			}

			// 只统计公开帖子的收藏
			if len(publicPostIDs) > 0 {
				totalQuery = totalQuery.Where(postaction.PostIDIn(publicPostIDs...))
			} else {
				// 没有公开帖子，总数为0
				totalQuery = totalQuery.Where(postaction.PostIDIn(-1))
			}
		} else {
			// 没有收藏，总数为0
			totalQuery = totalQuery.Where(postaction.PostIDIn(-1))
		}
	}
	total, err := totalQuery.Count(ctx)
	if err != nil {
		s.logger.Error("获取用户收藏总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户收藏总数失败: %w", err)
	}

	// 分页查询
	favorites, err := baseQuery.
		Order(ent.Desc(postaction.FieldCreatedAt)).
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取用户收藏列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取用户收藏列表失败: %w", err)
	}

	// 收集帖子ID
	postIDs := make([]int, len(favorites))
	for i, fav := range favorites {
		postIDs[i] = fav.PostID
	}

	// 批量查询帖子信息
	posts, _ := s.db.Post.Query().
		Where(post.IDIn(postIDs...)).
		All(ctx)
	postMap := make(map[int]*ent.Post)
	for _, postData := range posts {
		// 他人查看时，过滤掉私有或封禁帖子
		if !isOwner && (postData.Status == post.StatusPrivate || postData.Status == post.StatusBan) {
			continue
		}
		postMap[postData.ID] = postData
	}

	// 收集用户ID和版块ID
	userIDs := make(map[int]bool)
	categoryIDs := make(map[int]bool)
	for _, postData := range postMap {
		userIDs[postData.UserID] = true
		categoryIDs[postData.CategoryID] = true
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

	// 构建响应数据
	list := make([]schema.UserProfileFavoriteItem, 0, len(favorites))
	for _, fav := range favorites {
		postData := postMap[fav.PostID]
		if postData == nil {
			continue
		}

		username := userMap[postData.UserID]
		categoryName := categoryMap[postData.CategoryID]

		list = append(list, schema.UserProfileFavoriteItem{
			ID:            postData.ID,
			CategoryID:    postData.CategoryID,
			CategoryName:  categoryName,
			Title:         postData.Title,
			Username:      username,
			ViewCount:     postData.ViewCount,
			LikeCount:     postData.LikeCount,
			FavoriteCount: postData.FavoriteCount,
			CreatedAt:     postData.CreatedAt.Format(time_tools.DateTimeFormat),
			FavoritedAt:   fav.CreatedAt.Format(time_tools.DateTimeFormat),
		})
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

// CheckUsernameUpdatePermission 检查用户名修改权限(每七日可操作一次)
func (s *UserProfileService) CheckUsernameUpdatePermission(ctx context.Context, userID int) (bool, error) {
	// 生成Redis键名：user:username:update:limit:{userID}
	redisKey := fmt.Sprintf("user:username:update:limit:%d", userID)

	// 检查是否在限制期内
	lastUpdateTime, err := s.cache.Get(ctx, redisKey)
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
	err = s.cache.SetEx(ctx, redisKey, currentTime, 604800) // 604800秒 = 7天
	if err != nil {
		s.logger.Error("设置用户名修改限制失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 设置失败，但允许操作
		return true, nil
	}

	return true, nil
}

// SendEmailVerifyCode 发送邮箱验证码（直接发送到用户注册邮箱）
func (s *UserProfileService) SendEmailVerifyCode(ctx context.Context, userID int) (*schema.EmailVerifyCodeResponse, error) {
	s.logger.Info("发送邮箱验证码", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 查询用户信息
	userData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Select(user.FieldEmail, user.FieldEmailVerified).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("用户不存在")
		}
		s.logger.Error("查询用户失败", zap.Int("user_id", userID), zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查邮箱是否已验证
	if userData.EmailVerified {
		return nil, errors.New("邮箱已验证，无需重复验证")
	}

	// 检查发送频率限制
	limitKey := fmt.Sprintf("email:verify:limit:%d", userID)
	limitValue, err := s.cache.Get(ctx, limitKey)
	if err == nil && limitValue != "" {
		sendCount := 0
		if _, parseErr := fmt.Sscanf(limitValue, "%d", &sendCount); parseErr == nil && sendCount >= 3 {
			return nil, errors.New("发送次数过多，请1小时后再试")
		}
	}

	// 生成6位随机验证码
	code, err := s.generateVerifyCode()
	if err != nil {
		s.logger.Error("生成验证码失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存储验证码到Redis（10分钟有效期）
	codeKey := fmt.Sprintf("email:verify:code:%d", userID)
	err = s.cache.SetEx(ctx, codeKey, code, 600)
	if err != nil {
		s.logger.Error("存储验证码失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("存储验证码失败: %w", err)
	}

	// 更新发送频率限制
	newCount := 1
	if limitValue != "" {
		if val, err := strconv.Atoi(limitValue); err == nil {
			newCount = val + 1
		}
	}
	if err := s.cache.SetEx(ctx, limitKey, fmt.Sprintf("%d", newCount), 3600); err != nil {
		s.logger.Warn("更新发送频率限制失败", zap.String("key", limitKey), zap.Error(err), tracing.WithTraceIDField(ctx))
	}

	// 发送验证邮件
	err = s.sendVerificationEmail(ctx, userData.Email, code)
	if err != nil {
		s.logger.Error("发送验证邮件失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("发送验证邮件失败: %w", err)
	}

	s.logger.Info("邮箱验证码发送成功", zap.Int("user_id", userID), zap.String("email", userData.Email), tracing.WithTraceIDField(ctx))

	return &schema.EmailVerifyCodeResponse{
		Sent:      true,
		Message:   "验证码已发送到您的邮箱，请查收",
		ExpiresIn: 600,
	}, nil
}

// VerifyEmail 验证邮箱
func (s *UserProfileService) VerifyEmail(ctx context.Context, userID int, req schema.EmailVerifyRequest) (*schema.EmailVerifyResponse, error) {
	s.logger.Info("验证邮箱", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	// 获取存储的验证码数据
	codeKey := fmt.Sprintf("email:verify:code:%d", userID)
	storedCode, err := s.cache.Get(ctx, codeKey)
	if err != nil || storedCode == "" {
		return nil, errors.New("验证码不存在或已过期")
	}

	// 验证验证码
	if storedCode != req.Code {
		return nil, errors.New("验证码错误")
	}

	// 更新用户邮箱验证状态
	_, err = s.db.User.UpdateOneID(userID).
		SetEmailVerified(true).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新邮箱验证状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新邮箱验证状态失败: %w", err)
	}

	// 清除验证码缓存
	count, _ := s.cache.Del(ctx, codeKey)
	s.logger.Debug("清除验证码缓存", zap.Int("count", count), tracing.WithTraceIDField(ctx))

	s.logger.Info("邮箱验证成功", zap.Int("user_id", userID), tracing.WithTraceIDField(ctx))

	return &schema.EmailVerifyResponse{
		Verified: true,
		Message:  "邮箱验证成功",
	}, nil
}

// generateVerifyCode 生成6位随机验证码
func (s *UserProfileService) generateVerifyCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += num.String()
	}
	return code, nil
}

// sendVerificationEmail 发送验证邮件
func (s *UserProfileService) sendVerificationEmail(ctx context.Context, email, code string) error {
	// 检查是否启用了邮箱验证
	isVerifyEmail, err := s.settings.GetSettingByKey(ctx, _const.SafeVerifyEmail, _const.SettingBoolTrue.String())
	if err != nil {
		s.logger.Error("查询邮箱验证设置失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("查询邮箱验证设置失败: %w", err)
	}

	if isVerifyEmail != _const.SettingBoolTrue.String() {
		return errors.New("邮箱验证功能未启用")
	}

	// 获取网站设置
	siteConfig, err := s.settings.GetSeoSettings(ctx)
	if err != nil {
		s.logger.Error("获取网站配置失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取网站配置失败: %w", err)
	}

	// 获取SMTP配置
	smtpConfig, err := s.settings.GetSMTPConfig(ctx)
	if err != nil {
		s.logger.Error("获取SMTP配置失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取SMTP配置失败: %w", err)
	}

	// 创建邮件模板渲染器
	emailTemplate := smtp.NewEmailTemplate(s.settings, s.logger)

	// 渲染邮件模板
	htmlBody, err := emailTemplate.RenderEmailVerificationTemplate(ctx, code, siteConfig.WebSiteName)
	if err != nil {
		s.logger.Error("渲染邮件模板失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("渲染邮件模板失败: %w", err)
	}

	// 发送邮件
	sp := smtp.NewSMTPPool(smtp.SMTPConfig{
		Name:       siteConfig.WebSiteName,
		Address:    smtpConfig.Address,
		Host:       smtpConfig.Host,
		Port:       smtpConfig.Port,
		User:       smtpConfig.Username,
		Password:   smtpConfig.Password,
		Encryption: smtpConfig.ForcedSSL,
		Keepalive:  smtpConfig.ConnectionValidity,
	}, s.logger)
	defer sp.Close()

	if err = sp.Send(ctx, email, fmt.Sprintf("【%s】邮箱验证码", siteConfig.WebSiteName), htmlBody); err != nil {
		return err
	}

	return nil
}

// hashPassword 生成密码哈希值
func hashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}
