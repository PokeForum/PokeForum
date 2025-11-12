package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CategoryModerator holds the schema definition for the CategoryModerator entity.
// 版块版主关联表，用于维护版块和版主的多对多关系
type CategoryModerator struct {
	ent.Schema
}

// Fields of the CategoryModerator.
func (CategoryModerator) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID
		field.Int("id").
			Positive(),
		// 版块ID
		field.Int("category_id").
			Positive().
			Comment("版块ID"),
		// 用户ID（版主）
		field.Int("user_id").
			Positive().
			Comment("版主用户ID"),
	}
}

// Edges of the CategoryModerator.
// 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// 数据完整性由应用层逻辑保证
func (CategoryModerator) Edges() []ent.Edge {
	return nil
}

// Indexes of the CategoryModerator.
func (CategoryModerator) Indexes() []ent.Index {
	return []ent.Index{
		// 为关联字段创建索引以优化查询性能
		index.Fields("category_id"),
		index.Fields("user_id"),
		// 创建复合唯一索引，确保同一版块的同一用户只能有一条版主记录
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
