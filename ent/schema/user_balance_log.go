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
		// 主键ID
		field.Int("id").
			Positive(),
		// 用户ID，关联到用户表
		field.Int("user_id").
			Positive().
			Comment("用户ID，关联到用户表"),
		// 变动类型：points（积分）、currency（货币）、experience（经验值）
		field.Enum("type").
			Values("points", "currency", "experience").
			Comment("变动类型：points（积分）、currency（货币）、experience（经验值）"),
		// 变动数量：正数为增加，负数为减少
		field.Int("amount").
			Comment("变动数量：正数为增加，负数为减少"),
		// 变动前数量
		field.Int("before_amount").
			Comment("变动前数量"),
		// 变动后数量
		field.Int("after_amount").
			Comment("变动后数量"),
		// 变动原因/说明
		field.String("reason").
			NotEmpty().
			Comment("变动原因/说明"),
		// 操作者ID，可为空（系统操作时为空）
		field.Int("operator_id").
			Optional().
			Comment("操作者ID，可为空（系统操作时为空）"),
		// 操作者用户名
		field.String("operator_name").
			Optional().
			Comment("操作者用户名"),
		// 关联的业务ID（如帖子ID、订单ID等），可选
		field.Int("related_id").
			Optional().
			Comment("关联的业务ID（如帖子ID、订单ID等）"),
		// 关联的业务类型
		field.String("related_type").
			Optional().
			Comment("关联的业务类型"),
		// IP地址
		field.String("ip_address").
			Optional().
			Comment("IP地址"),
		// 用户代理
		field.String("user_agent").
			Optional().
			Comment("用户代理"),
	}
}

// Edges of the UserBalanceLog.
// 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// 数据完整性由应用层逻辑保证
func (UserBalanceLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserBalanceLog.
func (UserBalanceLog) Indexes() []ent.Index {
	return []ent.Index{
		// 为关联字段创建索引以优化查询性能
		index.Fields("user_id"),
		index.Fields("operator_id"),
		// 为常用查询字段创建索引
		index.Fields("type"),
		// 创建复合索引优化用户余额变动历史查询
		index.Fields("user_id", "type", "created_at"),
	}
}

// Mixin of the UserBalanceLog.
func (UserBalanceLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
