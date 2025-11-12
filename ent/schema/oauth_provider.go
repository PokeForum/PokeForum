package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OAuthProvider holds the schema definition for the OAuthProvider entity.
// 存储第三方OAuth登录提供商的配置信息
type OAuthProvider struct {
	ent.Schema
}

// Fields of the OAuthProvider.
func (OAuthProvider) Fields() []ent.Field {
	return []ent.Field{
		// 提供商ID，数据库主键自增
		field.Int("id").
			Positive(),
		// 提供商名称：QQ、GitHub、Apple、Google、Telegram、FIDO2
		field.Enum("provider").
			Values("QQ", "GitHub", "Apple", "Google", "Telegram", "FIDO2"),
		// 客户端ID（Client ID）
		field.String("client_id").
			Optional(),
		// 客户端密钥（Client Secret），敏感信息
		field.String("client_secret").
			Optional().
			Sensitive(),
		// 授权URL
		field.String("auth_url").
			Optional(),
		// Token获取URL
		field.String("token_url").
			Optional(),
		// 用户信息获取URL
		field.String("user_info_url").
			Optional(),
		// 回调URL
		field.String("redirect_url").
			Optional(),
		// 请求范围（Scopes），JSON数组格式存储
		field.JSON("scopes", []string{}).
			Optional(),
		// 额外配置参数，JSON格式存储，用于特殊提供商的自定义配置
		field.JSON("extra_config", map[string]interface{}{}).
			Optional(),
		// 是否启用该提供商
		field.Bool("enabled").
			Default(false),
		// 排序顺序，用于前端展示排序
		field.Int("sort_order").
			Default(0).
			NonNegative(),
	}
}

// Edges of the OAuthProvider.
// 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
func (OAuthProvider) Edges() []ent.Edge {
	return nil
}

// Indexes of the OAuthProvider.
func (OAuthProvider) Indexes() []ent.Index {
	return []ent.Index{
		// 为provider字段创建唯一索引，确保每个提供商只有一条配置记录
		index.Fields("provider").
			Unique(),
		// 为常用查询字段创建索引
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
