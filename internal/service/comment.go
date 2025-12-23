package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/comment"
	"github.com/PokeForum/PokeForum/ent/commentaction"
	"github.com/PokeForum/PokeForum/ent/post"
	"github.com/PokeForum/PokeForum/ent/user"
	_const "github.com/PokeForum/PokeForum/internal/consts"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/stats"
	"github.com/PokeForum/PokeForum/internal/pkg/time_tools"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"go.uber.org/zap"
)

// ICommentService 评论服务接口
type ICommentService interface {
	// CreateComment 创建评论
	CreateComment(ctx context.Context, userID int, clientIP, deviceInfo string, req schema.UserCommentCreateRequest) (*schema.UserCommentCreateResponse, error)
	// UpdateComment 更新评论
	UpdateComment(ctx context.Context, userID int, req schema.UserCommentUpdateRequest) (*schema.UserCommentUpdateResponse, error)
	// LikeComment 点赞评论
	LikeComment(ctx context.Context, userID int, req schema.UserCommentActionRequest) (*schema.UserCommentActionResponse, error)
	// DislikeComment 点踩评论
	DislikeComment(ctx context.Context, userID int, req schema.UserCommentActionRequest) (*schema.UserCommentActionResponse, error)
	// GetCommentList 获取评论列表
	GetCommentList(ctx context.Context, req schema.UserCommentListRequest) (*schema.UserCommentListResponse, error)
}

// CommentService 评论服务实现
type CommentService struct {
	db                  *ent.Client
	cache               cache.ICacheService
	logger              *zap.Logger
	commentStatsService ICommentStatsService
	settingsService     ISettingsService
}

