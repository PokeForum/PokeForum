package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserLoginLog holds the schema definition for the UserLoginLog entity.
type UserLoginLog struct {
	ent.Schema
}

// Fields of the UserLoginLog.
func (UserLoginLog) Fields() []ent.Field {
	return []ent.Field{
		// Primary key ID | 主键ID
		field.Int("id").
			Positive(),
		// User ID, foreign key reference to User table | 用户ID，外键关联到User表
		field.Int("user_id").
			Positive(),
		// IP address | IP地址
		field.String("ip_address").
			NotEmpty(),
		// IP country | IP国家
		field.String("ip_country").
			Optional(),
		// IP city | IP城市
		field.String("ip_city").
			Optional(),
		// Whether login was successful | 是否成功登录
		field.Bool("success").
			Default(true),
		// Device information, extracted from browser User-Agent | 设备信息，从浏览器UA中读取
		field.String("device_info").
			Optional(),
	}
}

// Edges of the UserLoginLog.
// Note: All associations are only for ORM queries, no foreign keys will be created at database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Data integrity is guaranteed by application logic | 数据完整性由应用层逻辑保证
func (UserLoginLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserLoginLog.
func (UserLoginLog) Indexes() []ent.Index {
	return []ent.Index{
		// Create index on association field to optimize query performance | 为关联字段创建索引以优化查询性能
		index.Fields("user_id"),
		// Create index on frequently queried field | 为常用查询字段创建索引
		index.Fields("success"),
		// Create composite index to optimize user login history queries | 创建复合索引优化用户登录历史查询
		index.Fields("user_id", "created_at"),
	}
}

// Mixin of the UserLoginLog.
func (UserLoginLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
