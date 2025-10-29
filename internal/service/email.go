package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/settings"
	_const "github.com/PokeForum/PokeForum/internal/const"
	"github.com/PokeForum/PokeForum/internal/pkg/email"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/gomodule/redigo/redis"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

// IEmailService 邮箱服务接口
type IEmailService interface {
	// GetSMTPConfig 获取SMTP配置
	GetSMTPConfig(ctx context.Context) (*schema.EmailSMTPConfigResponse, error)
	// UpdateSMTPConfig 更新SMTP配置
	UpdateSMTPConfig(ctx context.Context, req schema.EmailSMTPConfigRequest) error
	// SendTestEmail 发送测试邮件
	SendTestEmail(ctx context.Context, toEmail string) error
}

// EmailService 邮箱服务实现
type EmailService struct {
	db     *ent.Client
	cache  *redis.Pool
	logger *zap.Logger
}

// NewEmailService 创建邮箱服务实例
func NewEmailService(db *ent.Client, cache *redis.Pool, logger *zap.Logger) IEmailService {
	return &EmailService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// GetSMTPConfig 获取SMTP配置
// 从数据库查询邮箱服务的SMTP配置信息，返回完整的配置对象
func (s *EmailService) GetSMTPConfig(ctx context.Context) (*schema.EmailSMTPConfigResponse, error) {
	// 查询所有邮箱相关的配置，模块为Function的所有配置项
	configs, err := s.db.Settings.Query().
		Where(settings.ModuleEQ("Function")).
		All(ctx)
	if err != nil {
		s.logger.Error("查询邮箱配置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return nil, fmt.Errorf("查询邮箱配置失败: %w", err)
	}

	// 将配置列表转换为map便于按key快速查询
	configMap := make(map[string]string)
	for _, cfg := range configs {
		if cfg.Key != "" {
			configMap[cfg.Key] = cfg.Value
		}
	}

	// 构建响应对象，初始化为空值
	resp := &schema.EmailSMTPConfigResponse{}

	// 解析是否启用邮箱服务，从configMap中获取对应的值
	if isEnable, ok := configMap[_const.EmailIsEnableEmailService]; ok {
		resp.IsEnable = isEnable == "true"
	}

	// 解析发件人名称
	if sender, ok := configMap[_const.EmailSender]; ok {
		resp.Sender = sender
	}

	// 解析发件人邮箱地址
	if address, ok := configMap[_const.EmailAddress]; ok {
		resp.Address = address
	}

	// 解析SMTP服务器主机名
	if host, ok := configMap[_const.EmailHost]; ok {
		resp.Host = host
	}

	// 解析SMTP服务器端口，需要将字符串转换为整数
	if port, ok := configMap[_const.EmailPort]; ok {
		if p, err := strconv.Atoi(port); err == nil {
			resp.Port = p
		}
	}

	// 解析SMTP用户名
	if username, ok := configMap[_const.EmailUsername]; ok {
		resp.Username = username
	}

	// 解析是否强制使用SSL加密连接
	if forcedSSL, ok := configMap[_const.EmailForcedSSL]; ok {
		resp.ForcedSSL = forcedSSL == "true"
	}

	// 解析SMTP连接有效期（单位：秒），需要将字符串转换为整数
	if validity, ok := configMap[_const.EmailConnectionValidity]; ok {
		if v, err := strconv.Atoi(validity); err == nil {
			resp.ConnectionValidity = v
		}
	}

	return resp, nil
}

// UpdateSMTPConfig 更新SMTP配置
// 将SMTP配置保存到数据库，使用upsert操作确保配置存在，不存在则创建
func (s *EmailService) UpdateSMTPConfig(ctx context.Context, req schema.EmailSMTPConfigRequest) error {
	// 定义需要更新的配置项，将所有配置转换为字符串格式存储
	configItems := map[string]string{
		_const.EmailIsEnableEmailService: fmt.Sprintf("%v", req.IsEnable),
		_const.EmailSender:               req.Sender,
		_const.EmailAddress:              req.Address,
		_const.EmailHost:                 req.Host,
		_const.EmailPort:                 strconv.Itoa(req.Port),
		_const.EmailUsername:             req.Username,
		_const.EmailPassword:             req.Password,
		_const.EmailForcedSSL:            fmt.Sprintf("%v", req.ForcedSSL),
		_const.EmailConnectionValidity:   strconv.Itoa(req.ConnectionValidity),
	}

	// 遍历配置项，逐个更新或创建，确保所有配置都被正确保存
	for key, value := range configItems {
		// 先尝试查询是否存在该配置，使用module和key的组合查询
		existing, err := s.db.Settings.Query().
			Where(
				settings.ModuleEQ("Function"),
				settings.KeyEQ(key),
			).
			First(ctx)

		if err != nil && !ent.IsNotFound(err) {
			return fmt.Errorf("查询配置失败: %w", err)
		}

		// 如果配置存在，则更新；否则创建新配置
		if existing != nil {
			// 更新现有配置的值
			if _, err := s.db.Settings.UpdateOne(existing).
				SetValue(value).
				Save(ctx); err != nil {
				s.logger.Error("更新配置邮箱配置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
				return fmt.Errorf("更新配置 %s 失败: %w", key, err)
			}
		} else {
			// 创建新配置，设置module、key、value和valueType
			if _, err := s.db.Settings.Create().
				SetModule("Function").
				SetKey(key).
				SetValue(value).
				SetValueType("string").
				Save(ctx); err != nil {
				s.logger.Error("创建配置邮箱配置失败", tracing.WithTraceIDField(ctx), zap.Error(err))
				return fmt.Errorf("创建配置 %s 失败: %w", key, err)
			}
		}
	}

	return nil
}

// SendTestEmail 发送测试邮件
// 使用当前配置发送一封测试邮件到指定邮箱，用于验证SMTP配置是否正确
func (s *EmailService) SendTestEmail(ctx context.Context, toEmail string) error {
	// 获取当前SMTP配置
	config, err := s.GetSMTPConfig(ctx)
	if err != nil {
		return fmt.Errorf("获取SMTP配置失败: %w", err)
	}

	// 检查邮箱服务是否启用
	if !config.IsEnable {
		return errors.New("邮箱服务未启用")
	}

	// 检查必要的配置是否完整，确保主机、端口和用户名都已配置
	if config.Host == "" || config.Port == 0 || config.Username == "" {
		return errors.New("SMTP配置不完整")
	}

	// 获取密码（从数据库查询，密码不在GetSMTPConfig返回的对象中）
	passwordCfg, err := s.db.Settings.Query().
		Where(
			settings.ModuleEQ("Function"),
			settings.KeyEQ(_const.EmailPassword),
		).
		First(ctx)
	if err != nil {
		s.logger.Error("获取邮箱密码失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return fmt.Errorf("获取邮箱密码失败: %w", err)
	}

	// 创建SMTP配置对象，用于建立SMTP连接
	smtpConfig := email.SMTPConfig{
		Name:       config.Sender,
		Address:    config.Address,
		Host:       config.Host,
		Port:       config.Port,
		User:       config.Username,
		Password:   passwordCfg.Value,
		Encryption: config.ForcedSSL,
		Keepalive:  config.ConnectionValidity,
	}

	// 创建邮件消息对象
	m := mail.NewMsg()
	if err = m.FromFormat(smtpConfig.Name, smtpConfig.Address); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人邮箱地址
	if err = m.To(toEmail); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	// 设置邮件主题
	m.Subject("PokeForum 邮箱服务测试")
	m.SetMessageID()

	// 设置邮件内容为HTML格式
	htmlBody := `
	<html>
		<body>
			<h2>邮箱服务测试</h2>
			<p>这是来自 PokeForum 的测试邮件。</p>
			<p>如果您收到此邮件，说明邮箱服务配置成功。</p>
			<hr>
			<p>发送时间: <strong>` + fmt.Sprintf("%v", ctx.Value("timestamp")) + `</strong></p>
		</body>
	</html>
	`
	m.SetBodyString(mail.TypeTextHTML, htmlBody)

	// 建立SMTP连接并发送邮件，配置SMTP选项
	opts := []mail.Option{
		mail.WithPort(smtpConfig.Port),
		mail.WithTimeout(30),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithTLSPortPolicy(mail.TLSOpportunistic),
		mail.WithUsername(smtpConfig.User),
		mail.WithPassword(smtpConfig.Password),
	}

	// 如果启用SSL加密，添加SSL选项
	if smtpConfig.Encryption {
		opts = append(opts, mail.WithSSL())
	}

	// 创建SMTP客户端
	client, err := mail.NewClient(smtpConfig.Host, opts...)
	if err != nil {
		s.logger.Error("创建SMTP客户端失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}

	// 连接到SMTP服务器并发送邮件
	if err = client.DialAndSend(m); err != nil {
		s.logger.Error("发送邮件失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}
