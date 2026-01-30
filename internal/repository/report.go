package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/report"
	"github.com/PokeForum/PokeForum/ent/schema"
)

// IReportRepository Report repository interface | 举报仓储接口
type IReportRepository interface {
	// Create creates a new report | 创建举报
	Create(ctx context.Context, reporterID, targetID, targetAuthorID int, targetType,
		reportType, reason string) (*ent.Report, error)
	// GetByID gets report by ID | 根据ID获取举报
	GetByID(ctx context.Context, id int) (*ent.Report, error)
	// GetByReporterAndTarget checks if user already reported target | 检查用户是否已举报该目标
	GetByReporterAndTarget(ctx context.Context, reporterID, targetID int, targetType string) (*ent.Report, error)
	// ListByReporter gets reports by reporter with pagination | 获取用户的举报列表
	ListByReporter(ctx context.Context, reporterID int, page, pageSize int) ([]*ent.Report, int, error)
	// ListWithFilter gets reports with filters (for admin) | 带筛选的举报列表（管理用）
	ListWithFilter(ctx context.Context, filter ReportListFilter) ([]*ent.Report, int, error)
	// UpdateStatus updates report status | 更新举报状态
	UpdateStatus(ctx context.Context, id int, status string, handlerID int,
		handleResult, handleAction string, handledAt *time.Time) (*ent.Report, error)
	// BatchUpdateStatus batch updates report status | 批量更新举报状态
	BatchUpdateStatus(ctx context.Context, ids []int, status string, handlerID int,
		handleResult, handleAction string, handledAt *time.Time) (int, error)
	// GetStats gets report statistics | 获取举报统计
	GetStats(ctx context.Context) (*ReportStats, error)
	// GetPendingCount gets pending reports count | 获取待处理举报数量
	GetPendingCount(ctx context.Context) (int, error)
	// GetTargetReports gets all reports for a target | 获取目标的所有举报
	GetTargetReports(ctx context.Context, targetID int, targetType string) ([]*ent.Report, error)
}

// ReportRepository Report repository implementation | 举报仓储实现
type ReportRepository struct {
	db *ent.Client
}

// ReportListFilter Filter for report list query | 举报列表查询过滤器
type ReportListFilter struct {
	Page           int
	PageSize       int
	Status         string
	TargetType     string
	ReportType     string
	ReporterID     int
	TargetAuthorID int
}

// ReportStats Report statistics | 举报统计
type ReportStats struct {
	Total      int64
	Pending    int64
	Processing int64
	Resolved   int64
	Ignored    int64
	TodayCount int64
	ByType     map[string]int64
}

// NewReportRepository Create report repository instance | 创建举报仓储实例
func NewReportRepository(db *ent.Client) IReportRepository {
	return &ReportRepository{db: db}
}

// Create creates a new report | 创建举报
func (r *ReportRepository) Create(ctx context.Context, reporterID, targetID, targetAuthorID int,
	targetType, reportType, reason string) (*ent.Report, error) {
	builder := r.db.Report.Create().
		SetReporterID(reporterID).
		SetTargetID(targetID).
		SetTargetAuthorID(targetAuthorID).
		SetTargetType(schema.ReportTargetType(targetType)).
		SetReportType(schema.ReportType(reportType))

	if reason != "" {
		builder.SetReason(reason)
	}

	item, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create report failed | 创建举报失败: %w", err)
	}
	return item, nil
}

// GetByID gets report by ID | 根据ID获取举报
func (r *ReportRepository) GetByID(ctx context.Context, id int) (*ent.Report, error) {
	item, err := r.db.Report.Query().
		Where(report.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get report by id failed | 根据ID获取举报失败: %w", err)
	}
	return item, nil
}

// GetByReporterAndTarget checks if user already reported target | 检查用户是否已举报该目标
func (r *ReportRepository) GetByReporterAndTarget(ctx context.Context, reporterID, targetID int, targetType string) (*ent.Report, error) {
	item, err := r.db.Report.Query().
		Where(
			report.ReporterIDEQ(reporterID),
			report.TargetIDEQ(targetID),
			report.TargetTypeEQ(schema.ReportTargetType(targetType)),
		).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("get report by reporter and target failed | 查询举报失败: %w", err)
	}
	return item, nil
}

// ListByReporter gets reports by reporter with pagination | 获取用户的举报列表
func (r *ReportRepository) ListByReporter(ctx context.Context, reporterID int, page, pageSize int) ([]*ent.Report, int, error) {
	query := r.db.Report.Query().
		Where(report.ReporterIDEQ(reporterID)).
		Order(ent.Desc(report.FieldCreatedAt))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count reports failed | 统计举报数失败: %w", err)
	}

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	items, err := query.All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list reports failed | 查询举报列表失败: %w", err)
	}

	return items, total, nil
}

