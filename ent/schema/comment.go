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
// 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// 数据完整性由应用层逻辑保证
func (Comment) Edges() []ent.Edge {
	return nil
}

// Indexes of the Comment.
func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		// 为外键字段创建索引以优化查询性能
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
