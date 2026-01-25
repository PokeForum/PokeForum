package email

import "context"

// Driver Email driver | 邮件驱动
type Driver interface {
	// Send Send email | 发送邮件
	Send(ctx context.Context, to, title, body string) error
	// Close Close driver | 关闭驱动
	Close()
}
