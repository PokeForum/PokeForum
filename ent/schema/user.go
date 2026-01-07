package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		// User ID, database primary key auto-increment | 用户ID，数据库主键自增
		field.Int("id").
			Positive(),
		// Email, unique and not empty | 邮箱，唯一且不为空
		field.String("email").
			Unique().
			NotEmpty(),
		// Password, not empty | 密码，不为空
		field.String("password").
			NotEmpty(),
		// Password salt, not empty | 密码盐，不为空
		field.String("password_salt").
			NotEmpty(),
		// Username, unique and not empty | 用户名，唯一且不为空
		field.String("username").
			Unique().
			NotEmpty(),
		// Avatar URL | 头像URL
		field.String("avatar").
			Optional(),
		// Signature, Markdown format | 签名，MarkDown格式
		field.String("signature").
			Optional(),
		// README, Markdown format | README，MarkDown格式
		field.String("readme").
			Optional(),
		// Whether email is verified, default false | 邮箱是否已验证，默认false
		field.Bool("email_verified").
			Default(false),
		// Experience points, default 0 | 经验值，默认为0
		field.Int("experience").
			Default(0).
			NonNegative(),
		// Points, default 0 | 积分，默认为0
		field.Int("points").
			Default(0).
			NonNegative(),
		// Currency, default 0 | 货币，默认为0
		field.Int("currency").
			Default(0).
			NonNegative(),
		// User status: Normal, Mute, Blocked, Risk control | 用户状态：Normal、Mute、Blocked、Risk control
		field.Enum("status").
			Values("Normal", "Mute", "Blocked", "RiskControl").
			Default("Normal"),
		// User role: User, Moderator, Admin, SuperAdmin | 用户身份：User、Moderator、Admin、SuperAdmin
		field.Enum("role").
			Values("User", "Moderator", "Admin", "SuperAdmin").
			Default("User"),
	}
}

// Edges of the User.
// Note: All relationships are only used for ORM queries and will not create foreign keys at the database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// All user-related associations (moderator permissions, balance records, blacklists, etc.) are maintained at the application layer | 用户相关的所有关联（版主权限、余额记录、黑名单等）均在应用层维护
func (User) Edges() []ent.Edge {
	return nil
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		// email and username are already unique fields and will automatically create unique indexes | email和username已经是唯一字段，会自动创建唯一索引
		// Create indexes for commonly used query fields | 为常用查询字段创建索引
		index.Fields("status"),
		index.Fields("role"),
		index.Fields("email_verified"),
	}
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
