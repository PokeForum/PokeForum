package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	entschema "github.com/PokeForum/PokeForum/ent/schema"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IReportService Report service interface (user-side) | 举报服务接口（用户侧）
type IReportService interface {
	// CreateReport creates a new report | 创建举报
	CreateReport(ctx context.Context, reporterID int, req *schema.ReportCreateRequest) (*schema.ReportCreateResponse, error)
	// GetUserReports gets current user's reports | 获取当前用户的举报列表
	GetUserReports(ctx context.Context, reporterID, page, pageSize int) (*schema.UserReportListResponse, error)
}

// ReportService Report service implementation (user-side) | 举报服务实现（用户侧）
type ReportService struct {
	reportRepo  repository.IReportRepository
	userRepo    repository.IUserRepository
	postRepo    repository.IPostRepository
	commentRepo repository.ICommentRepository
	logger      *zap.Logger
}

// NewReportService Create report service instance | 创建举报服务实例
func NewReportService(
	reportRepo repository.IReportRepository,
	userRepo repository.IUserRepository,
	postRepo repository.IPostRepository,
	commentRepo repository.ICommentRepository,
	logger *zap.Logger,
) IReportService {
	return &ReportService{
		reportRepo:  reportRepo,
		userRepo:    userRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		logger:      logger,
	}
}

// CreateReport creates a new report | 创建举报
func (s *ReportService) CreateReport(ctx context.Context, reporterID int, req *schema.ReportCreateRequest) (*schema.ReportCreateResponse, error) {
	// Cannot report yourself | 不能举报自己
	if reporterID == req.TargetID && req.TargetType == string(entschema.ReportTargetUser) {
		s.logger.Warn("cannot report yourself",
			tracing.WithTraceIDField(ctx),
			zap.Int("reporter_id", reporterID),
		)
		return nil, errors.New("cannot report yourself | 不能举报自己")
	}

	// Check if already reported | 检查是否已举报
	existing, err := s.reportRepo.GetByReporterAndTarget(ctx, reporterID, req.TargetID, req.TargetType)
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Error("failed to check existing report",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check existing report | 检查举报记录失败: %w", err)
	}
	if existing != nil {
		s.logger.Warn("already reported this target",
			tracing.WithTraceIDField(ctx),
			zap.Int("reporter_id", reporterID),
			zap.Int("target_id", req.TargetID),
		)
		return nil, errors.New("you have already reported this target | 您已举报过该目标")
	}

	// Get target author ID | 获取目标作者ID
	targetAuthorID, err := s.getTargetAuthorID(ctx, req.TargetID, req.TargetType)
	if err != nil {
		s.logger.Error("failed to get target author",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, err
	}

	// Cannot report yourself (for posts/comments) | 不能举报自己（针对帖子/评论）
	if targetAuthorID == reporterID {
		s.logger.Warn("cannot report your own content",
			tracing.WithTraceIDField(ctx),
			zap.Int("reporter_id", reporterID),
		)
		return nil, errors.New("cannot report your own content | 不能举报自己的内容")
	}

	// Create report | 创建举报
	reportItem, err := s.reportRepo.Create(ctx, reporterID, req.TargetID, targetAuthorID,
		req.TargetType, req.ReportType, req.Reason)
	if err != nil {
		s.logger.Error("failed to create report",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create report | 创建举报失败: %w", err)
	}

	s.logger.Info("report created successfully",
		tracing.WithTraceIDField(ctx),
		zap.Int("report_id", reportItem.ID),
		zap.Int("reporter_id", reporterID),
		zap.Int("target_id", req.TargetID),
	)

	return &schema.ReportCreateResponse{
		ID:        reportItem.ID,
		Status:    string(reportItem.Status),
		CreatedAt: reportItem.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetUserReports gets current user's reports | 获取当前用户的举报列表
func (s *ReportService) GetUserReports(ctx context.Context, reporterID, page, pageSize int) (*schema.UserReportListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	reports, total, err := s.reportRepo.ListByReporter(ctx, reporterID, page, pageSize)
	if err != nil {
		s.logger.Error("failed to get user reports",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get user reports | 获取举报列表失败: %w", err)
	}

	items := make([]schema.UserReportItem, 0, len(reports))
	for _, r := range reports {
		item := schema.UserReportItem{
			ID:         r.ID,
			TargetID:   r.TargetID,
			TargetType: string(r.TargetType),
			ReportType: string(r.ReportType),
			Reason:     r.Reason,
			Status:     string(r.Status),
			CreatedAt:  r.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// Get target title | 获取目标标题
		title, err := s.getTargetTitle(ctx, r.TargetID, r.TargetType)
		if err == nil {
			item.TargetTitle = title
		}

		if r.HandleResult != "" {
			item.HandleResult = r.HandleResult
		}
		if r.HandledAt != nil {
			item.HandledAt = r.HandledAt.Format("2006-01-02 15:04:05")
		}

		items = append(items, item)
	}

	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return &schema.UserReportListResponse{
		List:       items,
		Total:      int64(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// getTargetAuthorID gets the author ID of the target | 获取目标的作者ID
func (s *ReportService) getTargetAuthorID(ctx context.Context, targetID int, targetType string) (int, error) {
	switch targetType {
	case string(entschema.ReportTargetPost):
		post, err := s.postRepo.GetByID(ctx, targetID)
		if err != nil {
			if ent.IsNotFound(err) {
				return 0, errors.New("post not found | 帖子不存在")
			}
			return 0, fmt.Errorf("failed to get post | 获取帖子失败: %w", err)
		}
		return post.UserID, nil

	case string(entschema.ReportTargetComment):
		comment, err := s.commentRepo.GetByID(ctx, targetID)
		if err != nil {
			if ent.IsNotFound(err) {
				return 0, errors.New("comment not found | 评论不存在")
			}
			return 0, fmt.Errorf("failed to get comment | 获取评论失败: %w", err)
		}
		return comment.UserID, nil

	case string(entschema.ReportTargetUser):
		user, err := s.userRepo.GetByID(ctx, targetID)
		if err != nil {
			if ent.IsNotFound(err) {
				return 0, errors.New("user not found | 用户不存在")
			}
			return 0, fmt.Errorf("failed to get user | 获取用户失败: %w", err)
		}
		return user.ID, nil

	default:
		return 0, errors.New("invalid target type | 无效的目标类型")
	}
}

// getTargetTitle gets the title of the target | 获取目标的标题
func (s *ReportService) getTargetTitle(ctx context.Context, targetID int, targetType entschema.ReportTargetType) (string, error) {
	switch targetType {
	case entschema.ReportTargetPost:
		post, err := s.postRepo.GetByID(ctx, targetID)
		if err != nil {
			return "", err
		}
		return post.Title, nil

	case entschema.ReportTargetComment:
		comment, err := s.commentRepo.GetByID(ctx, targetID)
		if err != nil {
			return "", err
		}
		// Return preview of comment content | 返回评论内容摘要
		content := comment.Content
		if len(content) > 50 {
			content = content[:50] + "..."
		}
		return content, nil

	case entschema.ReportTargetUser:
		user, err := s.userRepo.GetByID(ctx, targetID)
		if err != nil {
			return "", err
		}
		return user.Username, nil

	default:
		return "", errors.New("invalid target type")
	}
}
