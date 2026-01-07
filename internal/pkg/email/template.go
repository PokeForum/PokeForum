package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"go.uber.org/zap"

	_const "github.com/PokeForum/PokeForum/internal/consts"
)

// SettingsService Settings service interface, avoid circular import | 设置服务接口，避免循环导入
type SettingsService interface {
	GetSettingByKey(ctx context.Context, key, defaultValue string) (string, error)
}

// TemplateData Email template data | 邮件模板数据
type TemplateData struct {
	VerifyCode    string        // Verification code | 验证码
	CommonContext CommonContext // Common context | 通用上下文
}

// CommonContext Common email context | 通用邮件上下文
type CommonContext struct {
	SiteBasic SiteBasic // Site basic information | 网站基本信息
	Logo      Logo      // Site logo | 网站Logo
	SiteUrl   string    // Site URL | 网站URL
}

// SiteBasic Site basic information | 网站基本信息
type SiteBasic struct {
	Name string // Site name | 网站名称
}

// Logo Site logo | 网站Logo
type Logo struct {
	Normal string // Normal logo | 正常Logo
}

// Template Email template renderer | 邮件模板渲染器
type Template struct {
	settingsService SettingsService
	logger          *zap.Logger
}

// NewEmailTemplate Create email template renderer | 创建邮件模板渲染器
func NewEmailTemplate(settingsService SettingsService, logger *zap.Logger) *Template {
	return &Template{
		settingsService: settingsService,
		logger:          logger,
	}
}

// RenderEmailVerificationTemplate Render email verification template | 渲染邮箱验证模板
func (et *Template) RenderEmailVerificationTemplate(ctx context.Context, verifyCode string, siteName string) (string, error) {
	// Build template data | 构建模板数据
	data := TemplateData{
		VerifyCode: verifyCode,
		CommonContext: CommonContext{
			SiteBasic: SiteBasic{
				Name: siteName,
			},
		},
	}

	// Get custom template from database | 从数据库获取自定义模板
	customTemplate, err := et.settingsService.GetSettingByKey(ctx, _const.EmailAccountActivationTemplate, "")
	if err != nil {
		et.logger.Warn("获取自定义邮件模板失败，使用默认模板", zap.Error(err))
		customTemplate = ""
	}

	// If custom template is empty, use default template | 如果自定义模板为空，使用默认模板
	templateContent := customTemplate
	if strings.TrimSpace(templateContent) == "" {
		templateContent = _const.DefaultEmailAccountActivationTemplate
	}

	// Render template | 渲染模板
	return et.renderTemplate(templateContent, data)
}

// RenderPasswordResetTemplate Render password reset template | 渲染密码重置模板
func (et *Template) RenderPasswordResetTemplate(ctx context.Context, verifyCode string, siteName string) (string, error) {
	// Build template data | 构建模板数据
	data := TemplateData{
		VerifyCode: verifyCode,
		CommonContext: CommonContext{
			SiteBasic: SiteBasic{
				Name: siteName,
			},
		},
	}

	// Get custom template from database | 从数据库获取自定义模板
	customTemplate, err := et.settingsService.GetSettingByKey(ctx, _const.EmailPasswordResetTemplate, "")
	if err != nil {
		et.logger.Warn("获取自定义密码重置邮件模板失败，使用默认模板", zap.Error(err))
		customTemplate = ""
	}

	// If custom template is empty, use default template | 如果自定义模板为空，使用默认模板
	templateContent := customTemplate
	if strings.TrimSpace(templateContent) == "" {
		templateContent = _const.DefaultEmailPasswordResetTemplate
	}

	// Render template | 渲染模板
	return et.renderTemplate(templateContent, data)
}

// renderTemplate Render template content | 渲染模板内容
func (et *Template) renderTemplate(templateContent string, data TemplateData) (string, error) {
	// Create template instance | 创建模板实例
	tmpl, err := template.New("email").Parse(templateContent)
	if err != nil {
		et.logger.Error("解析邮件模板失败", zap.Error(err))
		return "", fmt.Errorf("解析邮件模板失败: %w", err)
	}

	// Render template | 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		et.logger.Error("渲染邮件模板失败", zap.Error(err))
		return "", fmt.Errorf("渲染邮件模板失败: %w", err)
	}

	return buf.String(), nil
}
