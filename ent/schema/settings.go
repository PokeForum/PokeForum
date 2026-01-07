package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Settings holds the schema definition for the Settings entity.
type Settings struct {
	ent.Schema
}

// Fields of the Settings.
func (Settings) Fields() []ent.Field {
	return []ent.Field{
		// System settings ID, database primary key auto-increment | 系统设置ID，数据库主键自增
		field.Int("id").
			Positive(),
		// Module enumeration: Site, HomePage, Comment, Seo, Security, Function, Signin | 模块枚举：Site、HomePage、Comment、Seo、Security、Function、Signin
		field.Enum("module").
			Values("Site", "HomePage", "Comment", "Seo", "Security", "Function", "Signin"),
		// Configuration key, unique identifier | 配置键，唯一标识
		field.String("key").
			NotEmpty(),
		// Configuration value | 配置值
		field.String("value").
			Optional(),
		// Data type enumeration: string, number, boolean, json, text, defaults to string | 数据类型枚举：string、number、boolean、json、text，默认为string
		field.Enum("value_type").
			Values("string", "number", "boolean", "json", "text").
			Default("string"),
	}
}

// Indexes of the Settings.
func (Settings) Indexes() []ent.Index {
	return []ent.Index{
		// Unique composite index on module and key | module和key的组合唯一索引
		index.Fields("module", "key").
			Unique(),
	}
}

// Mixin of the Settings.
func (Settings) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
