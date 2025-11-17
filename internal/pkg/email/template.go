package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	_const "github.com/PokeForum/PokeForum/internal/const"
	"go.uber.org/zap"
)

// SettingsService 设置服务接口，避免循环导入
type SettingsService interface {
	GetSettingByKey(ctx context.Context, key, defaultValue string) (string, error)
}

// TemplateData 邮件模板数据
type TemplateData struct {
	VerifyCode    string        // 验证码
	CommonContext CommonContext // 通用上下文
}

// CommonContext 通用邮件上下文
type CommonContext struct {
	SiteBasic SiteBasic // 网站基本信息
	Logo      Logo      // 网站Logo
	SiteUrl   string    // 网站URL
}

// SiteBasic 网站基本信息
type SiteBasic struct {
	Name string // 网站名称
}

// Logo 网站Logo
type Logo struct {
	Normal string // 正常Logo
}

// Template 邮件模板渲染器
type Template struct {
	settingsService SettingsService
	logger          *zap.Logger
}

// NewEmailTemplate 创建邮件模板渲染器
func NewEmailTemplate(settingsService SettingsService, logger *zap.Logger) *Template {
	return &Template{
		settingsService: settingsService,
		logger:          logger,
	}
}

// RenderEmailVerificationTemplate 渲染邮箱验证模板
func (et *Template) RenderEmailVerificationTemplate(ctx context.Context, verifyCode string, siteName string) (string, error) {
	// 构建模板数据
	data := TemplateData{
		VerifyCode: verifyCode,
		CommonContext: CommonContext{
			SiteBasic: SiteBasic{
				Name: siteName,
			},
		},
	}

	// 从数据库获取自定义模板
	customTemplate, err := et.settingsService.GetSettingByKey(ctx, _const.EmailAccountActivationTemplate, "")
	if err != nil {
		et.logger.Warn("获取自定义邮件模板失败，使用默认模板", zap.Error(err))
		customTemplate = ""
	}

	// 如果自定义模板为空，使用默认模板
	templateContent := customTemplate
	if strings.TrimSpace(templateContent) == "" {
		templateContent = _const.DefaultEmailAccountActivationTemplate
	}

	// 渲染模板
	return et.renderTemplate(templateContent, data)
}

// RenderPasswordResetTemplate 渲染密码重置模板
func (et *Template) RenderPasswordResetTemplate(ctx context.Context, verifyCode string, siteName string) (string, error) {
	// 构建模板数据
	data := TemplateData{
		VerifyCode: verifyCode,
		CommonContext: CommonContext{
			SiteBasic: SiteBasic{
				Name: siteName,
			},
		},
	}

	// 从数据库获取自定义模板
	customTemplate, err := et.settingsService.GetSettingByKey(ctx, _const.EmailPasswordResetTemplate, "")
	if err != nil {
		et.logger.Warn("获取自定义密码重置邮件模板失败，使用默认模板", zap.Error(err))
		customTemplate = ""
	}

	// 如果自定义模板为空，使用默认模板
	templateContent := customTemplate
	if strings.TrimSpace(templateContent) == "" {
		templateContent = _const.DefaultEmailPasswordResetTemplate
	}

	// 渲染模板
	return et.renderTemplate(templateContent, data)
}

// renderTemplate 渲染模板内容
func (et *Template) renderTemplate(templateContent string, data TemplateData) (string, error) {
	// 创建模板实例
	tmpl, err := template.New("email").Parse(templateContent)
	if err != nil {
		et.logger.Error("解析邮件模板失败", zap.Error(err))
		return "", fmt.Errorf("解析邮件模板失败: %w", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		et.logger.Error("渲染邮件模板失败", zap.Error(err))
		return "", fmt.Errorf("渲染邮件模板失败: %w", err)
	}

	return buf.String(), nil
}
