package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

// message 邮件消息
type message struct {
	msg     *mail.Msg
	to      string
	subject string
	traceID string
	userID  int
}

// SMTPPool SMTP协议发送邮件（基于channel队列）
type SMTPPool struct {
	config SMTPConfig
	ch     chan *message
	chOpen bool
	ready  chan struct{} // 初始化完成信号
	logger *zap.Logger
}

// SMTPConfig SMTP发送配置
type SMTPConfig struct {
	Name       string // 发送者名
	Address    string // 发送者地址
	Host       string // 服务器主机名
	Port       int    // 服务器端口
	User       string // 用户名
	Password   string // 密码
	Encryption bool   // 是否启用加密
	Keepalive  int    // SMTP 连接保留时长
}

// NewSMTPPool 初始化一个新的SMTP邮件发送队列
func NewSMTPPool(config SMTPConfig, logger *zap.Logger) *SMTPPool {
	client := &SMTPPool{
		config: config,
		ch:     make(chan *message, 30),
		chOpen: false,
		ready:  make(chan struct{}),
		logger: logger,
	}

	client.Init()

	// 等待初始化完成（最多5秒）
	select {
	case <-client.ready:
	case <-time.After(5 * time.Second):
		logger.Warn("SMTP队列初始化超时")
	}

	return client
}

// Send 发送邮件（提交到channel队列）
func (client *SMTPPool) Send(ctx context.Context, to, title, body string) error {
	if !client.chOpen {
		return fmt.Errorf("SMTP pool is closed")
	}

	// 忽略通过QQ登录的邮箱
	if strings.HasSuffix(to, "@login.qq.com") {
		return nil
	}

	// 创建邮件消息
	m := mail.NewMsg()
	if err := m.FromFormat(client.config.Name, client.config.Address); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject(title)
	m.SetMessageID()
	m.SetBodyString(mail.TypeTextHTML, body)

	// 提交到队列
	client.ch <- &message{
		msg:     m,
		subject: title,
		to:      to,
		traceID: tracing.GetTraceID(ctx),
		userID:  tracing.GetUserID(ctx),
	}
	return nil
}

// Close 关闭发送队列
func (client *SMTPPool) Close() {
	if client.ch != nil {
		close(client.ch)
	}
}

// Init 初始化发送队列
func (client *SMTPPool) Init() {
	go func() {
		client.logger.Info("初始化并启动SMTP邮件队列...")
		defer func() {
			if err := recover(); err != nil {
				client.chOpen = false
				client.logger.Error("邮件发送异常，队列将在10秒后重置", zap.Any("error", err))
				time.Sleep(10 * time.Second)
				client.Init()
			}
		}()

		// 创建SMTP客户端选项
		opts := []mail.Option{
			mail.WithPort(client.config.Port),
			mail.WithTimeout(time.Duration(client.config.Keepalive+5) * time.Second),
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithTLSPortPolicy(mail.TLSOpportunistic),
			mail.WithUsername(client.config.User),
			mail.WithPassword(client.config.Password),
		}
		if client.config.Encryption {
			opts = append(opts, mail.WithSSL())
		}

		// 创建SMTP客户端
		d, diaErr := mail.NewClient(client.config.Host, opts...)
		if diaErr != nil {
			client.logger.Panic("创建SMTP客户端失败", zap.Error(diaErr))
			return
		}

		client.chOpen = true
		close(client.ready) // 通知初始化完成

		var err error
		open := false
		for {
			select {
			case m, ok := <-client.ch:
				if !ok {
					client.logger.Info("邮件队列关闭中...")
					client.chOpen = false
					return
				}

				// 按需建立连接
				if !open {
					if err = d.DialWithContext(context.Background()); err != nil {
						panic(err)
					}
					open = true
				}

				// 发送邮件
				if err := d.Send(m.msg); err != nil {
					// 检查是否为SMTP RESET错误（邮件已发送成功）
					var sendErr *mail.SendError
					if errors.As(err, &sendErr) && sendErr.Reason == mail.ErrSMTPReset {
						open = false
						client.logger.Debug("SMTP RESET错误，关闭连接...",
							zap.String("traceId", m.traceID))
						continue
					}

					client.logger.Warn("邮件发送失败",
						zap.String("to", m.to),
						zap.String("subject", m.subject),
						zap.String("traceId", m.traceID),
						zap.Int("userId", m.userID),
						zap.Error(err))
				} else {
					client.logger.Info("邮件发送成功",
						zap.String("to", m.to),
						zap.String("subject", m.subject),
						zap.String("traceId", m.traceID),
						zap.Int("userId", m.userID))
				}

			// 长时间没有新邮件，则关闭SMTP连接
			case <-time.After(time.Duration(client.config.Keepalive) * time.Second):
				if open {
					if err := d.Close(); err != nil {
						client.logger.Warn("关闭SMTP连接失败", zap.Error(err))
					}
					open = false
				}
			}
		}
	}()
}
