package schema

// UserListRequest 用户列表查询请求体
type UserListRequest struct {
	Page     int    `form:"page" binding:"required,min=1" example:"1"`               // 页码
	PageSize int    `form:"page_size" binding:"required,min=1,max=100" example:"20"` // 每页数量
	Keyword  string `form:"keyword" example:"test"`                                  // 搜索关键词（用户名或邮箱）
	Status   string `form:"status" example:"Normal"`                                 // 用户状态筛选
	Role     string `form:"role" example:"User"`                                     // 用户身份筛选
}

// UserCreateRequest 创建用户请求体
type UserCreateRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=100" example:"testuser"`                 // 用户名
	Email     string `json:"email" binding:"required,email" example:"test@example.com"`                    // 邮箱
	Password  string `json:"password" binding:"required,min=8" example:"password123"`                      // 密码
	Role      string `json:"role" binding:"required,oneof=User Moderator Admin SuperAdmin" example:"User"` // 用户身份
	Avatar    string `json:"avatar" example:"https://example.com/avatar.jpg"`                              // 头像URL
	Signature string `json:"signature" example:"这是我的个性签名"`                                                 // 个性签名
	Readme    string `json:"readme" example:"## 关于我\n这是我的README内容"`                                        // README内容
}

// UserUpdateRequest 更新用户信息请求体
type UserUpdateRequest struct {
	ID        int    `json:"id" binding:"required" example:"1"`                             // 用户ID
	Username  string `json:"username" binding:"omitempty,min=3,max=100" example:"testuser"` // 用户名
	Email     string `json:"email" binding:"omitempty,email" example:"test@example.com"`    // 邮箱
	Avatar    string `json:"avatar" example:"https://example.com/avatar.jpg"`               // 头像URL
	Signature string `json:"signature" example:"这是我的个性签名"`                                  // 个性签名
	Readme    string `json:"readme" example:"## 关于我\n这是我的README内容"`                         // README内容
}

// UserStatusUpdateRequest 更新用户状态请求体
type UserStatusUpdateRequest struct {
	ID         int    `json:"id" binding:"required" example:"1"`                                                                  // 用户ID
	Status     string `json:"status" binding:"required,oneof=Normal Mute Blocked ActivationPending RiskControl" example:"Normal"` // 用户状态
	Reason     string `json:"reason" example:"违反社区规则"`                                                                            // 操作原因
	OperatorID int    `json:"-"`                                                                                                  // 操作者ID（内部使用）
}

// UserRoleUpdateRequest 更新用户身份请求体
type UserRoleUpdateRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"`                                            // 用户ID
	Role   string `json:"role" binding:"required,oneof=User Moderator Admin SuperAdmin" example:"User"` // 用户身份
	Reason string `json:"reason" example:"权限调整"`                                                        // 操作原因
}

// UserPointsUpdateRequest 更新用户积分请求体
type UserPointsUpdateRequest struct {
	ID     int    `json:"id" binding:"required" example:"1"`       // 用户ID
	Points int    `json:"points" binding:"required" example:"100"` // 积分变化量（正数为增加，负数为减少）
	Reason string `json:"reason" example:"活跃奖励"`                   // 操作原因
}

// UserCurrencyUpdateRequest 更新用户货币请求体
type UserCurrencyUpdateRequest struct {
	ID       int    `json:"id" binding:"required" example:"1"`        // 用户ID
	Currency int    `json:"currency" binding:"required" example:"50"` // 货币变化量（正数为增加，负数为减少）
	Reason   string `json:"reason" example:"发帖奖励"`                    // 操作原因
}

// ModeratorCategoryRequest 设置版主管理版块请求体
type ModeratorCategoryRequest struct {
	UserID      int    `json:"user_id" binding:"required" example:"1"` // 用户ID
	CategoryIDs []int  `json:"category_ids" binding:"required"`        // 版块ID列表
	Reason      string `json:"reason" example:"版主任命"`                  // 操作原因
}

// UserListItem 用户列表项响应体
type UserListItem struct {
	ID            int    `json:"id" example:"1"`                                  // 用户ID
	Username      string `json:"username" example:"testuser"`                     // 用户名
	Email         string `json:"email" example:"test@example.com"`                // 邮箱
	Avatar        string `json:"avatar" example:"https://example.com/avatar.jpg"` // 头像URL
	Signature     string `json:"signature" example:"这是我的个性签名"`                    // 个性签名
	EmailVerified bool   `json:"email_verified" example:"true"`                   // 邮箱是否已验证
	Points        int    `json:"points" example:"1000"`                           // 积分
	Currency      int    `json:"currency" example:"500"`                          // 货币
	PostCount     int    `json:"post_count" example:"50"`                         // 帖子数
	CommentCount  int    `json:"comment_count" example:"200"`                     // 评论数
	Status        string `json:"status" example:"Normal"`                         // 用户状态
	Role          string `json:"role" example:"User"`                             // 用户身份
	CreatedAt     string `json:"created_at" example:"2024-01-01T00:00:00Z"`       // 创建时间
	UpdatedAt     string `json:"updated_at" example:"2024-01-01T00:00:00Z"`       // 更新时间
}

// UserListResponse 用户列表响应体
type UserListResponse struct {
	List     []UserListItem `json:"list"`      // 用户列表
	Total    int64          `json:"total"`     // 总数量
	Page     int            `json:"page"`      // 当前页码
	PageSize int            `json:"page_size"` // 每页数量
}

