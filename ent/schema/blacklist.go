package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Blacklist holds the schema definition for the Blacklist entity.
type Blacklist struct {
	ent.Schema
}

// Fields of the Blacklist.
func (Blacklist) Fields() []ent.Field {
	return []ent.Field{
		// User ID, the user who performed the blocking action | 用户ID，执行拉黑操作的用户
		field.Int("user_id").
			Positive().
			Comment("User ID who performed the blocking action | 执行拉黑操作的用户ID"),
		// Blocked user ID | 被拉黑用户ID
		field.Int("blocked_user_id").
			Positive().
			Comment("Blocked user ID | 被拉黑的用户ID"),
	}
}

// Edges of the Blacklist.
// Note: All associations are only used for ORM queries and will not create foreign keys at the database level | 注意: 所有关联关系仅用于ORM查询，不会在数据库层面创建外键
// Data integrity is guaranteed by application layer logic | 数据完整性由应用层逻辑保证
func (Blacklist) Edges() []ent.Edge {
	return nil
}

// Indexes of the Blacklist.
func (Blacklist) Indexes() []ent.Index {
	return []ent.Index{
		// Create index for user ID to optimize query performance | 对用户ID创建索引，优化查询性能
		index.Fields("user_id"),
		// Create index for blocked user ID | 对被拉黑用户ID创建索引
		index.Fields("blocked_user_id"),
		// Create composite index to prevent duplicate blocking | 创建复合索引，防止重复拉黑
		index.Fields("user_id", "blocked_user_id").
			Unique(),
	}
}

// Mixin of the Blacklist.
func (Blacklist) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}
