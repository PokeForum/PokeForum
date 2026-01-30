package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	entschema "github.com/PokeForum/PokeForum/ent/schema"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/repository"
	"github.com/PokeForum/PokeForum/internal/schema"
)

// IReportManageService Report management service interface (admin-side) | 举报管理服务接口（管理侧）
type IReportManageService interface {
	// GetReportList gets report list for admins | 获取举报列表（管理员用）
	GetReportList(ctx context.Context, query *schema.ReportListQuery) (*schema.ReportListResponse, error)
	// HandleReport handles a report | 处理举报
	HandleReport(ctx context.Context, handlerID int, req *schema.ReportHandleRequest) (*schema.ReportHandleResponse, error)
	// BatchHandleReports batch handles reports | 批量处理举报
	BatchHandleReports(ctx context.Context, handlerID int, req *schema.ReportBatchHandleRequest) error
	// GetReportStats gets report statistics | 获取举报统计
	GetReportStats(ctx context.Context) (*schema.ReportStatsResponse, error)
	// GetPendingReportCount gets pending report count | 获取待处理举报数量
	GetPendingReportCount(ctx context.Context) (int, error)
}

// ReportManageService Report management service implementation (admin-side) | 举报管理服务实现（管理侧）
type ReportManageService struct {
	reportRepo  repository.IReportRepository
	userRepo    repository.IUserRepository
	postRepo    repository.IPostRepository
	commentRepo repository.ICommentRepository
	logger      *zap.Logger
}

// NewReportManageService Create report management service instance | 创建举报管理服务实例
func NewReportManageService(
	reportRepo repository.IReportRepository,
	userRepo repository.IUserRepository,
	postRepo repository.IPostRepository,
	commentRepo repository.ICommentRepository,
	logger *zap.Logger,
) IReportManageService {
	return &ReportManageService{
		reportRepo:  reportRepo,
		userRepo:    userRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		logger:      logger,
	}
}

