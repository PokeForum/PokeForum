package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ReportType Report type enumeration | 举报类型枚举
type ReportType string

const (
	// ReportTypeSpam Spam | 垃圾信息
	ReportTypeSpam ReportType = "spam"
	// ReportTypeHarassment Harassment | 人身攻击/骚扰
	ReportTypeHarassment ReportType = "harassment"
	// ReportTypeInappropriate Inappropriate content | 不当内容
	ReportTypeInappropriate ReportType = "inappropriate"
	// ReportTypeIllegal Illegal content | 违法违规
	ReportTypeIllegal ReportType = "illegal"
	// ReportTypePlagiarism Plagiarism/Infringement | 抄袭/侵权
	ReportTypePlagiarism ReportType = "plagiarism"
	// ReportTypeOther Other | 其他
	ReportTypeOther ReportType = "other"
)

// ReportStatus Report status enumeration | 举报状态枚举
type ReportStatus string

const (
	// ReportStatusPending Pending | 待处理
	ReportStatusPending ReportStatus = "pending"
	// ReportStatusProcessing Processing | 处理中
	ReportStatusProcessing ReportStatus = "processing"
	// ReportStatusResolved Resolved | 已处理
	ReportStatusResolved ReportStatus = "resolved"
	// ReportStatusIgnored Ignored | 已忽略
	ReportStatusIgnored ReportStatus = "ignored"
)

// ReportTargetType Report target type enumeration | 举报目标类型枚举
type ReportTargetType string

const (
	// ReportTargetPost Post | 帖子
	ReportTargetPost ReportTargetType = "post"
	// ReportTargetComment Comment | 评论
	ReportTargetComment ReportTargetType = "comment"
	// ReportTargetUser User | 用户
	ReportTargetUser ReportTargetType = "user"
)

// Report holds the schema definition for the Report entity.
// Report 举报实体
type Report struct {
	ent.Schema
}

// Fields of the Report.
func (Report) Fields() []ent.Field {
	return []ent.Field{
		// Reporter user ID | 举报者用户ID
		field.Int("reporter_id").
			Positive().
			Comment("Reporter user ID | 举报者用户ID"),
		// Target ID (post_id/comment_id/user_id) | 被举报目标ID
		field.Int("target_id").
			Positive().
			Comment("Target ID (post_id/comment_id/user_id) | 被举报目标ID"),
		// Target type: post/comment/user | 目标类型
		field.String("target_type").
			GoType(ReportTargetType("")).
			Default(string(ReportTargetPost)).
			Comment("Target type: post/comment/user | 目标类型：帖子/评论/用户"),
		// Target author user ID | 被举报目标作者ID
		field.Int("target_author_id").
			Positive().
			Comment("Target author user ID | 被举报目标作者ID"),
		// Report type | 举报类型
		field.String("report_type").
			GoType(ReportType("")).
			Default(string(ReportTypeOther)).
			Comment("Report type | 举报类型"),
		// Detailed reason (supports Markdown) | 详细原因（支持 Markdown）
		field.String("reason").
			MaxLen(5000).
			Optional().
			Comment("Detailed reason (supports Markdown for evidence) | 详细原因（支持 Markdown 格式添加证据）"),
		// Report status | 举报状态
		field.String("status").
			GoType(ReportStatus("")).
			Default(string(ReportStatusPending)).
			Comment("Report status: pending/processing/resolved/ignored | 举报状态"),
		// Handler (moderator/admin) ID | 处理人ID
		field.Int("handler_id").
			Positive().
			Optional().
			Nillable().
			Comment("Handler (moderator/admin) ID | 处理人ID"),
		// Handle result description | 处理结果说明
		field.String("handle_result").
			MaxLen(500).
			Optional().
			Comment("Handle result description | 处理结果说明"),
		// Handle action: delete/warn/ban/ignore | 处理动作
		field.String("handle_action").
			MaxLen(50).
			Optional().
			Comment("Handle action: delete/warn/ban/ignore | 处理动作"),
		// Handled time | 处理时间
		field.Time("handled_at").
			Optional().
			Nillable().
			Comment("Handled time | 处理时间"),
	}
}

// Indexes of the Report.
func (Report) Indexes() []ent.Index {
	return []ent.Index{
		// Query pending reports | 查询待处理举报
		index.Fields("status"),
		// Query user's report on a target (prevent duplicate) | 查询用户对某目标的举报（防重复）
		index.Fields("reporter_id", "target_id", "target_type").Unique(),
		// Query all reports for a target | 查询某目标的所有举报
		index.Fields("target_id", "target_type"),
		// Query reports for a specific author | 查询某作者被举报的记录
		index.Fields("target_author_id"),
		// Query reports by handler | 查询处理人的举报
		index.Fields("handler_id", "status"),
		// Composite index: status + created_at (for pagination) | 复合索引：状态+创建时间（分页查询）
		index.Fields("status", "created_at"),
	}
}

// Mixin of the Report.
func (Report) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
