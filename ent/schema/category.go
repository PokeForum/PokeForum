package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Category holds the schema definition for the Category entity.
type Category struct {
	ent.Schema
}

// Fields of the Category.
func (Category) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID
		field.Int("id").
			Positive(),
		// 版块名称
		field.String("name").
			NotEmpty(),
		// 版块英文标识，唯一
		field.String("slug").
			Unique().
			NotEmpty(),
		// 版块描述
		field.String("description").
			Optional(),
		// 版块图标，可以是icon名称或http地址
		field.String("icon").
			Optional(),
		// 权重排序，越小越靠前
		field.Int("weight").
			Default(0),
		// 版块权限/状态：正常、登录可见、隐藏、锁定
		field.Enum("status").
			Values("Normal", "LoginRequired", "Hidden", "Locked").
			Default("Normal"),
		// 版块公告
		field.Text("announcement").
			Optional(),
	}
}

// Edges of the Category.
func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		// 版块与用户的多对多关联，用于版主权限管理
		// 反向关联到User的managed_categories
		edge.From("moderators", User.Type).
			Ref("managed_categories"),
	}
}

// Mixin of the Category.
func (Category) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
