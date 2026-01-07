package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CategoryModerator holds the schema definition for the CategoryModerator entity.
// Category moderator association table, used to maintain many-to-many relationships between categories and moderators | 版块版主关联表，用于维护版块和版主的多对多关系
type CategoryModerator struct {
	ent.Schema
}

// Fields of the CategoryModerator.
func (CategoryModerator) Fields() []ent.Field {
	return []ent.Field{
		// Primary key ID | 主键ID
		field.Int("id").
			Positive(),
		// Category ID | 版块ID
		field.Int("category_id").
			Positive().
			Comment("Category ID | 版块ID"),
		// User ID (moderator) | 用户ID（版主）
		field.Int("user_id").
			Positive().
			Comment("Moderator user ID | 版主用户ID"),
	}
}

// Edges of the CategoryModerator.
// Note: All associations are only used for ORM queries and will not create foreign keys at the database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Data integrity is guaranteed by application layer logic | 数据完整性由应用层逻辑保证
func (CategoryModerator) Edges() []ent.Edge {
	return nil
}

// Indexes of the CategoryModerator.
func (CategoryModerator) Indexes() []ent.Index {
	return []ent.Index{
		// Create indexes for association fields to optimize query performance | 为关联字段创建索引以优化查询性能
		index.Fields("category_id"),
		index.Fields("user_id"),
		// Create a composite unique index to ensure that the same user in the same category can only have one moderator record | 创建复合唯一索引，确保同一版块的同一用户只能有一条版主记录
		index.Fields("category_id", "user_id").
			Unique(),
	}
}

// Mixin of the CategoryModerator.
func (CategoryModerator) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
