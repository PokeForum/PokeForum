package stats

import "fmt"

// Redis键名常量定义

// 帖子相关Redis键
const (
	// PostStatsKeyPrefix 帖子统计数据Hash键前缀
	// 格式: post:stats:{post_id}
	// Hash字段: like_count, dislike_count, favorite_count, view_count
	PostStatsKeyPrefix = "post:stats:"

	// PostUserActionKeyPrefix 用户对帖子的操作记录Hash键前缀
	// 格式: post:action:{user_id}:{post_id}
	// Hash字段: like, dislike, favorite (值为1表示已操作)
	PostUserActionKeyPrefix = "post:action:"

	// PostDirtySetKey 帖子脏数据集合键
	// 存储需要同步到数据库的帖子ID
	PostDirtySetKey = "post:dirty:set"
)

// 评论相关Redis键
const (
	// CommentStatsKeyPrefix 评论统计数据Hash键前缀
	// 格式: comment:stats:{comment_id}
	// Hash字段: like_count, dislike_count
	CommentStatsKeyPrefix = "comment:stats:"

	// CommentUserActionKeyPrefix 用户对评论的操作记录Hash键前缀
	// 格式: comment:action:{user_id}:{comment_id}
	// Hash字段: like, dislike (值为1表示已操作)
	CommentUserActionKeyPrefix = "comment:action:"

	// CommentDirtySetKey 评论脏数据集合键
	// 存储需要同步到数据库的评论ID
	CommentDirtySetKey = "comment:dirty:set"
)

// GetPostStatsKey 获取帖子统计数据键
func GetPostStatsKey(postID int) string {
	return fmt.Sprintf("%s%d", PostStatsKeyPrefix, postID)
}

// GetPostUserActionKey 获取用户对帖子的操作记录键
func GetPostUserActionKey(userID, postID int) string {
	return fmt.Sprintf("%s%d:%d", PostUserActionKeyPrefix, userID, postID)
}

// GetCommentStatsKey 获取评论统计数据键
func GetCommentStatsKey(commentID int) string {
	return fmt.Sprintf("%s%d", CommentStatsKeyPrefix, commentID)
}

// GetCommentUserActionKey 获取用户对评论的操作记录键
func GetCommentUserActionKey(userID, commentID int) string {
	return fmt.Sprintf("%s%d:%d", CommentUserActionKeyPrefix, userID, commentID)
}