// UserDetailResponse 用户详情响应体
type UserDetailResponse struct {
	ID                int                 `json:"id" example:"1"`                                  // 用户ID
	Username          string              `json:"username" example:"testuser"`                     // 用户名
	Email             string              `json:"email" example:"test@example.com"`                // 邮箱
	Avatar            string              `json:"avatar" example:"https://example.com/avatar.jpg"` // 头像URL
	Signature         string              `json:"signature" example:"这是我的个性签名"`                    // 个性签名
	Readme            string              `json:"readme" example:"## 关于我\n这是我的README内容"`           // README内容
	EmailVerified     bool                `json:"email_verified" example:"true"`                   // 邮箱是否已验证
	Points            int                 `json:"points" example:"1000"`                           // 积分
	Currency          int                 `json:"currency" example:"500"`                          // 货币
	PostCount         int                 `json:"post_count" example:"50"`                         // 帖子数
	CommentCount      int                 `json:"comment_count" example:"200"`                     // 评论数
	Status            string              `json:"status" example:"Normal"`                         // 用户状态
	Role              string              `json:"role" example:"User"`                             // 用户身份
	CreatedAt         string              `json:"created_at" example:"2024-01-01T00:00:00Z"`       // 创建时间
	UpdatedAt         string              `json:"updated_at" example:"2024-01-01T00:00:00Z"`       // 更新时间
	ManagedCategories []CategoryBasicInfo `json:"managed_categories"`                              // 管理的版块列表（仅版主显示）
}

// CategoryBasicInfo 版块基本信息
type CategoryBasicInfo struct {
	ID   int    `json:"id" example:"1"`         // 版块ID
	Name string `json:"name" example:"综合讨论"`    // 版块名称
	Slug string `json:"slug" example:"general"` // 版块标识
}

// UserBalanceLogRequest 用户余额变动记录查询请求体
type UserBalanceLogRequest struct {
	Page        int    `form:"page" binding:"required,min=1" example:"1"`               // 页码
	PageSize    int    `form:"page_size" binding:"required,min=1,max=100" example:"20"` // 每页数量
	UserID      int    `form:"user_id" example:"1"`                                     // 用户ID筛选
	Type        string `form:"type" example:"points"`                                   // 变动类型筛选
	StartDate   string `form:"start_date" example:"2024-01-01"`                         // 开始日期
	EndDate     string `form:"end_date" example:"2024-12-31"`                           // 结束日期
	OperatorID  int    `form:"operator_id" example:"2"`                                 // 操作者ID筛选
	RelatedType string `form:"related_type" example:"post"`                             // 关联业务类型筛选
}

// UserBalanceLogItem 用户余额变动记录项
type UserBalanceLogItem struct {
	ID           int    `json:"id" example:"1"`                            // 记录ID
	UserID       int    `json:"user_id" example:"1"`                       // 用户ID
	Username     string `json:"username" example:"testuser"`               // 用户名
	Type         string `json:"type" example:"points"`                     // 变动类型
	Amount       int    `json:"amount" example:"100"`                      // 变动数量
	BeforeAmount int    `json:"before_amount" example:"1000"`              // 变动前数量
	AfterAmount  int    `json:"after_amount" example:"1100"`               // 变动后数量
	Reason       string `json:"reason" example:"发帖奖励"`                     // 变动原因
	OperatorID   int    `json:"operator_id" example:"2"`                   // 操作者ID
	OperatorName string `json:"operator_name" example:"admin"`             // 操作者用户名
	RelatedID    int    `json:"related_id" example:"123"`                  // 关联业务ID
	RelatedType  string `json:"related_type" example:"post"`               // 关联业务类型
	IPAddress    string `json:"ip_address" example:"192.168.1.1"`          // IP地址
	CreatedAt    string `json:"created_at" example:"2024-01-01T00:00:00Z"` // 创建时间
}

// UserBalanceLogResponse 用户余额变动记录响应体
type UserBalanceLogResponse struct {
	List     []UserBalanceLogItem `json:"list"`      // 记录列表
	Total    int64                `json:"total"`     // 总数量
	Page     int                  `json:"page"`      // 当前页码
	PageSize int                  `json:"page_size"` // 每页数量
}

// UserBalanceSummary 用户余额汇总信息
type UserBalanceSummary struct {
	UserID           int    `json:"user_id" example:"1"`              // 用户ID
	Username         string `json:"username" example:"testuser"`      // 用户名
	CurrentPoints    int    `json:"current_points" example:"1000"`    // 当前积分
	CurrentCurrency  int    `json:"current_currency" example:"500"`   // 当前货币
	TotalPointsIn    int    `json:"total_points_in" example:"1500"`   // 总积分收入
	TotalPointsOut   int    `json:"total_points_out" example:"500"`   // 总积分支出
	TotalCurrencyIn  int    `json:"total_currency_in" example:"800"`  // 总货币收入
	TotalCurrencyOut int    `json:"total_currency_out" example:"300"` // 总货币支出
}

// UserBanRequest 用户封禁请求体
type UserBanRequest struct {
	ID         int    `json:"id" binding:"required" example:"1"`       // 用户ID
	Duration   int64  `json:"duration" binding:"gte=0" example:"3600"` // 封禁时长（秒），0表示永久封禁
	Reason     string `json:"reason" example:"违反社区规则"`                 // 封禁原因
	OperatorID int    `json:"-"`                                       // 操作者ID（内部使用）
}

// UserUnbanRequest 用户解封请求体
type UserUnbanRequest struct {
	ID         int    `json:"id" binding:"required" example:"1"` // 用户ID
	Reason     string `json:"reason" example:"申诉通过"`             // 解封原因
	OperatorID int    `json:"-"`                                 // 操作者ID（内部使用）
}
