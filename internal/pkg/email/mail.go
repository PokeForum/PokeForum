package email

import "context"

// Driver 邮件驱动
type Driver interface {
	// Send 发送邮件
	Send(ctx context.Context, to, title, body string) error
	// Close 关闭驱动
	Close()
}
