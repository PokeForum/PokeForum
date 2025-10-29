package _const

// 邮箱服务
const (
	// EmailIsEnableEmailService 是否启用邮箱服务
	EmailIsEnableEmailService = "email:is_enable_email_service"
	// EmailSender 发件人
	EmailSender = "email:sender"
	// EmailAddress 发件人邮箱
	EmailAddress = "email:address"
	// EmailHost SMTP 服务器
	EmailHost = "email:host"
	// EmailPort SMTP 端口
	EmailPort = "email:port"
	// EmailUsername SMTP 用户名
	EmailUsername = "email:username"
	// EmailPassword 发件人密码
	EmailPassword = "email:password"
	// EmailForcedSSL 强制使用 SSL 连接
	EmailForcedSSL = "email:forced_ssl"
	// EmailConnectionValidity SMTP 连接有效期 (秒)
	EmailConnectionValidity = "email:connection_validity"

	// EmailAccountActivationTemplate 账户激活模板
	EmailAccountActivationTemplate = "email:account_activation_template"
	// EmailPasswordResetTemplate 重置密码模板
	EmailPasswordResetTemplate = "email:password_reset_template"
)
