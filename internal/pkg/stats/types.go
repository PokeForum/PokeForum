package stats

// ActionType Action type | 行为类型
type ActionType string

const (
	// ActionTypeLike Like | 点赞
	ActionTypeLike ActionType = "Like"
	// ActionTypeDislike Dislike | 点踩
	ActionTypeDislike ActionType = "Dislike"
	// ActionTypeFavorite Favorite | 收藏
	ActionTypeFavorite ActionType = "Favorite"
)

// Stats Statistics data structure | 统计数据结构
type Stats struct {
	// ID Object ID (post ID or comment ID) | ID 对象ID（帖子ID或评论ID）
	ID int `json:"id"`
	// LikeCount Number of likes | LikeCount 点赞数
	LikeCount int `json:"like_count"`
	// DislikeCount Number of dislikes | DislikeCount 点踩数
	DislikeCount int `json:"dislike_count"`
	// FavoriteCount Number of favorites (for posts only) | FavoriteCount 收藏数（仅用于帖子）
	FavoriteCount int `json:"favorite_count,omitempty"`
	// ViewCount Number of views (for posts only) | ViewCount 浏览数（仅用于帖子）
	ViewCount int `json:"view_count,omitempty"`
}

// UserActionStatus User action status | 用户操作状态
type UserActionStatus struct {
	// HasLiked Whether has liked | HasLiked 是否已点赞
	HasLiked bool `json:"has_liked"`
	// HasDisliked Whether has disliked | HasDisliked 是否已点踩
	HasDisliked bool `json:"has_disliked"`
	// HasFavorited Whether has favorited (for posts only) | HasFavorited 是否已收藏（仅用于帖子）
	HasFavorited bool `json:"has_favorited,omitempty"`
}
