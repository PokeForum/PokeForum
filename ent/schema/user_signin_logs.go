package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserSigninLogs User sign-in log table | 用户签到日志表
// Purpose: Save every sign-in record (audit, security, user history) | 用途：保存每一笔签到记录（审计、安全、用户历史）
type UserSigninLogs struct {
	ent.Schema
}

// Fields Table field definitions | 表字段定义
func (UserSigninLogs) Fields() []ent.Field {
	return []ent.Field{
		// User ID | 用户ID
		field.Int64("user_id").
			Positive().
			Comment("User ID | 用户ID"),
		// Sign-in date | 签到日期
		field.Time("sign_date").
			SchemaType(map[string]string{"mysql": "date", "postgres": "date"}).
			Comment("Sign-in date | 签到日期"),
	}
}

// Edges Table relationships | 表关联关系
// Note: All relationships are for ORM queries only, no foreign keys will be created at database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
func (UserSigninLogs) Edges() []ent.Edge {
	return nil
}

// Indexes Table indexes | 表索引
func (UserSigninLogs) Indexes() []ent.Index {
	return []ent.Index{
		// Composite index: user ID + sign-in date, for querying user sign-in history | 复合索引：用户ID + 签到日期，用于查询用户签到历史
		index.Fields("user_id", "sign_date"),
		// Regular index: by sign-in date, for statistics and leaderboards | 普通索引：按签到日期查询，用于统计和排行榜
		index.Fields("sign_date"),
	}
}

// Mixin Time mixin | 时间混入
func (UserSigninLogs) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
