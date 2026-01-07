package _const

// SettingPrefix Cache prefix | 缓存前缀
const SettingPrefix = "setting:"

// GetSettingKey Get setting key name | 获取设置键名
func GetSettingKey(key string) string {
	return SettingPrefix + key
}

type SettingBool string

const (
	SettingBoolTrue  SettingBool = "true"
	SettingBoolFalse SettingBool = "false"
)

func (s SettingBool) String() string {
	return string(s)
}

// Routine settings | 常规设置
const (
	// RoutineWebSiteLogo Website Logo | 网站Logo
	RoutineWebSiteLogo = "routine:website_logo"
	// RoutineWebSiteIcon Website Icon | 网站Icon
	RoutineWebSiteIcon = "routine:website_icon"
	// RoutineICPRecord Website ICP filing number | 网站ICP备案号
	RoutineICPRecord = "routine:icp_record"
	// RoutinePublicSecurityNetwork Website public security network filing number | 网站公安联网备案号
	RoutinePublicSecurityNetwork = "routine:public_security_network"
	// RoutineIsCloseCopyright Whether to close copyright information | 是否关闭版权信息
	RoutineIsCloseCopyright = "routine:is_close_copyright"
)

// Home page settings | 首页设置
const (
	// HomeSlide Slide show | 幻灯片
	HomeSlide = "home:slide"
	// HomeLinks Friendly links | 友情链接
	HomeLinks = "home:links"
)

// Comment settings | 评论设置
const (
	// CommentShowCommentInfo Whether to show commenter information | 是否显示评论者信息
	CommentShowCommentInfo = "comment:show_comment_info"
	// CommentRequireApproval Whether comment approval is required | 是否需要审核评论
	CommentRequireApproval = "comment:require_approval"
	// CommentKeywordBlacklist Keyword blacklist | 关键词黑名单
	CommentKeywordBlacklist = "comment:keyword_blacklist"
)

// Email service settings | 邮箱服务设置
const (
	// EmailIsEnableEmailService Whether to enable email service | 是否启用邮箱服务
	EmailIsEnableEmailService = "email:is_enable_email_service"
	// EmailSender Sender name | 发件人
	EmailSender = "email:sender"
	// EmailAddress Sender email address | 发件人邮箱
	EmailAddress = "email:address"
	// EmailHost SMTP server | SMTP 服务器
	EmailHost = "email:host"
	// EmailPort SMTP port | SMTP 端口
	EmailPort = "email:port"
	// EmailUsername SMTP username | SMTP 用户名
	EmailUsername = "email:username"
	// EmailPassword Sender password | 发件人密码
	EmailPassword = "email:password"
	// EmailForcedSSL Force SSL connection | 强制使用 SSL 连接
	EmailForcedSSL = "email:forced_ssl"
	// EmailConnectionValidity SMTP connection validity (seconds) | SMTP 连接有效期 (秒)
	EmailConnectionValidity = "email:connection_validity"

	// EmailAccountActivationTemplate Account activation template | 账户激活模板
	EmailAccountActivationTemplate = "email:account_activation_template"
	// EmailPasswordResetTemplate Password reset template | 重置密码模板
	EmailPasswordResetTemplate = "email:password_reset_template"
)

// SEO settings | SEO设置
const (
	// SeoWebSiteName Website name | 网站名称
	SeoWebSiteName = "seo:web_site_name"
	// SeoWebSiteKeyword Website keywords | 网站关键词
	SeoWebSiteKeyword = "seo:web_site_keyword"
	// SeoWebSiteDescription Website description | 网站描述
	SeoWebSiteDescription = "seo:web_site_description"
)

// Code configuration | 代码配置
const (
	// CodeHeader Header code | 页头代码
	CodeHeader = "code:header"
	// CodeFooter Footer code | 页脚代码
	CodeFooter = "code:footer"
	// CodeCustomizationCSS Custom CSS | 自定义CSS
	CodeCustomizationCSS = "code:customization_css"
)

// Security settings | 安全设置
const (
	// SafeIsCloseRegister Whether to close registration | 是否关闭注册
	SafeIsCloseRegister = "safe:is_close_register"
	// SafeIsEnableEmailWhitelist Whether to enable email whitelist | 是否开启邮箱白名单
	SafeIsEnableEmailWhitelist = "safe:is_enable_email_whitelist"
	// SafeEmailWhitelist Email whitelist | 邮箱白名单
	SafeEmailWhitelist = "safe:email_whitelist"
	// SafeVerifyEmail Whether to verify email | 是否验证邮箱
	SafeVerifyEmail = "safe:verify_email"
)

// Sign-in settings | 签到设置
const (
	// SigninIsEnable Whether to enable sign-in feature | 是否启用签到功能
	SigninIsEnable = "signin:is_enable"
	// SigninMode Sign-in mode: fixed, increment, random | 签到模式：fixed、increment、random
	SigninMode = "signin:mode"
	// SigninFixedReward Fixed mode reward points | 固定模式奖励积分
	SigninFixedReward = "signin:fixed_reward"
	// SigninIncrementBase Increment mode base reward | 递增模式基础奖励
	SigninIncrementBase = "signin:increment_base"
	// SigninIncrementStep Increment mode step | 递增模式步长
	SigninIncrementStep = "signin:increment_step"
	// SigninIncrementCycle Increment cycle (days), restart increment after this period | 递增周期（天数），超过此周期后重新开始递增
	SigninIncrementCycle = "signin:increment_cycle"
	// SigninRandomMin Random mode minimum reward | 随机模式最小奖励
	SigninRandomMin = "signin:random_min"
	// SigninRandomMax Random mode maximum reward | 随机模式最大奖励
	SigninRandomMax = "signin:random_max"
	// SigninExperienceReward Experience reward | 经验值奖励
	SigninExperienceReward = "signin:experience_reward"
)
