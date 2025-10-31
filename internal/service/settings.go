package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/ent/settings"
	_const "github.com/PokeForum/PokeForum/internal/const"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/pkg/email"
	"github.com/PokeForum/PokeForum/internal/pkg/tracing"
	"github.com/PokeForum/PokeForum/internal/schema"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

// ISettingsService 设置服务接口
type ISettingsService interface {
	// GetRoutineSettings 常规设置
	GetRoutineSettings(ctx context.Context) (*schema.RoutineSettingsResponse, error)
	UpdateRoutineSettings(ctx context.Context, req schema.RoutineSettingsRequest) error

	// GetHomeSettings 首页设置
	GetHomeSettings(ctx context.Context) (*schema.HomeSettingsResponse, error)
	UpdateHomeSettings(ctx context.Context, req schema.HomeSettingsRequest) error

	// GetCommentSettings 评论设置
	GetCommentSettings(ctx context.Context) (*schema.CommentSettingsResponse, error)
	UpdateCommentSettings(ctx context.Context, req schema.CommentSettingsRequest) error

	// GetSeoSettings SEO设置
	GetSeoSettings(ctx context.Context) (*schema.SeoSettingsResponse, error)
	UpdateSeoSettings(ctx context.Context, req schema.SeoSettingsRequest) error

	// GetCodeSettings 代码配置
	GetCodeSettings(ctx context.Context) (*schema.CodeSettingsResponse, error)
	UpdateCodeSettings(ctx context.Context, req schema.CodeSettingsRequest) error

	// GetSafeSettings 安全设置
	GetSafeSettings(ctx context.Context) (*schema.SafeSettingsResponse, error)
	UpdateSafeSettings(ctx context.Context, req schema.SafeSettingsRequest) error

	// GetSMTPConfig 邮箱设置
	GetSMTPConfig(ctx context.Context) (*schema.EmailSMTPConfigResponse, error)
	UpdateSMTPConfig(ctx context.Context, req schema.EmailSMTPConfigRequest) error
	SendTestEmail(ctx context.Context, toEmail string) error

	// GetSettingByKey 根据key获取设置值 - 公共方法
	GetSettingByKey(ctx context.Context, key string, defaultValue string) (string, error)

	// ClearSettingCache 清理指定设置的缓存 - 公共方法
	ClearSettingCache(key string)
}

// SettingsService 设置服务实现
type SettingsService struct {
	db     *ent.Client
	cache  cache.ICacheService
	logger *zap.Logger
}

// NewSettingsService 创建设置服务实例
func NewSettingsService(db *ent.Client, cacheService cache.ICacheService, logger *zap.Logger) ISettingsService {
	return &SettingsService{
		db:     db,
		cache:  cacheService,
		logger: logger,
	}
}

// getSettingsByModule 通用方法：根据模块获取配置
func (s *SettingsService) getSettingsByModule(ctx context.Context, module settings.Module) (map[string]string, error) {
	configs, err := s.db.Settings.Query().
		Where(settings.ModuleEQ(module)).
		All(ctx)
	if err != nil {
		s.logger.Error("查询配置失败", tracing.WithTraceIDField(ctx), zap.String("module", module.String()), zap.Error(err))
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}

	configMap := make(map[string]string)
	for _, cfg := range configs {
		if cfg.Key != "" {
			configMap[cfg.Key] = cfg.Value
		}
	}

	return configMap, nil
}

// upsertSetting 通用方法：更新或插入单个配置项
func (s *SettingsService) upsertSetting(ctx context.Context, module settings.Module, key, value string, valueType settings.ValueType) error {
	existing, err := s.db.Settings.Query().
		Where(
			settings.ModuleEQ(module),
			settings.KeyEQ(key),
		).
		First(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("查询配置失败: %w", err)
	}

	if existing != nil {
		// 更新现有配置
		if _, err := s.db.Settings.UpdateOne(existing).
			SetValue(value).
			Save(ctx); err != nil {
			s.logger.Error("更新配置失败", tracing.WithTraceIDField(ctx), zap.String("key", key), zap.Error(err))
			return fmt.Errorf("更新配置 %s 失败: %w", key, err)
		}
	} else {
		// 创建新配置
		if _, err := s.db.Settings.Create().
			SetModule(module).
			SetKey(key).
			SetValue(value).
			SetValueType(valueType).
			Save(ctx); err != nil {
			s.logger.Error("创建配置失败", tracing.WithTraceIDField(ctx), zap.String("key", key), zap.Error(err))
			return fmt.Errorf("创建配置 %s 失败: %w", key, err)
		}
	}

	// 清理对应的Redis缓存
	s.ClearSettingCache(key)

	return nil
}

