package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserSigninStatus 用户签到状态表
// 用途：保存用户最新签到状态（结构化、可查询）
type UserSigninStatus struct {
	ent.Schema
}

// Fields 表字段定义
func (UserSigninStatus) Fields() []ent.Field {
	return []ent.Field{
		// 用户ID（唯一索引）
		field.Int64("user_id").
			Positive().
			Comment("用户ID"),
		// 最近签到日期
		field.Time("last_signin_date").
			SchemaType(map[string]string{"mysql": "date", "postgres": "date"}).
			Comment("最近签到日期"),
		// 连续签到天数
		field.Int("continuous_days").
			NonNegative().
			Default(0).
			Comment("连续签到天数"),
		// 累计签到天数
		field.Int("total_days").
			NonNegative().
			Default(0).
			Comment("累计签到天数"),
	}
}

// Edges 表关联关系
// 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
func (UserSigninStatus) Edges() []ent.Edge {
	return nil
}

// Indexes 表索引
func (UserSigninStatus) Indexes() []ent.Index {
	return []ent.Index{
		// 唯一索引：每个用户只能有一条签到状态记录
		index.Fields("user_id").Unique(),
		// 普通索引：按更新时间查询
		index.Fields("updated_at"),
	}
}

// Mixin 时间混入
func (UserSigninStatus) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
