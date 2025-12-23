package _const

// SettingPrefix 缓存前缀
const SettingPrefix = "setting:"

// GetSettingKey 获取设置键名
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

// 常规设置
const (
	// RoutineWebSiteLogo 网站Logo
	RoutineWebSiteLogo = "routine:website_logo"
	// RoutineWebSiteIcon 网站Icon
	RoutineWebSiteIcon = "routine:website_icon"
	// RoutineICPRecord 网站ICP备案号
	RoutineICPRecord = "routine:icp_record"
	// RoutinePublicSecurityNetwork 网站公安联网备案号
	RoutinePublicSecurityNetwork = "routine:public_security_network"
	// RoutineIsCloseCopyright 是否关闭版权信息
	RoutineIsCloseCopyright = "routine:is_close_copyright"
)

// 首页设置
const (
	// HomeSlide 幻灯片
	HomeSlide = "home:slide"
	// HomeLinks 友情链接
	HomeLinks = "home:links"
)

// 评论设置
const (
	// CommentShowCommentInfo 是否显示评论者信息
	CommentShowCommentInfo = "comment:show_comment_info"
	// CommentRequireApproval 是否需要审核评论
	CommentRequireApproval = "comment:require_approval"
	// CommentKeywordBlacklist 关键词黑名单
	CommentKeywordBlacklist = "comment:keyword_blacklist"
)

// 邮箱服务设置
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

// SEO设置
const (
	// SeoWebSiteName 网站名称
	SeoWebSiteName = "seo:web_site_name"
	// SeoWebSiteKeyword 网站关键词
	SeoWebSiteKeyword = "seo:web_site_keyword"
	// SeoWebSiteDescription 网站描述
	SeoWebSiteDescription = "seo:web_site_description"
)

// 代码配置
const (
	// CodeHeader 页头代码
	CodeHeader = "code:header"
	// CodeFooter 页脚代码
	CodeFooter = "code:footer"
	// CodeCustomizationCSS 自定义CSS
	CodeCustomizationCSS = "code:customization_css"
)

// 安全设置
const (
	// SafeIsCloseRegister 是否关闭注册
	SafeIsCloseRegister = "safe:is_close_register"
	// SafeIsEnableEmailWhitelist 是否开启邮箱白名单
	SafeIsEnableEmailWhitelist = "safe:is_enable_email_whitelist"
	// SafeEmailWhitelist 邮箱白名单
	SafeEmailWhitelist = "safe:email_whitelist"
	// SafeVerifyEmail 是否验证邮箱
	SafeVerifyEmail = "safe:verify_email"
)

// 签到设置
const (
	// SigninIsEnable 是否启用签到功能
	SigninIsEnable = "signin:is_enable"
	// SigninMode 签到模式：fixed、increment、random
	SigninMode = "signin:mode"
	// SigninFixedReward 固定模式奖励积分
	SigninFixedReward = "signin:fixed_reward"
	// SigninIncrementBase 递增模式基础奖励
	SigninIncrementBase = "signin:increment_base"
	// SigninIncrementStep 递增模式步长
	SigninIncrementStep = "signin:increment_step"
	// SigninIncrementCycle 递增周期（天数），超过此周期后重新开始递增
	SigninIncrementCycle = "signin:increment_cycle"
	// SigninRandomMin 随机模式最小奖励
	SigninRandomMin = "signin:random_min"
	// SigninRandomMax 随机模式最大奖励
	SigninRandomMax = "signin:random_max"
	// SigninExperienceReward 经验值奖励
	SigninExperienceReward = "signin:experience_reward"
)
