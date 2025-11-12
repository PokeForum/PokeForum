package stats

// ActionType 行为类型
type ActionType string

const (
	// ActionTypeLike 点赞
	ActionTypeLike ActionType = "Like"
	// ActionTypeDislike 点踩
	ActionTypeDislike ActionType = "Dislike"
	// ActionTypeFavorite 收藏
	ActionTypeFavorite ActionType = "Favorite"
)

// Stats 统计数据结构
type Stats struct {
	// ID 对象ID（帖子ID或评论ID）
	ID int `json:"id"`
	// LikeCount 点赞数
	LikeCount int `json:"like_count"`
	// DislikeCount 点踩数
	DislikeCount int `json:"dislike_count"`
	// FavoriteCount 收藏数（仅用于帖子）
	FavoriteCount int `json:"favorite_count,omitempty"`
	// ViewCount 浏览数（仅用于帖子）
	ViewCount int `json:"view_count,omitempty"`
}

// UserActionStatus 用户操作状态
type UserActionStatus struct {
	// HasLiked 是否已点赞
	HasLiked bool `json:"has_liked"`
	// HasDisliked 是否已点踩
	HasDisliked bool `json:"has_disliked"`
	// HasFavorited 是否已收藏（仅用于帖子）
	HasFavorited bool `json:"has_favorited,omitempty"`
}
