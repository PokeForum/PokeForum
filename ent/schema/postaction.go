package schema

import (
	"entgo.io/ent"
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
		// User ID, foreign key reference to User table | 用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// Post ID, foreign key reference to Post table | 帖子ID，外键关联到Post表
		field.Int("post_id").
			Positive(),
		// Action type: Like (upvote), Dislike (downvote), Favorite (bookmark) | 行为类型：Like（点赞）、Dislike（点踩）、Favorite（收藏）
		field.Enum("action_type").
			Values("Like", "Dislike", "Favorite"),
	}
}

// Edges of the PostAction.
// Note: All relationships are for ORM queries only, no foreign keys will be created at database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Data integrity is ensured by application layer logic | 数据完整性由应用层逻辑保证
func (PostAction) Edges() []ent.Edge {
	return nil
}

// Indexes of the PostAction.
func (PostAction) Indexes() []ent.Index {
	return []ent.Index{
		// Create indexes on relational fields to optimize query performance | 为关联字段创建索引以优化查询性能
		index.Fields("user_id"),
		index.Fields("post_id"),
		// Create composite unique index to ensure each user can only have one record for the same action on the same post | 创建复合唯一索引，确保每个用户对同一帖子的同一行为只能有一条记录
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
