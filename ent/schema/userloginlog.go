package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// UserLoginLog holds the schema definition for the UserLoginLog entity.
type UserLoginLog struct {
	ent.Schema
}

// Fields of the UserLoginLog.
func (UserLoginLog) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID
		field.Int("id").
			Positive(),
		// 用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// IP地址
		field.String("ip_address").
			NotEmpty(),
		// IP国家
		field.String("ip_country").
			Optional(),
		// IP城市
		field.String("ip_city").
			Optional(),
		// 是否成功登录
		field.Bool("success").
			Default(true),
		// 设备信息，从浏览器UA中读取
		field.String("device_info").
			Optional(),
	}
}

// Edges of the UserLoginLog.
func (UserLoginLog) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联到User表
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
	}
}

// Mixin of the UserLoginLog.
func (UserLoginLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
