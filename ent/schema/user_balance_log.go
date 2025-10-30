package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
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
		// 变动类型：points（积分）、currency（货币）
		field.Enum("type").
			Values("points", "currency").
			Comment("变动类型：points（积分）、currency（货币）"),
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
func (UserBalanceLog) Edges() []ent.Edge {
	return []ent.Edge{
		// 反向关联到用户表
		edge.From("user", User.Type).
			Ref("balance_logs").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Mixin of the UserBalanceLog.
func (UserBalanceLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
