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

// SMTPPool SMTP协议发送邮件
type SMTPPool struct {
	config SMTPConfig
	ch     chan *message // 队列
	chOpen bool          // 队列状态
	l      zap.Logger    // 日志
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

type message struct {
	msg     *mail.Msg
	to      string
	subject string
	cid     string // 任务ID
	userID  int
}

// NewSMTPPool 初始化一个新的基于 SMTP 的电子邮件发送队列。
func NewSMTPPool(config SMTPConfig, logger zap.Logger) *SMTPPool {
	client := &SMTPPool{
		config: config,
		ch:     make(chan *message, 30),
		chOpen: false,
		l:      logger,
	}

	client.Init()
	return client
}

// Send 发送邮件
func (client *SMTPPool) Send(ctx context.Context, to, title, body string) error {
	if !client.chOpen {
		return fmt.Errorf("SMTP pool is closed")
	}

	// 忽略通过QQ登录的邮箱
	if strings.HasSuffix(to, "@login.qq.com") {
		return nil
	}

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
	
	// 从context中获取链路ID和用户ID
	cid := tracing.GetTraceID(ctx)
	userID := tracing.GetUserID(ctx)
	
	client.ch <- &message{
		msg:     m,
		subject: title,
		to:      to,
		cid:     cid,
		userID:  userID,
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
		client.l.Info("初始化并启动 SMTP 邮件队列...")
		defer func() {
			if err := recover(); err != nil {
				client.chOpen = false
				client.l.Error("发送邮件异常，队列将在10秒后重置", zap.Any("error", err))
				time.Sleep(time.Duration(10) * time.Second)
				client.Init()
			}
		}()

		opts := []mail.Option{
			mail.WithPort(client.config.Port),
			mail.WithTimeout(time.Duration(client.config.Keepalive+5) * time.Second),
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover), mail.WithTLSPortPolicy(mail.TLSOpportunistic),
			mail.WithUsername(client.config.User), mail.WithPassword(client.config.Password),
		}
		if client.config.Encryption {
			opts = append(opts, mail.WithSSL())
		}

		d, diaErr := mail.NewClient(client.config.Host, opts...)
		if diaErr != nil {
			client.l.Panic("创建 SMTP 客户端失败", zap.Error(diaErr))
			return
		}

		client.chOpen = true

		var err error
		open := false
		for {
			select {
			case m, ok := <-client.ch:
				if !ok {
					client.l.Info("邮件队列关闭中...")
					client.chOpen = false
					return
				}

				if !open {
					if err = d.DialWithContext(context.Background()); err != nil {
						panic(err)
					}
					open = true
				}

				if err = d.Send(m.msg); err != nil {
					// 检查是否为成功发送后的 SMTP RESET 错误
					var sendErr *mail.SendError
					if errors.As(err, &sendErr) && sendErr.Reason == mail.ErrSMTPReset {
						open = false
						client.l.Debug("SMTP RESET 错误，关闭连接")
						// https://github.com/wneessen/go-mail/issues/463
						continue // 邮件已发送，不视为发送失败
					}

					client.l.Warn("邮件发送失败", 
						zap.String("to", m.to), 
						zap.String("cid", m.cid), 
						zap.Int("userID", m.userID),
						zap.Error(err))
				} else {
					client.l.Info("邮件发送成功", 
						zap.String("to", m.to), 
						zap.String("subject", m.subject),
						zap.String("cid", m.cid),
						zap.Int("userID", m.userID))
				}
			// 长时间没有新邮件，则关闭 SMTP 连接
			case <-time.After(time.Duration(client.config.Keepalive) * time.Second):
				if open {
					if err = d.Close(); err != nil {
						client.l.Warn("关闭 SMTP 连接失败", zap.Error(err))
					}
					open = false
				}
			}
		}
	}()
}
