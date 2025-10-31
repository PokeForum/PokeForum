package schema

// EmailSMTPConfigRequest SMTP配置请求体
// 用于接收前端提交的邮箱SMTP服务配置信息
type EmailSMTPConfigRequest struct {
	// 是否启用邮箱服务，true表示启用，false表示禁用
	IsEnable bool `json:"is_enable" binding:"required" example:"true"`
	// 发件人名称，显示在邮件发件人处
	Sender string `json:"sender" binding:"required,min=1,max=100" example:"PokeForum"`
	// 发件人邮箱地址，用于SMTP认证和邮件发送
	Address string `json:"address" binding:"required,email" example:"noreply@example.com"`
	// SMTP服务器主机名或IP地址
	Host string `json:"host" binding:"required,min=1,max=255" example:"smtp.example.com"`
	// SMTP服务器端口号，常见端口为25、587、465等
	Port int `json:"port" binding:"required,min=1,max=65535" example:"587"`
	// SMTP用户名，通常为邮箱地址或账户名
	Username string `json:"username" binding:"required,min=1,max=255" example:"user@example.com"`
	// SMTP密码或授权码，用于SMTP认证
	Password string `json:"password" binding:"required,min=1" example:"password123"`
	// 是否强制使用SSL加密连接，true表示使用SSL，false表示不使用
	ForcedSSL bool `json:"forced_ssl" example:"false"`
	// SMTP连接有效期（单位：秒），长时间无邮件发送时自动断开连接
	ConnectionValidity int `json:"connection_validity" binding:"required,min=10,max=3600" example:"300"`
}

// EmailSMTPConfigResponse SMTP配置响应体
// 返回当前系统的SMTP邮箱服务配置信息
type EmailSMTPConfigResponse struct {
	// 是否启用邮箱服务
	IsEnable bool `json:"is_enable" example:"true"`
	// 发件人名称
	Sender string `json:"sender" example:"PokeForum"`
	// 发件人邮箱地址
	Address string `json:"address" example:"noreply@example.com"`
	// SMTP服务器主机名
	Host string `json:"host" example:"smtp.example.com"`
	// SMTP服务器端口号
	Port int `json:"port" example:"587"`
	// SMTP用户名
	Username string `json:"username" example:"user@example.com"`
	// SMTP密码
	Password string `json:"password" example:"password123"`
	// 是否强制使用SSL加密连接
	ForcedSSL bool `json:"forced_ssl" example:"false"`
	// SMTP连接有效期（单位：秒）
	ConnectionValidity int `json:"connection_validity" example:"300"`
}

// EmailTestRequest 发送测试邮件请求体
// 用于接收测试邮件的收件人邮箱地址
type EmailTestRequest struct {
	// 收件人邮箱地址，用于接收测试邮件
	ToEmail string `json:"to_email" binding:"required,email" example:"test@example.com"`
}

// EmailTestResponse 发送测试邮件响应体
// 返回测试邮件发送的结果信息
type EmailTestResponse struct {
	// 是否发送成功，true表示成功，false表示失败
	Success bool `json:"success" example:"true"`
	// 提示信息，包含发送结果的详细说明
	Message string `json:"message" example:"测试邮件已发送"`
}
