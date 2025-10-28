package schema

// RegisterRequest 用户注册请求体
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100" example:"testuser"`
	Email    string `json:"email" binding:"required,email" example:"test@example.com"`
	Password string `json:"password" binding:"required,min=8,max=24" example:"password123"`
}

// LoginRequest 用户登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100" example:"testuser"`
	Password string `json:"password" binding:"required,min=8,max=24" example:"password123"`
}

// UserResponse 用户响应体
type UserResponse struct {
	ID       int    `json:"id" example:"1"`
	Username string `json:"username" example:"testuser"`
	Email    string `json:"email" example:"test@example.com"`
}
