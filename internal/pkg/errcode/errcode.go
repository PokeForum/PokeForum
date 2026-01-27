package errcode

type ErrCode string

const (
	Success                     ErrCode = "SUCCESS"
	InvalidParam                ErrCode = "INVALID_PARAM"
	NoPermission                ErrCode = "NO_PERMISSION"
	GenericError                ErrCode = "GENERIC_ERROR"
	ServerBusy                  ErrCode = "SERVER_BUSY"
	TooManyRequests             ErrCode = "TOO_MANY_REQUESTS"
	NeedLogin                   ErrCode = "NEED_LOGIN"
	UserNotFound                ErrCode = "USER_NOT_FOUND"
	UserDisabled                ErrCode = "USER_DISABLED"
	UserStatusInvalid           ErrCode = "USER_STATUS_INVALID"
	PostNotFound                ErrCode = "POST_NOT_FOUND"
	PostDeleted                 ErrCode = "POST_DELETED"
	PostLocked                  ErrCode = "POST_LOCKED"
	CategoryNotFound            ErrCode = "CATEGORY_NOT_FOUND"
	CategoryDisabled            ErrCode = "CATEGORY_DISABLED"
	CommentNotFound             ErrCode = "COMMENT_NOT_FOUND"
	InvalidPassword             ErrCode = "INVALID_PASSWORD"
	InvalidToken                ErrCode = "INVALID_TOKEN"
	TokenExpired                ErrCode = "TOKEN_EXPIRED"
	DuplicateUsername           ErrCode = "DUPLICATE_USERNAME"
	DuplicateEmail              ErrCode = "DUPLICATE_EMAIL"
	InvalidEmail                ErrCode = "INVALID_EMAIL"
	InvalidVerificationCode     ErrCode = "INVALID_VERIFICATION_CODE"
	VerificationCodeExpired     ErrCode = "VERIFICATION_CODE_EXPIRED"
	VerificationCodeUsed        ErrCode = "VERIFICATION_CODE_USED"
	OAuthProviderNotFound       ErrCode = "OAUTH_PROVIDER_NOT_FOUND"
	OAuthBindFailed             ErrCode = "OAUTH_BIND_FAILED"
	PostEditPermissionDenied    ErrCode = "POST_EDIT_PERMISSION_DENIED"
	PostPrivatePermissionDenied ErrCode = "POST_PRIVATE_PERMISSION_DENIED"
	InsufficientPoints          ErrCode = "INSUFFICIENT_POINTS"
	SignInAlreadyDone           ErrCode = "SIGNIN_ALREADY_DONE"
	SignInRecordNotFound        ErrCode = "SIGNIN_RECORD_NOT_FOUND"
	MaxDraftsReached            ErrCode = "MAX_DRAFTS_REACHED"
	DatabaseError               ErrCode = "DATABASE_ERROR"
	CacheError                  ErrCode = "CACHE_ERROR"
	RedisError                  ErrCode = "REDIS_ERROR"
	LoginRequired               ErrCode = "LOGIN_REQUIRED"
)

var errCodeMsgMap = map[ErrCode]string{
	Success:                     "Success",
	InvalidParam:                "Invalid request parameters | 请求参数错误",
	NoPermission:                "No read permission | 无阅读权限",
	GenericError:                "Error",
	ServerBusy:                  "The system is busy, please try again later | 系统繁忙，请稍候再试",
	TooManyRequests:             "Too many requests, please try again later | 请求过于频繁，请稍后再试",
	NeedLogin:                   "Not logged in | 未登录",
	UserNotFound:                "User not found | 用户不存在",
	UserDisabled:                "User account is disabled | 用户账号已被禁用",
	UserStatusInvalid:           "User status is invalid | 用户状态异常",
	PostNotFound:                "Post not found | 帖子不存在",
	PostDeleted:                 "Post has been deleted | 帖子已被删除",
	PostLocked:                  "Post has been locked | 帖子已被锁定",
	CategoryNotFound:            "Category not found | 版块不存在",
	CategoryDisabled:            "Category is disabled | 版块已被禁用",
	CommentNotFound:             "Comment not found | 评论不存在",
	InvalidPassword:             "Invalid password | 密码错误",
	InvalidToken:                "Invalid token | 令牌无效",
	TokenExpired:                "Token has expired | 令牌已过期",
	DuplicateUsername:           "Username already exists | 用户名已存在",
	DuplicateEmail:              "Email already exists | 邮箱已存在",
	InvalidEmail:                "Invalid email format | 邮箱格式错误",
	InvalidVerificationCode:     "Invalid verification code | 验证码错误",
	VerificationCodeExpired:     "Verification code has expired | 验证码已过期",
	VerificationCodeUsed:        "Verification code has been used | 验证码已使用",
	OAuthProviderNotFound:       "OAuth provider not found | OAuth提供商不存在",
	OAuthBindFailed:             "Failed to bind OAuth account | OAuth账号绑定失败",
	PostEditPermissionDenied:    "No permission to edit post | 无编辑帖子权限",
	PostPrivatePermissionDenied: "No permission to set post as private | 无设置帖子私有权限",
	InsufficientPoints:          "Insufficient points | 积分不足",
	SignInAlreadyDone:           "Already signed in today | 今日已签到",
	SignInRecordNotFound:        "Sign-in record not found | 签到记录不存在",
	MaxDraftsReached:            "Maximum number of drafts reached | 已达到草稿数量上限",
	DatabaseError:               "Database operation failed | 数据库操作失败",
	CacheError:                  "Cache operation failed | 缓存操作失败",
	RedisError:                  "Redis operation failed | Redis操作失败",
	LoginRequired:               "Login required | 需要登录",
}

func (c ErrCode) Msg() string {
	msg, ok := errCodeMsgMap[c]
	if !ok {
		msg = errCodeMsgMap[ServerBusy]
	}
	return msg
}
