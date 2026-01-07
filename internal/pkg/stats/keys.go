package stats

import "fmt"

// Redis键名常量定义 | Redis key name constant definitions

// 帖子相关Redis键 | Post-related Redis keys
const (
	// PostStatsKeyPrefix Post statistics data Hash key prefix | 帖子统计数据Hash键前缀
	// 格式: post:stats:{post_id} | Format: post:stats:{post_id}
	// Hash字段: like_count, dislike_count, favorite_count, view_count | Hash fields: like_count, dislike_count, favorite_count, view_count
	PostStatsKeyPrefix = "post:stats:"

	// PostUserActionKeyPrefix User's action record Hash key prefix for posts | 用户对帖子的操作记录Hash键前缀
	// 格式: post:action:{user_id}:{post_id} | Format: post:action:{user_id}:{post_id}
	// Hash字段: like, dislike, favorite (值为1表示已操作) | Hash fields: like, dislike, favorite (value is 1 for completed action)
	PostUserActionKeyPrefix = "post:action:"

	// PostDirtySetKey Post dirty data set key | 帖子脏数据集合键
	// 存储需要同步到数据库的帖子ID | Stores post IDs that need to be synchronized to database
	PostDirtySetKey = "post:dirty:set"
)

// 评论相关Redis键 | Comment-related Redis keys
const (
	// CommentStatsKeyPrefix Comment statistics data Hash key prefix | 评论统计数据Hash键前缀
	// 格式: comment:stats:{comment_id} | Format: comment:stats:{comment_id}
	// Hash字段: like_count, dislike_count | Hash fields: like_count, dislike_count
	CommentStatsKeyPrefix = "comment:stats:"

	// CommentUserActionKeyPrefix User's action record Hash key prefix for comments | 用户对评论的操作记录Hash键前缀
	// 格式: comment:action:{user_id}:{comment_id} | Format: comment:action:{user_id}:{comment_id}
	// Hash字段: like, dislike (值为1表示已操作) | Hash fields: like, dislike (value is 1 for completed action)
	CommentUserActionKeyPrefix = "comment:action:"

	// CommentDirtySetKey Comment dirty data set key | 评论脏数据集合键
	// 存储需要同步到数据库的评论ID | Stores comment IDs that need to be synchronized to database
	CommentDirtySetKey = "comment:dirty:set"
)

// GetPostStatsKey Get post statistics data key | 获取帖子统计数据键
func GetPostStatsKey(postID int) string {
	return fmt.Sprintf("%s%d", PostStatsKeyPrefix, postID)
}

// GetPostUserActionKey Get user's action record key for post | 获取用户对帖子的操作记录键
func GetPostUserActionKey(userID, postID int) string {
	return fmt.Sprintf("%s%d:%d", PostUserActionKeyPrefix, userID, postID)
}

// GetCommentStatsKey Get comment statistics data key | 获取评论统计数据键
func GetCommentStatsKey(commentID int) string {
	return fmt.Sprintf("%s%d", CommentStatsKeyPrefix, commentID)
}

// GetCommentUserActionKey Get user's action record key for comment | 获取用户对评论的操作记录键
func GetCommentUserActionKey(userID, commentID int) string {
	return fmt.Sprintf("%s%d:%d", CommentUserActionKeyPrefix, userID, commentID)
}
