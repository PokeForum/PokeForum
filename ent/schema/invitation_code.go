package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InvitationCode holds the schema definition for the InvitationCode entity.
// 邀请码表，记录邀请码详细数据，方便管理员审计
type InvitationCode struct {
	ent.Schema
}

// Fields of the InvitationCode.
func (InvitationCode) Fields() []ent.Field {
	return []ent.Field{
		// Invitation code (unique) | 邀请码（唯一）
		field.String("code").
			Unique().
			NotEmpty().
			Comment("Invitation code (unique) | 邀请码（唯一）"),
		// Creator user ID | 创建者用户ID
		field.Int("creator_id").
			Positive().
			Comment("Creator user ID | 创建者用户ID"),
		// Used by user ID (null if not used) | 使用者用户ID（未使用时为空）
		field.Int("used_by_id").
			Optional().
			Nillable().
			Comment("Used by user ID (null if not used) | 使用者用户ID（未使用时为空）"),
		// Status: unused, used, expired, disabled | 状态：unused（未使用）、used（已使用）、expired（已过期）、disabled（已禁用）
		field.Enum("status").
			Values("unused", "used", "expired", "disabled").
			Default("unused").
			Comment("Status: unused, used, expired, disabled | 状态：unused、used、expired、disabled"),
		// Generation mode: direct, points, currency | 生成方式：direct（直接）、points（积分）、currency（货币）
		field.Enum("generation_mode").
			Values("direct", "points", "currency").
			Default("direct").
			Comment("Generation mode: direct, points, currency | 生成方式：direct、points、currency"),
		// Cost amount (only for points/currency mode) | 消耗数量（仅限积分/货币模式）
		field.Int("cost_amount").
			Default(0).
			NonNegative().
			Comment("Cost amount (only for points/currency mode) | 消耗数量（仅限积分/货币模式）"),
		// Used at timestamp | 使用时间
		field.Time("used_at").
			Optional().
			Nillable().
			Comment("Used at timestamp | 使用时间"),
		// Expiration time (null means never expires) | 过期时间（为空表示永不过期）
		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("Expiration time (null means never expires) | 过期时间"),
		// Used IP address | 使用时的IP地址
		field.String("used_ip").
			Optional().
			Comment("Used IP address | 使用时的IP地址"),
		// Used user agent | 使用时的用户代理
		field.String("used_user_agent").
			Optional().
			Comment("Used user agent | 使用时的用户代理"),
		// Remark (admin notes) | 备注（管理员备注）
		field.String("remark").
			Optional().
			Comment("Remark (admin notes) | 备注"),
	}
}

// Edges of the InvitationCode.
// Note: All relationships are for ORM queries only, no foreign keys will be created at database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
func (InvitationCode) Edges() []ent.Edge {
	return nil
}

// Indexes of the InvitationCode.
func (InvitationCode) Indexes() []ent.Index {
	return []ent.Index{
		// Index on creator_id for querying user's created codes | 创建者ID索引，用于查询用户创建的邀请码
		index.Fields("creator_id"),
		// Index on used_by_id for querying which code a user used | 使用者ID索引，用于查询用户使用的邀请码
		index.Fields("used_by_id"),
		// Index on status for filtering | 状态索引，用于筛选
		index.Fields("status"),
		// Composite index for creator's code list with status | 复合索引，用于查询创建者的邀请码列表
		index.Fields("creator_id", "status"),
		// Index on expires_at for expiration check | 过期时间索引，用于过期检查
		index.Fields("expires_at"),
	}
}

// Mixin of the InvitationCode.
func (InvitationCode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
