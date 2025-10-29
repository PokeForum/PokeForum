package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID
		field.Int("id").
			Positive(),
		// 用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// 版块ID，外键关联到Category表
		field.Int("category_id").
			Positive(),
		// 帖子标题
		field.String("title").
			NotEmpty(),
		// 帖子正文内容，MarkDown格式
		field.String("content").
			NotEmpty(),
		// 阅读限制
		field.String("read_permission").
			Optional(),
		// 浏览数，默认为0
		field.Int("view_count").
			Default(0).
			NonNegative(),
		// 点赞数，默认为0
		field.Int("like_count").
			Default(0).
			NonNegative(),
		// 点踩数，默认为0
		field.Int("dislike_count").
			Default(0).
			NonNegative(),
		// 收藏数，默认为0
		field.Int("favorite_count").
			Default(0).
			NonNegative(),
		// 是否精华帖，默认false
		field.Bool("is_essence").
			Default(false),
		// 是否置顶，默认false
		field.Bool("is_pinned").
			Default(false),
		// 发布IP
		field.String("publish_ip").
			Optional(),
		// 帖子状态：Normal、Locked、Draft、Private、Ban
		field.Enum("status").
			Values("Normal", "Locked", "Draft", "Private", "Ban").
			Default("Normal"),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		// 关联到User表
		edge.To("author", User.Type).
			Field("user_id").
			Unique().
			Required(),
		// 关联到Category表
		edge.To("category", Category.Type).
			Field("category_id").
			Unique().
			Required(),
	}
}

// Mixin of the Post.
func (Post) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