// batchUpsertSettings 通用方法：批量更新或插入配置项
func (s *SettingsService) batchUpsertSettings(ctx context.Context, module settings.Module, configItems map[string]string) error {
	for key, value := range configItems {
		if err := s.upsertSetting(ctx, module, key, value, settings.ValueTypeString); err != nil {
			return err
		}
	}
	return nil
}

// GetRoutineSettings 获取常规设置
func (s *SettingsService) GetRoutineSettings(ctx context.Context) (*schema.RoutineSettingsResponse, error) {
	configMap, err := s.getSettingsByModule(ctx, settings.ModuleSite)
	if err != nil {
		return nil, err
	}

	resp := &schema.RoutineSettingsResponse{
		WebSiteLogo:           configMap[_const.RoutineWebSiteLogo],
		WebSiteIcon:           configMap[_const.RoutineWebSiteIcon],
		ICPRecord:             configMap[_const.RoutineICPRecord],
		PublicSecurityNetwork: configMap[_const.RoutinePublicSecurityNetwork],
		IsCloseCopyright:      configMap[_const.RoutineIsCloseCopyright] == _const.SettingBoolTrue.String(),
	}

	return resp, nil
}

// UpdateRoutineSettings 更新常规设置
func (s *SettingsService) UpdateRoutineSettings(ctx context.Context, req schema.RoutineSettingsRequest) error {
	configItems := map[string]string{
		_const.RoutineWebSiteLogo:           req.WebSiteLogo,
		_const.RoutineWebSiteIcon:           req.WebSiteIcon,
		_const.RoutineICPRecord:             req.ICPRecord,
		_const.RoutinePublicSecurityNetwork: req.PublicSecurityNetwork,
		_const.RoutineIsCloseCopyright:      strconv.FormatBool(req.IsCloseCopyright),
	}

	return s.batchUpsertSettings(ctx, settings.ModuleSite, configItems)
}