// GetReportList gets report list for admins | 获取举报列表（管理员用）
func (s *ReportManageService) GetReportList(ctx context.Context, query *schema.ReportListQuery) (*schema.ReportListResponse, error) {
	filter := repository.ReportListFilter{
		Page:           query.Page,
		PageSize:       query.PageSize,
		Status:         query.Status,
		TargetType:     query.TargetType,
		ReportType:     query.ReportType,
		ReporterID:     query.ReporterID,
		TargetAuthorID: query.TargetAuthorID,
	}

	reports, total, err := s.reportRepo.ListWithFilter(ctx, filter)
	if err != nil {
		s.logger.Error("failed to get report list",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get report list | 获取举报列表失败: %w", err)
	}

	items := make([]schema.ReportListItem, 0, len(reports))
	for _, r := range reports {
		item := schema.ReportListItem{
			ID:             r.ID,
			ReporterID:     r.ReporterID,
			TargetID:       r.TargetID,
			TargetType:     string(r.TargetType),
			TargetAuthorID: r.TargetAuthorID,
			ReportType:     string(r.ReportType),
			Reason:         r.Reason,
			Status:         string(r.Status),
			HandleResult:   r.HandleResult,
			HandleAction:   r.HandleAction,
			CreatedAt:      r.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// Get reporter name | 获取举报者名称
		reporter, err := s.userRepo.GetByID(ctx, r.ReporterID)
		if err == nil && reporter != nil {
			item.ReporterName = reporter.Username
		}

		// Get target author name | 获取目标作者名称
		author, err := s.userRepo.GetByID(ctx, r.TargetAuthorID)
		if err == nil && author != nil {
			item.TargetAuthorName = author.Username
		}

		// Get target title and content | 获取目标标题和内容
		title, content, err := s.getTargetInfo(ctx, r.TargetID, r.TargetType)
		if err == nil {
			item.TargetTitle = title
			item.TargetContent = content
		}

		if r.HandlerID != nil && *r.HandlerID > 0 {
			item.HandlerID = *r.HandlerID
			handler, err := s.userRepo.GetByID(ctx, *r.HandlerID)
			if err == nil && handler != nil {
				item.HandlerName = handler.Username
			}
		}

		if r.HandledAt != nil {
			item.HandledAt = r.HandledAt.Format("2006-01-02 15:04:05")
		}

		items = append(items, item)
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return &schema.ReportListResponse{
		List:       items,
		Total:      int64(total),
		Page:       query.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// HandleReport handles a report | 处理举报
func (s *ReportManageService) HandleReport(ctx context.Context, handlerID int, req *schema.ReportHandleRequest) (*schema.ReportHandleResponse, error) {
	// Get report | 获取举报
	reportItem, err := s.reportRepo.GetByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("failed to get report",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get report | 获取举报失败: %w", err)
	}

	// Check status | 检查状态
	if reportItem.Status == entschema.ReportStatusResolved || reportItem.Status == entschema.ReportStatusIgnored {
		return nil, errors.New("report already handled | 举报已被处理")
	}

	// Determine status based on action | 根据动作确定状态
	status := string(entschema.ReportStatusResolved)
	if req.Action == "ignore" {
		status = string(entschema.ReportStatusIgnored)
	}

	now := time.Now()
	updatedReport, err := s.reportRepo.UpdateStatus(ctx, req.ID, status, handlerID,
		req.HandleResult, req.Action, &now)
	if err != nil {
		s.logger.Error("failed to update report status",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update report status | 更新举报状态失败: %w", err)
	}

	// Handle related reports if action is delete/warn/ban | 如果动作是删除/警告/封禁，处理相关举报
	if req.Action == "delete" || req.Action == "ban" {
		relatedReports, err := s.reportRepo.GetTargetReports(ctx, reportItem.TargetID, string(reportItem.TargetType))
		if err == nil && len(relatedReports) > 0 {
			relatedIDs := make([]int, 0, len(relatedReports))
			for _, r := range relatedReports {
				if r.ID != req.ID && r.Status != entschema.ReportStatusResolved && r.Status != entschema.ReportStatusIgnored {
					relatedIDs = append(relatedIDs, r.ID)
				}
			}
			if len(relatedIDs) > 0 {
				_, err := s.reportRepo.BatchUpdateStatus(ctx, relatedIDs, status, handlerID,
					"Handled together with related report", req.Action, &now)
				if err != nil {
					s.logger.Warn("failed to update related reports",
						tracing.WithTraceIDField(ctx),
						zap.Error(err),
						zap.Ints("related_ids", relatedIDs),
					)
				}
			}
		}
	}

	s.logger.Info("report handled successfully",
		tracing.WithTraceIDField(ctx),
		zap.Int("report_id", req.ID),
		zap.Int("handler_id", handlerID),
		zap.String("action", req.Action),
	)

	return &schema.ReportHandleResponse{
		ID:        updatedReport.ID,
		Status:    string(updatedReport.Status),
		HandledAt: now.Format("2006-01-02 15:04:05"),
	}, nil
}

// BatchHandleReports batch handles reports | 批量处理举报
func (s *ReportManageService) BatchHandleReports(ctx context.Context, handlerID int, req *schema.ReportBatchHandleRequest) error {
	status := string(entschema.ReportStatusResolved)
	if req.Action == "ignore" {
		status = string(entschema.ReportStatusIgnored)
	}

	now := time.Now()
	_, err := s.reportRepo.BatchUpdateStatus(ctx, req.IDs, status, handlerID,
		req.HandleResult, req.Action, &now)
	if err != nil {
		s.logger.Error("failed to batch update reports",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return fmt.Errorf("failed to batch update reports | 批量更新举报失败: %w", err)
	}

	s.logger.Info("reports batch handled successfully",
		tracing.WithTraceIDField(ctx),
		zap.Ints("report_ids", req.IDs),
		zap.Int("handler_id", handlerID),
		zap.String("action", req.Action),
	)

	return nil
}

// GetReportStats gets report statistics | 获取举报统计
func (s *ReportManageService) GetReportStats(ctx context.Context) (*schema.ReportStatsResponse, error) {
	stats, err := s.reportRepo.GetStats(ctx)
	if err != nil {
		s.logger.Error("failed to get report stats",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get report stats | 获取举报统计失败: %w", err)
	}

	byTypeStats := make([]schema.ReportStatsItem, 0, len(stats.ByType))
	for t, count := range stats.ByType {
		byTypeStats = append(byTypeStats, schema.ReportStatsItem{
			Type:  t,
			Count: count,
		})
	}

	return &schema.ReportStatsResponse{
		TotalReports:    stats.Total,
		PendingCount:    stats.Pending,
		ProcessingCount: stats.Processing,
		ResolvedCount:   stats.Resolved,
		IgnoredCount:    stats.Ignored,
		TodayCount:      stats.TodayCount,
		ByTypeStats:     byTypeStats,
	}, nil
}

// GetPendingReportCount gets pending report count | 获取待处理举报数量
func (s *ReportManageService) GetPendingReportCount(ctx context.Context) (int, error) {
	count, err := s.reportRepo.GetPendingCount(ctx)
	if err != nil {
		s.logger.Error("failed to get pending count",
			tracing.WithTraceIDField(ctx),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to get pending count | 获取待处理数失败: %w", err)
	}
	return count, nil
}

// getTargetInfo gets the title and content of the target | 获取目标的标题和内容
func (s *ReportManageService) getTargetInfo(ctx context.Context, targetID int, targetType entschema.ReportTargetType) (string, string, error) {
	switch targetType {
	case entschema.ReportTargetPost:
		post, err := s.postRepo.GetByID(ctx, targetID)
		if err != nil {
			return "", "", err
		}
		content := post.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		return post.Title, content, nil

	case entschema.ReportTargetComment:
		comment, err := s.commentRepo.GetByID(ctx, targetID)
		if err != nil {
			return "", "", err
		}
		content := comment.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		return "Comment", content, nil

	case entschema.ReportTargetUser:
		user, err := s.userRepo.GetByID(ctx, targetID)
		if err != nil {
			return "", "", err
		}
		return user.Username, user.Signature, nil

	default:
		return "", "", errors.New("invalid target type")
	}
}
