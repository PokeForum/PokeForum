package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserFollow 用户关注关系表 | User follow relationship table
type UserFollow struct {
	ent.Schema
}

// Fields of the UserFollow | 用户关注关系表的字段
func (UserFollow) Fields() []ent.Field {
	return []ent.Field{
		field.Int("follower_id").
			Positive().
			Comment("关注者ID | Follower ID"),

		field.Int("following_id").
			Positive().
			Comment("被关注者ID | Following ID"),
	}
}

// Edges of the UserFollow | 用户关注关系表的边
func (UserFollow) Edges() []ent.Edge {
	return []ent.Edge{
		// 注意：虽然系统不使用外键约束，但可以定义edge用于Ent的查询API
		// edge.To("follower", User.Type).
		//     Field("follower_id").
		//     Unique().
		//     Required(),
		// edge.To("following", User.Type).
		//     Field("following_id").
		//     Unique().
		//     Required(),
	}
}

// Indexes of the UserFollow | 用户关注关系表的索引
func (UserFollow) Indexes() []ent.Index {
	return []ent.Index{
		// 复合唯一索引，防止重复关注
		index.Fields("follower_id", "following_id").
			Unique(),

		// 查询某人的粉丝列表
		index.Fields("following_id"),

		// 查询某人关注的人
		index.Fields("follower_id"),

		// 复合索引用于分页查询
		index.Fields("following_id", "created_at"),
		index.Fields("follower_id", "created_at"),
	}
}

// Mixin of the UserFollow | 用户关注关系表的混入
func (UserFollow) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{}, // created_at, updated_at
	}
}
