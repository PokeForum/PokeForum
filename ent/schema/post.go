package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		field.Text("content").
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
		// 最后编辑时间，可选字段，用于记录帖子的最后编辑时间
		field.Time("last_edited_at").
			Optional(),
	}
}

// Edges of the Post.
// 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// 数据完整性由应用层逻辑保证
func (Post) Edges() []ent.Edge {
	return nil
}

// Indexes of the Post.
func (Post) Indexes() []ent.Index {
	return []ent.Index{
		// 为关联字段创建索引以优化查询性能
		index.Fields("user_id"),
		index.Fields("category_id"),
		// 为常用查询字段创建索引
		index.Fields("status"),
		index.Fields("is_essence"),
		index.Fields("is_pinned"),
		// 为最后编辑时间创建索引
		index.Fields("last_edited_at"),
		// 创建复合索引优化版块内帖子查询
		index.Fields("category_id", "status", "created_at"),
	}
}

// Mixin of the Post.
func (Post) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
