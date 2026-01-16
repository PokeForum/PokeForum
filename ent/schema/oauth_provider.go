package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OAuthProvider holds the schema definition for the OAuthProvider entity.
// Store configuration information for third-party OAuth login providers | 存储第三方OAuth登录提供商的配置信息
type OAuthProvider struct {
	ent.Schema
}

// Fields of the OAuthProvider.
func (OAuthProvider) Fields() []ent.Field {
	return []ent.Field{
		// Provider ID, database primary key auto-increment | 提供商ID，数据库主键自增
		field.Int("id").
			Positive(),
		// Provider name: QQ, GitHub, Google, FIDO2 | 提供商名称：QQ、GitHub、Google、FIDO2
		field.Enum("provider").
			Values("QQ", "GitHub", "Google", "FIDO2"),
		// Client ID | 客户端ID（Client ID）
		field.String("client_id").
			Optional(),
		// Client Secret, sensitive information | 客户端密钥（Client Secret），敏感信息
		field.String("client_secret").
			Optional().
			Sensitive(),
		// Authorization URL | 授权URL
		field.String("auth_url").
			Optional(),
		// Token URL | Token获取URL
		field.String("token_url").
			Optional(),
		// User info URL | 用户信息获取URL
		field.String("user_info_url").
			Optional(),
		// Request scopes, stored in JSON array format | 请求范围（Scopes），JSON数组格式存储
		field.JSON("scopes", []string{}).
			Optional(),
		// Extra configuration parameters, stored in JSON format, used for custom configuration of special providers | 额外配置参数，JSON格式存储，用于特殊提供商的自定义配置
		field.JSON("extra_config", map[string]interface{}{}).
			Optional(),
		// Whether to enable this provider | 是否启用该提供商
		field.Bool("enabled").
			Default(false),
		// Sort order, used for frontend display sorting | 排序顺序，用于前端展示排序
		field.Int("sort_order").
			Default(0).
			NonNegative(),
	}
}

// Edges of the OAuthProvider.
// Note: All associations are only used for ORM queries and will not create foreign keys at the database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
func (OAuthProvider) Edges() []ent.Edge {
	return nil
}

// Indexes of the OAuthProvider.
func (OAuthProvider) Indexes() []ent.Index {
	return []ent.Index{
		// Create a unique index for the provider field to ensure that each provider has only one configuration record | 为provider字段创建唯一索引，确保每个提供商只有一条配置记录
		index.Fields("provider").
			Unique(),
		// Create indexes for commonly queried fields | 为常用查询字段创建索引
		index.Fields("enabled"),
		index.Fields("sort_order"),
	}
}

// Mixin of the OAuthProvider.
func (OAuthProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
