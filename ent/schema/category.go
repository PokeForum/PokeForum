package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Category holds the schema definition for the Category entity.
type Category struct {
	ent.Schema
}

// Fields of the Category.
func (Category) Fields() []ent.Field {
	return []ent.Field{
		// Category name | 版块名称
		field.String("name").
			NotEmpty(),
		// Category slug in English, unique | 版块英文标识，唯一
		field.String("slug").
			Unique().
			NotEmpty(),
		// Category description | 版块描述
		field.String("description").
			Optional(),
		// Category icon, can be icon name or http address | 版块图标，可以是icon名称或http地址
		field.String("icon").
			Optional(),
		// Weight for sorting, smaller values come first | 权重排序，越小越靠前
		field.Int("weight").
			Default(0),
		// Category permission/status: Normal, LoginRequired, Hidden, Locked | 版块权限/状态：正常、登录可见、隐藏、锁定
		field.Enum("status").
			Values("Normal", "LoginRequired", "Hidden", "Locked").
			Default("Normal"),
		// Category announcement | 版块公告
		field.Text("announcement").
			Optional(),
	}
}

// Edges of the Category.
// Note: All relationships are only used for ORM queries and will not create foreign keys at the database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Moderator permission management is maintained at the application layer through an intermediate table | 版主权限管理通过中间表在应用层维护
func (Category) Edges() []ent.Edge {
	return nil
}

// Indexes of the Category.
func (Category) Indexes() []ent.Index {
	return []ent.Index{
		// Create indexes for commonly used query fields | 为常用查询字段创建索引
		index.Fields("status"),
		index.Fields("weight"),
		// slug is already a unique field and will automatically create a unique index | slug已经是唯一字段，会自动创建唯一索引
	}
}

// Mixin of the Category.
func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
