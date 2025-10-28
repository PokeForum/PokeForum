package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		// 用户ID，数据库主键自增
		field.Int("id").
			Positive(),
		// 邮箱，唯一且不为空
		field.String("email").
			Unique().
			NotEmpty(),
		// 密码，不为空
		field.String("password").
			NotEmpty(),
		// 密码盐，不为空
		field.String("password_salt").
			NotEmpty(),
		// 用户名，唯一且不为空
		field.String("username").
			Unique().
			NotEmpty(),
		// 头像URL
		field.String("avatar").
			Optional(),
		// 签名，MarkDown格式
		field.String("signature").
			Optional(),
		// README，MarkDown格式
		field.String("readme").
			Optional(),
		// 邮箱是否已验证，默认false
		field.Bool("email_verified").
			Default(false),
		// 积分，默认为0
		field.Int("points").
			Default(0).
			NonNegative(),
		// 货币，默认为0
		field.Int("currency").
			Default(0).
			NonNegative(),
		// 帖子数，默认为0
		field.Int("post_count").
			Default(0).
			NonNegative(),
		// 评论数，默认为0
		field.Int("comment_count").
			Default(0).
			NonNegative(),
		// 用户状态：Normal、Mute、Blocked、Activation pending、Risk control
		field.Enum("status").
			Values("Normal", "Mute", "Blocked", "ActivationPending", "RiskControl").
			Default("Normal"),
		// 用户身份：User、Moderator、Admin、SuperAdmin
		field.Enum("role").
			Values("User", "Moderator", "Admin", "SuperAdmin").
			Default("User"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// 用户与版块的多对多关联，用于版主权限管理
		// 当用户身份为Moderator时，可以通过此关联查询其管理的版块
		edge.To("managed_categories", Category.Type),
	}
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