// GetHomeSettings 获取首页设置
func (s *SettingsService) GetHomeSettings(ctx context.Context) (*schema.HomeSettingsResponse, error) {
	configMap, err := s.getSettingsByModule(ctx, settings.ModuleHomePage)
	if err != nil {
		return nil, err
	}

	resp := &schema.HomeSettingsResponse{
		Slides: []schema.SlideItem{},
		Links:  []schema.LinkItem{},
	}

	// 解析幻灯片JSON数据
	if slideData, ok := configMap[_const.HomeSlide]; ok && slideData != "" {
		if err := json.Unmarshal([]byte(slideData), &resp.Slides); err != nil {
			s.logger.Error("解析幻灯片数据失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		}
	}

	// 解析友情链接JSON数据
	if linkData, ok := configMap[_const.HomeLinks]; ok && linkData != "" {
		if err := json.Unmarshal([]byte(linkData), &resp.Links); err != nil {
			s.logger.Error("解析友情链接数据失败", tracing.WithTraceIDField(ctx), zap.Error(err))
		}
	}

	return resp, nil
}

// UpdateHomeSettings 更新首页设置
func (s *SettingsService) UpdateHomeSettings(ctx context.Context, req schema.HomeSettingsRequest) error {
	configItems := make(map[string]string)

	// 序列化幻灯片数据为JSON
	if req.Slides != nil {
		slideData, err := json.Marshal(req.Slides)
		if err != nil {
			return fmt.Errorf("序列化幻灯片数据失败: %w", err)
		}
		configItems[_const.HomeSlide] = string(slideData)
	} else {
		configItems[_const.HomeSlide] = "[]"
	}

	// 序列化友情链接数据为JSON
	if req.Links != nil {
		linkData, err := json.Marshal(req.Links)
		if err != nil {
			return fmt.Errorf("序列化友情链接数据失败: %w", err)
		}
		configItems[_const.HomeLinks] = string(linkData)
	} else {
		configItems[_const.HomeLinks] = "[]"
	}

	return s.batchUpsertSettings(ctx, settings.ModuleHomePage, configItems)
}

// GetCommentSettings 获取评论设置
func (s *SettingsService) GetCommentSettings(ctx context.Context) (*schema.CommentSettingsResponse, error) {
	configMap, err := s.getSettingsByModule(ctx, settings.ModuleComment)
	if err != nil {
		return nil, err
	}

	resp := &schema.CommentSettingsResponse{
		ShowCommentInfo:  configMap[_const.CommentShowCommentInfo] == _const.SettingBoolTrue.String(),
		RequireApproval:  configMap[_const.CommentRequireApproval] == _const.SettingBoolTrue.String(),
		KeywordBlacklist: configMap[_const.CommentKeywordBlacklist],
	}

	return resp, nil
}

// UpdateCommentSettings 更新评论设置
func (s *SettingsService) UpdateCommentSettings(ctx context.Context, req schema.CommentSettingsRequest) error {
	configItems := map[string]string{
		_const.CommentShowCommentInfo:  strconv.FormatBool(req.ShowCommentInfo),
		_const.CommentRequireApproval:  strconv.FormatBool(req.RequireApproval),
		_const.CommentKeywordBlacklist: req.KeywordBlacklist,
	}

	return s.batchUpsertSettings(ctx, settings.ModuleComment, configItems)
}

// GetSeoSettings 获取SEO设置
func (s *SettingsService) GetSeoSettings(ctx context.Context) (*schema.SeoSettingsResponse, error) {
	configMap, err := s.getSettingsByModule(ctx, settings.ModuleSeo)
	if err != nil {
		return nil, err
	}

	resp := &schema.SeoSettingsResponse{
		WebSiteName:        configMap[_const.SeoWebSiteName],
		WebSiteKeyword:     configMap[_const.SeoWebSiteKeyword],
		WebSiteDescription: configMap[_const.SeoWebSiteDescription],
	}

	return resp, nil
}

// UpdateSeoSettings 更新SEO设置
func (s *SettingsService) UpdateSeoSettings(ctx context.Context, req schema.SeoSettingsRequest) error {
	configItems := map[string]string{
		_const.SeoWebSiteName:        req.WebSiteName,
		_const.SeoWebSiteKeyword:     req.WebSiteKeyword,
		_const.SeoWebSiteDescription: req.WebSiteDescription,
	}

	return s.batchUpsertSettings(ctx, settings.ModuleSeo, configItems)
}

// GetCodeSettings 获取代码配置
func (s *SettingsService) GetCodeSettings(ctx context.Context) (*schema.CodeSettingsResponse, error) {
	configMap, err := s.getSettingsByModule(ctx, settings.ModuleSite)
	if err != nil {
		return nil, err
	}

	resp := &schema.CodeSettingsResponse{
		Header:           configMap[_const.CodeHeader],
		Footer:           configMap[_const.CodeFooter],
		CustomizationCSS: configMap[_const.CodeCustomizationCSS],
	}

	return resp, nil
}

// UpdateCodeSettings 更新代码配置
func (s *SettingsService) UpdateCodeSettings(ctx context.Context, req schema.CodeSettingsRequest) error {
	configItems := map[string]string{
		_const.CodeHeader:           req.Header,
		_const.CodeFooter:           req.Footer,
		_const.CodeCustomizationCSS: req.CustomizationCSS,
	}

	return s.batchUpsertSettings(ctx, settings.ModuleSite, configItems)
}

// GetSafeSettings 获取安全设置
func (s *SettingsService) GetSafeSettings(ctx context.Context) (*schema.SafeSettingsResponse, error) {
	configMap, err := s.getSettingsByModule(ctx, settings.ModuleSecurity)
	if err != nil {
		return nil, err
	}

	resp := &schema.SafeSettingsResponse{
		IsCloseRegister:        configMap[_const.SafeIsCloseRegister] == _const.SettingBoolTrue.String(),
		IsEnableEmailWhitelist: configMap[_const.SafeIsEnableEmailWhitelist] == _const.SettingBoolTrue.String(),
		EmailWhitelist:         configMap[_const.SafeEmailWhitelist],
		VerifyEmail:            configMap[_const.SafeVerifyEmail] == _const.SettingBoolTrue.String(),
	}

	return resp, nil
}

// UpdateSafeSettings 更新安全设置
func (s *SettingsService) UpdateSafeSettings(ctx context.Context, req schema.SafeSettingsRequest) error {
	configItems := map[string]string{
		_const.SafeIsCloseRegister:        strconv.FormatBool(req.IsCloseRegister),
		_const.SafeIsEnableEmailWhitelist: strconv.FormatBool(req.IsEnableEmailWhitelist),
		_const.SafeEmailWhitelist:         req.EmailWhitelist,
		_const.SafeVerifyEmail:            strconv.FormatBool(req.VerifyEmail),
	}

	return s.batchUpsertSettings(ctx, settings.ModuleSecurity, configItems)
}

// GetSMTPConfig 获取SMTP配置
// 从数据库查询邮箱服务的SMTP配置信息，返回完整的配置对象
func (s *SettingsService) GetSMTPConfig(ctx context.Context) (*schema.EmailSMTPConfigResponse, error) {
	configMap, err := s.getSettingsByModule(ctx, settings.ModuleFunction)
	if err != nil {
		return nil, err
	}

	resp := &schema.EmailSMTPConfigResponse{}

	// 解析是否启用邮箱服务
	if isEnable, ok := configMap[_const.EmailIsEnableEmailService]; ok {
		resp.IsEnable = isEnable == _const.SettingBoolTrue.String()
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

	// 解析SMTP密码
	if password, ok := configMap[_const.EmailPassword]; ok {
		resp.Password = password
	}

	// 解析是否强制使用SSL加密连接
	if forcedSSL, ok := configMap[_const.EmailForcedSSL]; ok {
		resp.ForcedSSL = forcedSSL == _const.SettingBoolTrue.String()
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
// 将SMTP配置保存到数据库，使用upsert操作确保配置存在
func (s *SettingsService) UpdateSMTPConfig(ctx context.Context, req schema.EmailSMTPConfigRequest) error {
	configItems := map[string]string{
		_const.EmailIsEnableEmailService: strconv.FormatBool(req.IsEnable),
		_const.EmailSender:               req.Sender,
		_const.EmailAddress:              req.Address,
		_const.EmailHost:                 req.Host,
		_const.EmailPort:                 strconv.Itoa(req.Port),
		_const.EmailUsername:             req.Username,
		_const.EmailPassword:             req.Password,
		_const.EmailForcedSSL:            strconv.FormatBool(req.ForcedSSL),
		_const.EmailConnectionValidity:   strconv.Itoa(req.ConnectionValidity),
	}

	return s.batchUpsertSettings(ctx, settings.ModuleFunction, configItems)
}

// SendTestEmail 发送测试邮件
// 使用当前配置发送一封测试邮件到指定邮箱，用于验证SMTP配置是否正确
func (s *SettingsService) SendTestEmail(ctx context.Context, toEmail string) error {
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

	// 创建SMTP配置对象，用于建立SMTP连接
	smtpConfig := email.SMTPConfig{
		Name:       config.Sender,
		Address:    config.Address,
		Host:       config.Host,
		Port:       config.Port,
		User:       config.Username,
		Password:   config.Password, // 直接使用配置中的密码
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
		mail.WithTimeout(60),
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

// GetSettingByKey 根据key获取设置值 - 公共方法
// 提供给其他服务查询单个设置项的通用方法，使用Redis缓存加速查询
func (s *SettingsService) GetSettingByKey(ctx context.Context, key string, defaultValue string) (string, error) {
	// 先从Redis缓存中查询
	cacheKey := _const.GetSettingKey(key)
	cachedValue, err := s.cache.Get(cacheKey)
	if err == nil && cachedValue != "" {
		// 缓存命中，直接返回
		return cachedValue, nil
	}

	// 缓存未命中，从数据库查询
	setting, err := s.db.Settings.Query().
		Where(settings.KeyEQ(key)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// 如果设置不存在，缓存默认值并返回默认值
			if err = s.cache.SetEx(cacheKey, defaultValue, 300); err != nil {
				s.logger.Warn("设置缓存默认值失败", tracing.WithTraceIDField(ctx), zap.Error(err))
			} // 缓存5分钟
			return defaultValue, nil
		}
		s.logger.Error("查询设置失败", tracing.WithTraceIDField(ctx), zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("查询设置失败: %w", err)
	}

	// 将查询结果缓存到Redis，设置1天过期时间
	if err = s.cache.SetEx(cacheKey, setting.Value, 86400); err != nil {
		s.logger.Warn("缓存设置值失败", tracing.WithTraceIDField(ctx), zap.String("key", key), zap.Error(err))
	}

	return setting.Value, nil
}

// ClearSettingCache 清理指定设置的Redis缓存 - 公共方法
func (s *SettingsService) ClearSettingCache(key string) {
	cacheKey := _const.GetSettingKey(key)
	if _, err := s.cache.Del(cacheKey); err != nil {
		s.logger.Warn("清理设置缓存失败", zap.String("key", key), zap.Error(err))
	}
}
