package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Comment holds the schema definition for the Comment entity.
type Comment struct {
	ent.Schema
}

// Fields of the Comment.
func (Comment) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID
		field.Int("id").
			Positive(),
		// 帖子ID，外键关联到Post表
		field.Int("post_id").
			Positive(),
		// 评论用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// 父评论ID，用于构建评论树，可选
		field.Int("parent_id").
			Optional(),
		// 回复的目标用户ID（@某人），可选
		field.Int("reply_to_user_id").
			Optional(),
		// 评论内容，MarkDown格式
		field.String("content").
			NotEmpty(),
		// 点赞数，默认为0
		field.Int("like_count").
			Default(0).
			NonNegative(),
		// 点踩数，默认为0
		field.Int("dislike_count").
			Default(0).
			NonNegative(),
		// 是否精选，默认false
		field.Bool("is_selected").
			Default(false),
		// 是否置顶，默认false
		field.Bool("is_pinned").
			Default(false),
		// 评论者IP
		field.String("commenter_ip").
			Optional(),
		// 评论者设备信息
		field.String("device_info").
			Optional(),
	}
}

// Edges of the Comment.
func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联到Post表
		edge.To("post", Post.Type).
			Field("post_id").
			Unique().
			Required(),
		// 关联到评论用户
		edge.To("author", User.Type).
			Field("user_id").
			Unique().
			Required(),
		// 关联到父评论（自引用）
		edge.To("parent", Comment.Type).
			Field("parent_id").
			Unique(),
		// 关联到回复的目标用户
		edge.To("reply_to_user", User.Type).
			Field("reply_to_user_id").
			Unique(),
	}
}

// Mixin of the Comment.
func (Comment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
