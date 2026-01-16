package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
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
		// User ID, foreign key to User table | 用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// Category ID, foreign key to Category table | 版块ID，外键关联到Category表
		field.Int("category_id").
			Positive(),
		// Post title | 帖子标题
		field.String("title").
			NotEmpty(),
		// Post content, Markdown format | 帖子正文内容，MarkDown格式
		field.Text("content").
			NotEmpty(),
		// Read permission | 阅读限制
		field.String("read_permission").
			Default("public").
			Optional(),
		// View count, default 0 | 浏览数，默认为0
		field.Int("view_count").
			Default(0).
			NonNegative(),
		// Like count, default 0 | 点赞数，默认为0
		field.Int("like_count").
			Default(0).
			NonNegative(),
		// Dislike count, default 0 | 点踩数，默认为0
		field.Int("dislike_count").
			Default(0).
			NonNegative(),
		// Favorite count, default 0 | 收藏数，默认为0
		field.Int("favorite_count").
			Default(0).
			NonNegative(),
		// Whether it's an essence post, default false | 是否精华帖，默认false
		field.Bool("is_essence").
			Default(false),
		// Whether it's pinned, default false | 是否置顶，默认false
		field.Bool("is_pinned").
			Default(false),
		// Publish IP | 发布IP
		field.String("publish_ip").
			Optional(),
		// Post status: Normal, Locked, Draft, Private, Ban | 帖子状态：Normal、Locked、Draft、Private、Ban
		field.Enum("status").
			Values("Normal", "Locked", "Draft", "Private", "Ban").
			Default("Normal"),
		// Last edited time, optional field to record the last edit time of the post | 最后编辑时间，可选字段，用于记录帖子的最后编辑时间
		field.Time("last_edited_at").
			Optional(),
	}
}

// Edges of the Post.
// Note: All relationships are only used for ORM queries and will not create foreign keys at the database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Data integrity is guaranteed by application layer logic | 数据完整性由应用层逻辑保证
func (Post) Edges() []ent.Edge {
	return nil
}

// Indexes of the Post.
func (Post) Indexes() []ent.Index {
	return []ent.Index{
		// Create indexes for associated fields to optimize query performance | 为关联字段创建索引以优化查询性能
		index.Fields("user_id"),
		index.Fields("category_id"),
		// Create indexes for commonly used query fields | 为常用查询字段创建索引
		index.Fields("status"),
		index.Fields("is_essence"),
		index.Fields("is_pinned"),
		// Create index for last edited time | 为最后编辑时间创建索引
		index.Fields("last_edited_at"),
		// Create composite index to optimize post queries within categories | 创建复合索引优化版块内帖子查询
		index.Fields("category_id", "status", "created_at"),
		// GIN index for title field to support fast fuzzy search | 为 title 字段创建 GIN 索引以支持快速模糊搜索
		index.Fields("title").
			Annotations(entsql.IndexTypes(map[string]string{
				"postgres": "GIN",
			})).
			Annotations(entsql.OpClass("gin_trgm_ops")),
	}
}

// Annotations of the Post.
func (Post) Annotations() []schema.Annotation {
	return nil
}

// Mixin of the Post.
func (Post) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
