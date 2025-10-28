package schema

// RegisterRequest 用户注册请求体
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100" example:"testuser"` // 用户名
	Email    string `json:"email" binding:"required,email" example:"test@example.com"`    // 邮箱
	Password string `json:"password" binding:"required,min=8" example:"password123"`      // 密码
}

// LoginRequest 用户登录请求体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"test@example.com"` // 邮箱
	Password string `json:"password" binding:"required,min=8" example:"password123"`   // 密码
}

// LoginResponse 用户登录响应体
type LoginResponse struct {
	Token    string `json:"token"`    // Token
	ID       int    `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
}

// UserResponse 用户响应体
type UserResponse struct {
	ID       int    `json:"id" example:"1"`                   // 用户ID
	Username string `json:"username" example:"testuser"`      // 用户名
	Email    string `json:"email" example:"test@example.com"` // 用户邮箱
}
