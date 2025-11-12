package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserOAuth holds the schema definition for the UserOAuth entity.
// 存储用户与第三方OAuth账号的绑定关系
type UserOAuth struct {
	ent.Schema
}

// Fields of the UserOAuth.
func (UserOAuth) Fields() []ent.Field {
	return []ent.Field{
		// 绑定记录ID，数据库主键自增
		field.Int("id").
			Positive(),
		// 用户ID，关联users表
		field.Int("user_id").
			Positive(),
		// OAuth提供商：QQ、GitHub、Apple、Google、Telegram、FIDO2
		field.Enum("provider").
			Values("QQ", "GitHub", "Apple", "Google", "Telegram", "FIDO2"),
		// 第三方平台的用户唯一标识（OpenID、UnionID等）
		field.String("provider_user_id").
			NotEmpty(),
		// 第三方平台的用户名
		field.String("provider_username").
			Optional(),
		// 第三方平台的用户邮箱
		field.String("provider_email").
			Optional(),
		// 第三方平台的用户头像URL
		field.String("provider_avatar").
			Optional(),
		// 额外的用户信息，JSON格式存储
		field.JSON("extra_data", map[string]interface{}{}).
			Optional(),
	}
}

// Edges of the UserOAuth.
// 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// user_id字段关联users表，但不创建外键约束，关联逻辑在应用层维护
func (UserOAuth) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserOAuth.
func (UserOAuth) Indexes() []ent.Index {
	return []ent.Index{
		// 联合唯一索引：同一个提供商的同一个用户ID只能绑定一次
		index.Fields("provider", "provider_user_id").
			Unique(),
		// 用户ID索引，用于查询用户的所有OAuth绑定
		index.Fields("user_id"),
		// 提供商索引，用于按提供商查询
		index.Fields("provider"),
		// 联合索引：用户ID + 提供商，用于快速查询用户在特定平台的绑定
		index.Fields("user_id", "provider"),
	}
}

// Mixin of the UserOAuth.
func (UserOAuth) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
