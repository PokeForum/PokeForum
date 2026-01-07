package schema

import (
	"entgo.io/ent"
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
		// Primary key ID | 主键ID
		field.Int("id").
			Positive(),
		// User ID, foreign key reference to User table | 用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// Comment ID, foreign key reference to Comment table | 评论ID，外键关联到Comment表
		field.Int("comment_id").
			Positive(),
		// Action type: Like (upvote), Dislike (downvote) | 行为类型：Like（点赞）、Dislike（点踩）
		field.Enum("action_type").
			Values("Like", "Dislike"),
	}
}

// Edges of the CommentAction.
// Note: All relationships are for ORM queries only, no foreign keys will be created at database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Data integrity is ensured by application layer logic | 数据完整性由应用层逻辑保证
func (CommentAction) Edges() []ent.Edge {
	return nil
}

// Indexes of the CommentAction.
func (CommentAction) Indexes() []ent.Index {
	return []ent.Index{
		// Create indexes on relational fields to optimize query performance | 为关联字段创建索引以优化查询性能
		index.Fields("user_id"),
		index.Fields("comment_id"),
		// Create composite unique index to ensure each user can only have one record for the same action on the same comment | 创建复合唯一索引，确保每个用户对同一评论的同一行为只能有一条记录
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
