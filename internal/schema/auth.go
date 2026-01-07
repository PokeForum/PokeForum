package schema

// RegisterRequest User registration request | 用户注册请求体
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100" example:"testuser"` // Username | 用户名
	Email    string `json:"email" binding:"required,email" example:"test@example.com"`    // Email address | 邮箱
	Password string `json:"password" binding:"required,min=8" example:"password123"`      // Password | 密码
}

// LoginRequest User login request | 用户登录请求体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"test@example.com"` // Email address | 邮箱
	Password string `json:"password" binding:"required,min=8" example:"password123"`   // Password | 密码
}

// LoginResponse User login response | 用户登录响应体
type LoginResponse struct {
	Token    string `json:"token"`    // Authentication token | Token
	ID       int    `json:"id"`       // User ID | 用户ID
	Username string `json:"username"` // Username | 用户名
}

// UserResponse User response | 用户响应体
type UserResponse struct {
	ID       int    `json:"id" example:"1"`                   // User ID | 用户ID
	Username string `json:"username" example:"testuser"`      // Username | 用户名
	Email    string `json:"email" example:"test@example.com"` // User email | 用户邮箱
}

// ForgotPasswordRequest Forgot password request (send verification code) | 忘记密码请求体（发送验证码）
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"test@example.com"` // Email address | 邮箱
}

// ForgotPasswordResponse Forgot password response | 忘记密码响应体
type ForgotPasswordResponse struct {
	Sent      bool   `json:"sent" example:"true"`               // Verification code sent status | 验证码发送状态
	Message   string `json:"message" example:"验证码已发送到您的邮箱，请查收"` // Notification message | 提示信息
	ExpiresIn int    `json:"expires_in" example:"600"`          // Verification code expiry time (seconds) | 验证码有效期（秒）
}

// ResetPasswordRequest Reset password request | 重置密码请求体
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email" example:"test@example.com"`  // Email address | 邮箱
	Code        string `json:"code" binding:"required,len=6" example:"123456"`             // Verification code | 验证码
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newpass123"` // New password | 新密码
}

// ResetPasswordResponse Reset password response | 重置密码响应体
type ResetPasswordResponse struct {
	Success bool   `json:"success" example:"true"`   // Whether successful | 是否成功
	Message string `json:"message" example:"密码重置成功"` // Notification message | 提示信息
}
