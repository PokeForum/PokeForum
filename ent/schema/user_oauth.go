package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserOAuth holds the schema definition for the UserOAuth entity.
// Store the binding relationship between users and third-party OAuth accounts | 存储用户与第三方OAuth账号的绑定关系
type UserOAuth struct {
	ent.Schema
}

// Fields of the UserOAuth.
func (UserOAuth) Fields() []ent.Field {
	return []ent.Field{
		// Binding record ID, database primary key auto-increment | 绑定记录ID，数据库主键自增
		field.Int("id").
			Positive(),
		// User ID, associated with users table | 用户ID，关联users表
		field.Int("user_id").
			Positive(),
		// OAuth provider: QQ, GitHub, Google, FIDO2 | OAuth提供商：QQ、GitHub、Google、FIDO2
		field.Enum("provider").
			Values("QQ", "GitHub", "Google", "FIDO2"),
		// Third-party platform user unique identifier (OpenID, UnionID, etc.) | 第三方平台的用户唯一标识（OpenID、UnionID等）
		field.String("provider_user_id").
			NotEmpty(),
		// Third-party platform username | 第三方平台的用户名
		field.String("provider_username").
			Optional(),
		// Third-party platform user email | 第三方平台的用户邮箱
		field.String("provider_email").
			Optional(),
		// Third-party platform user avatar URL | 第三方平台的用户头像URL
		field.String("provider_avatar").
			Optional(),
		// Extra user information, stored in JSON format | 额外的用户信息，JSON格式存储
		field.JSON("extra_data", map[string]interface{}{}).
			Optional(),
	}
}

// Edges of the UserOAuth.
// Note: All associations are only used for ORM queries and will not create foreign keys at the database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// user_id field is associated with users table, but does not create foreign key constraint, association logic is maintained at the application layer | user_id字段关联users表，但不创建外键约束，关联逻辑在应用层维护
func (UserOAuth) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserOAuth.
func (UserOAuth) Indexes() []ent.Index {
	return []ent.Index{
		// Composite unique index: the same user ID from the same provider can only be bound once | 联合唯一索引：同一个提供商的同一个用户ID只能绑定一次
		index.Fields("provider", "provider_user_id").
			Unique(),
		// User ID index, used to query all OAuth bindings of a user | 用户ID索引，用于查询用户的所有OAuth绑定
		index.Fields("user_id"),
		// Provider index, used to query by provider | 提供商索引，用于按提供商查询
		index.Fields("provider"),
		// Composite index: user ID + provider, used to quickly query user bindings on a specific platform | 联合索引：用户ID + 提供商，用于快速查询用户在特定平台的绑定
		index.Fields("user_id", "provider"),
	}
}

// Mixin of the UserOAuth.
func (UserOAuth) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
