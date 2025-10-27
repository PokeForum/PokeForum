package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PostAction holds the schema definition for the PostAction entity.
type PostAction struct {
	ent.Schema
}

// Fields of the PostAction.
func (PostAction) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID
		field.Int("id").
			Positive(),
		// 用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// 帖子ID，外键关联到Post表
		field.Int("post_id").
			Positive(),
		// 行为类型：Like（点赞）、Dislike（点踩）、Favorite（收藏）
		field.Enum("action_type").
			Values("Like", "Dislike", "Favorite"),
	}
}

// Edges of the PostAction.
func (PostAction) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联到User表
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		// 关联到Post表
		edge.To("post", Post.Type).
			Field("post_id").
			Unique().
			Required(),
	}
}

// Indexes of the PostAction.
func (PostAction) Indexes() []ent.Index {
	return []ent.Index{
		// 创建复合唯一索引，确保每个用户对同一帖子的同一行为只能有一条记录
		index.Fields("user_id", "post_id", "action_type").
			Unique(),
	}
}

// Mixin of the PostAction.
func (PostAction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
