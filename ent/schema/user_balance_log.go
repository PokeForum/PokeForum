package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserBalanceLog holds the schema definition for the UserBalanceLog entity.
type UserBalanceLog struct {
	ent.Schema
}

// Fields of the UserBalanceLog.
func (UserBalanceLog) Fields() []ent.Field {
	return []ent.Field{
		// User ID, reference to user table | 用户ID，关联到用户表
		field.Int("user_id").
			Positive().
			Comment("User ID, reference to user table | 用户ID，关联到用户表"),
		// Change type: points (points), currency (currency), experience (experience points) | 变动类型：points（积分）、currency（货币）、experience（经验值）
		field.Enum("type").
			Values("points", "currency", "experience").
			Comment("Change type: points (points), currency (currency), experience (experience points) | 变动类型：points（积分）、currency（货币）、experience（经验值）"),
		// Change amount: positive for increase, negative for decrease | 变动数量：正数为增加，负数为减少
		field.Int("amount").
			Comment("Change amount: positive for increase, negative for decrease | 变动数量：正数为增加，负数为减少"),
		// Amount before change | 变动前数量
		field.Int("before_amount").
			Comment("Amount before change | 变动前数量"),
		// Amount after change | 变动后数量
		field.Int("after_amount").
			Comment("Amount after change | 变动后数量"),
		// Reason/description for the change | 变动原因/说明
		field.String("reason").
			NotEmpty().
			Comment("Reason/description for the change | 变动原因/说明"),
		// Operator ID, nullable (null for system operations) | 操作者ID，可为空（系统操作时为空）
		field.Int("operator_id").
			Optional().
			Comment("Operator ID, nullable (null for system operations) | 操作者ID，可为空（系统操作时为空）"),
		// Operator username | 操作者用户名
		field.String("operator_name").
			Optional().
			Comment("Operator username | 操作者用户名"),
		// Related business ID (e.g., post ID, order ID, etc.), optional | 关联的业务ID（如帖子ID、订单ID等），可选
		field.Int("related_id").
			Optional().
			Comment("Related business ID (e.g., post ID, order ID, etc.), optional | 关联的业务ID（如帖子ID、订单ID等）"),
		// Related business type | 关联的业务类型
		field.String("related_type").
			Optional().
			Comment("Related business type | 关联的业务类型"),
		// IP address | IP地址
		field.String("ip_address").
			Optional().
			Comment("IP address | IP地址"),
		// User agent | 用户代理
		field.String("user_agent").
			Optional().
			Comment("User agent | 用户代理"),
	}
}

// Edges of the UserBalanceLog.
// Note: All relationships are for ORM queries only, no foreign keys will be created at database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Data integrity is ensured by application layer logic | 数据完整性由应用层逻辑保证
func (UserBalanceLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserBalanceLog.
func (UserBalanceLog) Indexes() []ent.Index {
	return []ent.Index{
		// Create indexes on relational fields to optimize query performance | 为关联字段创建索引以优化查询性能
		index.Fields("user_id"),
		index.Fields("operator_id"),
		// Create indexes on frequently queried fields | 为常用查询字段创建索引
		index.Fields("type"),
		// Create composite index to optimize user balance change history queries | 创建复合索引优化用户余额变动历史查询
		index.Fields("user_id", "type", "created_at"),
	}
}

// Mixin of the UserBalanceLog.
func (UserBalanceLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
