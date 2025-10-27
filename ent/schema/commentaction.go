package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CommentAction holds the schema definition for the CommentAction entity.
type CommentAction struct {
	ent.Schema
}

// Fields of the CommentAction.
func (CommentAction) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID
		field.Int("id").
			Positive(),
		// 用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// 评论ID，外键关联到Comment表
		field.Int("comment_id").
			Positive(),
		// 行为类型：Like（点赞）、Dislike（点踩）
		field.Enum("action_type").
			Values("Like", "Dislike"),
	}
}

// Edges of the CommentAction.
func (CommentAction) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联到User表
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		// 关联到Comment表
		edge.To("comment", Comment.Type).
			Field("comment_id").
			Unique().
			Required(),
	}
}

// Indexes of the CommentAction.
func (CommentAction) Indexes() []ent.Index {
	return []ent.Index{
		// 创建复合唯一索引，确保每个用户对同一评论的同一行为只能有一条记录
		index.Fields("user_id", "comment_id", "action_type").
			Unique(),
	}
}

// Mixin of the CommentAction.
func (CommentAction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