// ListWithFilter gets reports with filters (for admin) | 带筛选的举报列表（管理用）
func (r *ReportRepository) ListWithFilter(ctx context.Context, filter ReportListFilter) ([]*ent.Report, int, error) {
	query := r.db.Report.Query()

	// Apply filters | 应用过滤器
	if filter.Status != "" {
		query = query.Where(report.StatusEQ(schema.ReportStatus(filter.Status)))
	}
	if filter.TargetType != "" {
		query = query.Where(report.TargetTypeEQ(schema.ReportTargetType(filter.TargetType)))
	}
	if filter.ReportType != "" {
		query = query.Where(report.ReportTypeEQ(schema.ReportType(filter.ReportType)))
	}
	if filter.ReporterID > 0 {
		query = query.Where(report.ReporterIDEQ(filter.ReporterID))
	}
	if filter.TargetAuthorID > 0 {
		query = query.Where(report.TargetAuthorIDEQ(filter.TargetAuthorID))
	}

	query = query.Order(ent.Desc(report.FieldCreatedAt))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count reports failed | 统计举报数失败: %w", err)
	}

	page := filter.Page
	pageSize := filter.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	items, err := query.Offset(offset).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list reports failed | 查询举报列表失败: %w", err)
	}

	return items, total, nil
}

// UpdateStatus updates report status | 更新举报状态
func (r *ReportRepository) UpdateStatus(ctx context.Context, id int, status string, handlerID int,
	handleResult, handleAction string, handledAt *time.Time) (*ent.Report, error) {
	builder := r.db.Report.UpdateOneID(id).
		SetStatus(schema.ReportStatus(status))

	if handlerID > 0 {
		builder.SetHandlerID(handlerID)
	}
	if handleResult != "" {
		builder.SetHandleResult(handleResult)
	}
	if handleAction != "" {
		builder.SetHandleAction(handleAction)
	}
	if handledAt != nil {
		builder.SetHandledAt(*handledAt)
	}

	item, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update report status failed | 更新举报状态失败: %w", err)
	}
	return item, nil
}

// BatchUpdateStatus batch updates report status | 批量更新举报状态
func (r *ReportRepository) BatchUpdateStatus(ctx context.Context, ids []int, status string, handlerID int,
	handleResult, handleAction string, handledAt *time.Time) (int, error) {
	builder := r.db.Report.Update().
		Where(report.IDIn(ids...)).
		SetStatus(schema.ReportStatus(status))

	if handlerID > 0 {
		builder.SetHandlerID(handlerID)
	}
	if handleResult != "" {
		builder.SetHandleResult(handleResult)
	}
	if handleAction != "" {
		builder.SetHandleAction(handleAction)
	}
	if handledAt != nil {
		builder.SetHandledAt(*handledAt)
	}

	affected, err := builder.Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("batch update report status failed | 批量更新举报状态失败: %w", err)
	}
	return affected, nil
}

// GetStats gets report statistics | 获取举报统计
func (r *ReportRepository) GetStats(ctx context.Context) (*ReportStats, error) {
	stats := &ReportStats{
		ByType: make(map[string]int64),
	}

	// Total count | 总数
	total, err := r.db.Report.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("get total reports count failed | 获取总举报数失败: %w", err)
	}
	stats.Total = int64(total)

	// Pending count | 待处理数
	pending, err := r.db.Report.Query().Where(report.StatusEQ(schema.ReportStatusPending)).Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("get pending count failed | 获取待处理数失败: %w", err)
	}
	stats.Pending = int64(pending)

	// Processing count | 处理中数
	processing, err := r.db.Report.Query().Where(report.StatusEQ(schema.ReportStatusProcessing)).Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("get processing count failed | 获取处理中数失败: %w", err)
	}
	stats.Processing = int64(processing)

	// Resolved count | 已处理数
	resolved, err := r.db.Report.Query().Where(report.StatusEQ(schema.ReportStatusResolved)).Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("get resolved count failed | 获取已处理数失败: %w", err)
	}
	stats.Resolved = int64(resolved)

	// Ignored count | 已忽略数
	ignored, err := r.db.Report.Query().Where(report.StatusEQ(schema.ReportStatusIgnored)).Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("get ignored count failed | 获取已忽略数失败: %w", err)
	}
	stats.Ignored = int64(ignored)

	// Today's count | 今日数
	today := time.Now().Truncate(24 * time.Hour)
	todayCount, err := r.db.Report.Query().
		Where(report.CreatedAtGTE(today)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("get today count failed | 获取今日数失败: %w", err)
	}
	stats.TodayCount = int64(todayCount)

	// Count by report type | 按举报类型统计
	reportTypes := []schema.ReportType{
		schema.ReportTypeSpam,
		schema.ReportTypeHarassment,
		schema.ReportTypeInappropriate,
		schema.ReportTypeIllegal,
		schema.ReportTypePlagiarism,
		schema.ReportTypeOther,
	}
	for _, rt := range reportTypes {
		count, err := r.db.Report.Query().Where(report.ReportTypeEQ(rt)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("get count by type failed | 按类型统计失败: %w", err)
		}
		stats.ByType[string(rt)] = int64(count)
	}

	return stats, nil
}

// GetPendingCount gets pending reports count | 获取待处理举报数量
func (r *ReportRepository) GetPendingCount(ctx context.Context) (int, error) {
	count, err := r.db.Report.Query().
		Where(report.StatusEQ(schema.ReportStatusPending)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("get pending count failed | 获取待处理数失败: %w", err)
	}
	return count, nil
}

// GetTargetReports gets all reports for a target | 获取目标的所有举报
func (r *ReportRepository) GetTargetReports(ctx context.Context, targetID int, targetType string) ([]*ent.Report, error) {
	items, err := r.db.Report.Query().
		Where(
			report.TargetIDEQ(targetID),
			report.TargetTypeEQ(schema.ReportTargetType(targetType)),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get target reports failed | 获取目标举报失败: %w", err)
	}
	return items, nil
}
