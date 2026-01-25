package _const

// Cache key constants | 缓存键常量
const (
	// UserCategoryListCacheKey User category list cache key (guest) | 用户版块列表缓存键（未登录）
	UserCategoryListCacheKey = "user:categories:list"
	// UserCategoryListLoggedInCacheKey User category list cache key (logged in) | 用户版块列表缓存键（已登录）
	UserCategoryListLoggedInCacheKey = "user:categories:list:logged_in"
)