// NewCommentService 创建评论服务实例
func NewCommentService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) ICommentService {
	return &CommentService{
		db:                  db,
		cache:               cacheService,
		logger:              logger,
		commentStatsService: NewCommentStatsService(db, cacheService, logger),
		settingsService:     NewSettingsService(db, cacheService, logger),
	}
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(ctx context.Context, userID int, clientIP, deviceInfo string, req schema.UserCommentCreateRequest) (*schema.UserCommentCreateResponse, error) {
	s.logger.Info("创建评论", zap.Int("user_id", userID), zap.Int("post_id", req.PostID), zap.String("client_ip", clientIP), zap.String("device_info", deviceInfo), tracing.WithTraceIDField(ctx))

	// 检查用户状态
	if err := s.checkUserStatus(ctx, userID); err != nil {
		return nil, err
	}

	// 检查内容安全
	if err := s.checkContentSafety(ctx, req.Content); err != nil {
		return nil, err
	}

	// 检查帖子是否存在且状态正常
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(req.PostID), post.StatusEQ(post.StatusNormal)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("帖子不存在或已删除")
		}
		s.logger.Error("获取帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子失败: %w", err)
	}

	// 检查是否被楼主拉黑
	blacklistService := NewBlacklistService(s.db, s.logger)
	isBlockedByAuthor, err := blacklistService.IsUserBlocked(ctx, postData.UserID, userID)
	if err != nil {
		s.logger.Error("检查是否被楼主拉黑失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查拉黑状态失败: %w", err)
	}
	if isBlockedByAuthor {
		return nil, errors.New("您已被楼主拉黑，无法评论该帖子")
	}

	// 如果是回复评论，检查父评论是否存在
	var replyToUsername string
	if req.ParentID != nil {
		parentComment, err := s.db.Comment.Query().
			Where(comment.IDEQ(*req.ParentID), comment.PostIDEQ(req.PostID)).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, errors.New("父评论不存在或不属于该帖子")
			}
			s.logger.Error("获取父评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("获取父评论失败: %w", err)
		}
		// 设置回复目标用户ID
		req.ReplyToUserID = &parentComment.UserID
	}

	// 如果指定了回复目标用户，获取用户名并检查拉黑状态
	if req.ReplyToUserID != nil {
		// 检查是否被回复目标用户拉黑
		isBlockedByTarget, err := blacklistService.IsUserBlocked(ctx, *req.ReplyToUserID, userID)
		if err != nil {
			s.logger.Error("检查是否被回复目标用户拉黑失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("检查拉黑状态失败: %w", err)
		}
		if isBlockedByTarget {
			return nil, errors.New("您已被该用户拉黑，无法回复其评论")
		}

		replyToUser, err := s.db.User.Query().
			Where(user.IDEQ(*req.ReplyToUserID)).
			Only(ctx)
		if err != nil {
			s.logger.Error("获取回复目标用户失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			return nil, fmt.Errorf("获取回复目标用户失败: %w", err)
		}
		replyToUsername = replyToUser.Username
	}

	// 创建评论
	newComment, err := s.db.Comment.Create().
		SetUserID(userID).
		SetPostID(req.PostID).
		SetContent(req.Content).
		SetNillableParentID(req.ParentID).
		SetNillableReplyToUserID(req.ReplyToUserID).
		SetCommenterIP(clientIP).  // 从请求中获取真实IP
		SetDeviceInfo(deviceInfo). // 从请求中获取设备信息
		Save(ctx)
	if err != nil {
		s.logger.Error("创建评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}

	// 构建响应数据
	result := &schema.UserCommentCreateResponse{
		ID:              newComment.ID,
		PostID:          newComment.PostID,
		ReplyToUsername: replyToUsername,
		Content:         newComment.Content,
		LikeCount:       newComment.LikeCount,
		DislikeCount:    newComment.DislikeCount,
		CreatedAt:       newComment.CreatedAt.Format(time_tools.DateTimeFormat),
	}

	// 只有当ParentID不为0时才设置
	if newComment.ParentID != 0 {
		result.ParentID = &newComment.ParentID
	}

	// 只有当ReplyToUserID不为0时才设置
	if newComment.ReplyToUserID != 0 {
		result.ReplyToUserID = &newComment.ReplyToUserID
	}

	s.logger.Info("创建评论成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// UpdateComment 更新评论
func (s *CommentService) UpdateComment(ctx context.Context, userID int, req schema.UserCommentUpdateRequest) (*schema.UserCommentUpdateResponse, error) {
	s.logger.Info("更新评论", zap.Int("user_id", userID), zap.Int("comment_id", req.ID), tracing.WithTraceIDField(ctx))

	// 检查用户状态
	if err := s.checkUserStatus(ctx, userID); err != nil {
		return nil, err
	}

	// 检查内容安全
	if err := s.checkContentSafety(ctx, req.Content); err != nil {
		return nil, err
	}

	// 检查评论是否存在且属于当前用户
	commentData, err := s.db.Comment.Query().
		Where(comment.IDEQ(req.ID), comment.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.New("评论不存在或无权限修改")
		}
		s.logger.Error("获取评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论失败: %w", err)
	}

	// 获取评论所属帖子信息
	postData, err := s.db.Post.Query().
		Where(post.IDEQ(commentData.PostID)).
		Only(ctx)
	if err != nil {
		s.logger.Error("获取评论所属帖子失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取帖子信息失败: %w", err)
	}

	// 检查是否被楼主拉黑
	blacklistService := NewBlacklistService(s.db, s.logger)
	isBlockedByAuthor, err := blacklistService.IsUserBlocked(ctx, postData.UserID, userID)
	if err != nil {
		s.logger.Error("检查是否被楼主拉黑失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("检查拉黑状态失败: %w", err)
	}
	if isBlockedByAuthor {
		return nil, errors.New("您已被楼主拉黑，无法编辑该评论")
	}

	// 更新评论
	updatedComment, err := s.db.Comment.UpdateOne(commentData).
		SetContent(req.Content).
		Save(ctx)
	if err != nil {
		s.logger.Error("更新评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("更新评论失败: %w", err)
	}

	// 构建响应数据
	result := &schema.UserCommentUpdateResponse{
		ID:        updatedComment.ID,
		Content:   updatedComment.Content,
		UpdatedAt: updatedComment.UpdatedAt.Format(time_tools.DateTimeFormat),
	}

	s.logger.Info("更新评论成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// LikeComment 点赞评论
func (s *CommentService) LikeComment(ctx context.Context, userID int, req schema.UserCommentActionRequest) (*schema.UserCommentActionResponse, error) {
	s.logger.Info("点赞评论", zap.Int("user_id", userID), zap.Int("comment_id", req.ID), tracing.WithTraceIDField(ctx))

	// 使用统计服务执行点赞操作
	action, err := s.commentStatsService.PerformAction(ctx, userID, req.ID, "Like")
	if err != nil {
		s.logger.Error("点赞评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 构建响应数据
	result := &schema.UserCommentActionResponse{
		ID:           action.ID,
		LikeCount:    action.LikeCount,
		DislikeCount: action.DislikeCount,
	}

	s.logger.Info("点赞评论成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// DislikeComment 点踩评论
func (s *CommentService) DislikeComment(ctx context.Context, userID int, req schema.UserCommentActionRequest) (*schema.UserCommentActionResponse, error) {
	s.logger.Info("点踩评论", zap.Int("user_id", userID), zap.Int("comment_id", req.ID), tracing.WithTraceIDField(ctx))

	// 使用统计服务执行点踩操作
	action, err := s.commentStatsService.PerformAction(ctx, userID, req.ID, "Dislike")
	if err != nil {
		s.logger.Error("点踩评论失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, err
	}

	// 构建响应数据
	result := &schema.UserCommentActionResponse{
		ID:           action.ID,
		LikeCount:    action.LikeCount,
		DislikeCount: action.DislikeCount,
	}

	s.logger.Info("点踩评论成功", zap.Int("comment_id", result.ID), tracing.WithTraceIDField(ctx))
	return result, nil
}

// GetCommentList 获取评论列表
func (s *CommentService) GetCommentList(ctx context.Context, req schema.UserCommentListRequest) (*schema.UserCommentListResponse, error) {
	s.logger.Info("获取评论列表", zap.Int("post_id", req.PostID), zap.Int("page", req.Page), tracing.WithTraceIDField(ctx))

	// 构建查询条件，固定按创建时间升序排序（盖楼形式：最早的在最前面）
	query := s.db.Comment.Query().
		Where(comment.PostIDEQ(req.PostID)).
		Order(ent.Asc(comment.FieldCreatedAt))

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		s.logger.Error("获取评论总数失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论总数失败: %w", err)
	}

	// 分页查询
	comments, err := query.
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		All(ctx)
	if err != nil {
		s.logger.Error("获取评论列表失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("获取评论列表失败: %w", err)
	}

	// 收集所有需要查询的用户ID
	userIDs := make(map[int]bool)
	for _, c := range comments {
		userIDs[c.UserID] = true
		if c.ReplyToUserID != 0 {
			userIDs[c.ReplyToUserID] = true
		}
	}

	// 批量查询用户信息
	userIDList := make([]int, 0, len(userIDs))
	for id := range userIDs {
		userIDList = append(userIDList, id)
	}
	users, err := s.db.User.Query().
		Where(user.IDIn(userIDList...)).
		Select(user.FieldID, user.FieldUsername).
		All(ctx)
	if err != nil {
		s.logger.Error("查询用户信息失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}

	// 获取当前用户ID，如果未登录则为0
	currentUserID := tracing.GetUserID(ctx)

	// 批量查询用户点赞状态（仅当用户已登录时）
	var userLikeStatus map[int]map[string]bool // commentID -> {like: bool, dislike: bool}
	if currentUserID != 0 {
		// 获取评论ID列表
		commentIDs := make([]int, len(comments))
		for i, c := range comments {
			commentIDs[i] = c.ID
		}

		// 批量查询用户对这些评论的点赞记录
		actions, err := s.db.CommentAction.Query().
			Where(
				commentaction.UserIDEQ(currentUserID),
				commentaction.CommentIDIn(commentIDs...),
			).
			Select(commentaction.FieldCommentID, commentaction.FieldActionType).
			All(ctx)
		if err != nil {
			s.logger.Warn("查询用户点赞状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
			// 失败时不阻断流程，使用默认状态
			userLikeStatus = make(map[int]map[string]bool)
		} else {
			// 构建点赞状态映射
			userLikeStatus = make(map[int]map[string]bool)
			for _, action := range actions {
				if _, exists := userLikeStatus[action.CommentID]; !exists {
					userLikeStatus[action.CommentID] = map[string]bool{"like": false, "dislike": false}
				}
				if action.ActionType == commentaction.ActionTypeLike {
					userLikeStatus[action.CommentID]["like"] = true
				} else if action.ActionType == commentaction.ActionTypeDislike {
					userLikeStatus[action.CommentID]["dislike"] = true
				}
			}
		}
	} else {
		// 未登录用户，所有状态为false
		userLikeStatus = make(map[int]map[string]bool)
	}

	// 获取评论ID列表
	commentIDs := make([]int, len(comments))
	for i, c := range comments {
		commentIDs[i] = c.ID
	}

	// 批量获取实时统计数据
	statsMap, err := s.commentStatsService.GetStatsMap(ctx, commentIDs)
	if err != nil {
		s.logger.Warn("获取实时统计数据失败，将使用数据库中的旧数据", zap.Error(err), tracing.WithTraceIDField(ctx))
		// 失败时不阻断流程，降级使用数据库数据
		statsMap = make(map[int]*stats.Stats)
	}

	// 创建用户ID到用户名的映射
	userMap := make(map[int]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// 计算起始楼号
	startFloorNumber := (req.Page-1)*req.PageSize + 1

	// 构建响应数据
	list := make([]schema.UserCommentListItem, len(comments))
	for i, commentData := range comments {
		username := "未知用户"
		if name, ok := userMap[commentData.UserID]; ok {
			username = name
		}

		replyToUsername := ""
		if commentData.ReplyToUserID != 0 {
			if name, ok := userMap[commentData.ReplyToUserID]; ok {
				replyToUsername = name
			}
		}

		// 优先使用实时统计数据
		likeCount := commentData.LikeCount
		dislikeCount := commentData.DislikeCount
		if statsData, ok := statsMap[commentData.ID]; ok {
			likeCount = statsData.LikeCount
			dislikeCount = statsData.DislikeCount
		}

		// 获取用户点赞状态
		userLiked := false
		userDisliked := false
		if status, exists := userLikeStatus[commentData.ID]; exists {
			userLiked = status["like"]
			userDisliked = status["dislike"]
		}

		list[i] = schema.UserCommentListItem{
			ID:           commentData.ID,
			FloorNumber:  startFloorNumber + i, // 计算楼号
			UserID:       commentData.UserID,
			Username:     username,
			Content:      commentData.Content,
			LikeCount:    likeCount,
			DislikeCount: dislikeCount,
			UserLiked:    userLiked,
			UserDisliked: userDisliked,
			IsSelected:   commentData.IsSelected,
			IsPinned:     commentData.IsPinned,
			CreatedAt:    commentData.CreatedAt.Format(time_tools.DateTimeFormat),
			UpdatedAt:    commentData.UpdatedAt.Format(time_tools.DateTimeFormat),
		}

		// 只有当ParentID不为0时才设置
		if commentData.ParentID != 0 {
			list[i].ParentID = &commentData.ParentID
		}

		// 只有当ReplyToUserID不为0时才设置
		if commentData.ReplyToUserID != 0 {
			list[i].ReplyToUserID = &commentData.ReplyToUserID
			list[i].ReplyToUsername = replyToUsername
		}
	}

	result := &schema.UserCommentListResponse{
		List:     list,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	s.logger.Info("获取评论列表成功", zap.Int("total", total), tracing.WithTraceIDField(ctx))
	return result, nil
}

// checkUserStatus 检查用户状态是否允许操作
func (s *CommentService) checkUserStatus(ctx context.Context, userID int) error {
	userData, err := s.db.User.Query().
		Where(user.IDEQ(userID)).
		Select(user.FieldStatus, user.FieldEmailVerified).
		Only(ctx)
	if err != nil {
		s.logger.Error("获取用户状态失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("获取用户状态失败: %w", err)
	}

	switch userData.Status {
	case user.StatusNormal:
		// 检查是否需要验证邮箱
		verifyEmail, _ := s.settingsService.GetSettingByKey(ctx, _const.SafeVerifyEmail, "false")
		if verifyEmail == _const.SettingBoolTrue.String() && !userData.EmailVerified {
			return errors.New("您的邮箱尚未验证，请先完成验证")
		}
		return nil
	case user.StatusRiskControl:
		// TODO: RiskControl状态需要管理员审核发布，暂时放行
		return nil
	case user.StatusMute:
		return errors.New("您已被禁言，无法进行此操作")
	case user.StatusBlocked:
		return errors.New("您的账号已被封禁，无法进行此操作")
	default:
		return errors.New("账号状态异常，无法进行此操作")
	}
}

// checkContentSafety 检查评论内容是否安全
func (s *CommentService) checkContentSafety(ctx context.Context, content string) error {
	// 获取评论设置
	settings, err := s.settingsService.GetCommentSettings(ctx)
	if err != nil {
		s.logger.Error("获取评论设置失败", zap.Error(err), tracing.WithTraceIDField(ctx))
		return fmt.Errorf("检查内容安全失败: %w", err)
	}

	// 如果未开启审核，直接通过
	if !settings.RequireApproval {
		return nil
	}

	// 检查关键词黑名单
	if settings.KeywordBlacklist != "" {
		keywords := strings.Split(settings.KeywordBlacklist, ",")
		for _, keyword := range keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				continue
			}
			if strings.Contains(content, keyword) {
				return errors.New("评论内容包含违禁词")
			}
		}
	}

	return nil
}
