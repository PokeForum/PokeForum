package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Comment holds the schema definition for the Comment entity.
type Comment struct {
	ent.Schema
}

// Fields of the Comment.
func (Comment) Fields() []ent.Field {
	return []ent.Field{
		// Post ID, foreign key to Post table | 帖子ID，外键关联到Post表
		field.Int("post_id").
			Positive(),
		// Comment user ID, foreign key to User table | 评论用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// Parent comment ID, used to build comment tree, optional | 父评论ID，用于构建评论树，可选
		field.Int("parent_id").
			Optional(),
		// Reply to user ID (@someone), optional | 回复的目标用户ID（@某人），可选
		field.Int("reply_to_user_id").
			Optional(),
		// Comment content, Markdown format | 评论内容，MarkDown格式
		field.Text("content").
			NotEmpty(),
		// Like count, default 0 | 点赞数，默认为0
		field.Int("like_count").
			Default(0).
			NonNegative(),
		// Dislike count, default 0 | 点踩数，默认为0
		field.Int("dislike_count").
			Default(0).
			NonNegative(),
		// Whether it's selected, default false | 是否精选，默认false
		field.Bool("is_selected").
			Default(false),
		// Whether it's pinned, default false | 是否置顶，默认false
		field.Bool("is_pinned").
			Default(false),
		// Commenter IP | 评论者IP
		field.String("commenter_ip").
			Optional(),
		// Commenter device information | 评论者设备信息
		field.String("device_info").
			Optional(),
	}
}

// Edges of the Comment.
// Note: All relationships are only used for ORM queries and will not create foreign keys at the database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Data integrity is guaranteed by application layer logic | 数据完整性由应用层逻辑保证
func (Comment) Edges() []ent.Edge {
	return nil
}

// Indexes of the Comment.
func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		// Create indexes for foreign key fields to optimize query performance | 为外键字段创建索引以优化查询性能
		index.Fields("post_id"),
		index.Fields("user_id"),
		index.Fields("parent_id"),
		index.Fields("reply_to_user_id"),
	}
}

// Mixin of the Comment.
func (Comment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
