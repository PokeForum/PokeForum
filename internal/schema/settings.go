package schema

// RoutineSettingsRequest Routine settings request | 常规设置请求体
type RoutineSettingsRequest struct {
	// Website Logo URL | 网站Logo URL
	WebSiteLogo string `json:"website_logo" binding:"omitempty,url" example:"https://example.com/logo.png"`
	// Website Icon URL | 网站Icon URL
	WebSiteIcon string `json:"website_icon" binding:"omitempty,url" example:"https://example.com/icon.ico"`
	// ICP filing number | ICP备案号
	ICPRecord string `json:"icp_record" binding:"omitempty,max=100" example:"京ICP备12345678号"`
	// Public security network filing number | 公安联网备案号
	PublicSecurityNetwork string `json:"public_security_network" binding:"omitempty,max=100" example:"京公网安备11010802012345号"`
	// Whether to close copyright information display | 是否关闭版权信息显示
	IsCloseCopyright bool `json:"is_close_copyright" example:"false"`
}

// RoutineSettingsResponse Routine settings response | 常规设置响应体
type RoutineSettingsResponse struct {
	// Website Logo URL | 网站Logo URL
	WebSiteLogo string `json:"website_logo" example:"https://example.com/logo.png"`
	// Website Icon URL | 网站Icon URL
	WebSiteIcon string `json:"website_icon" example:"https://example.com/icon.ico"`
	// ICP filing number | ICP备案号
	ICPRecord string `json:"icp_record" example:"京ICP备12345678号"`
	// Public security network filing number | 公安联网备案号
	PublicSecurityNetwork string `json:"public_security_network" example:"京公网安备11010802012345号"`
	// Whether to close copyright information display | 是否关闭版权信息显示
	IsCloseCopyright bool `json:"is_close_copyright" example:"false"`
}

// SlideItem Slide item | 幻灯片项
type SlideItem struct {
	// Image URL | 图片URL
	ImageURL string `json:"image_url" binding:"required,url" example:"https://example.com/slide1.jpg"`
	// Link URL | 链接URL
	LinkURL string `json:"link_url" binding:"omitempty,url" example:"https://example.com/article/1"`
	// Title | 标题
	Title string `json:"title" binding:"omitempty,max=100" example:"欢迎来到PokeForum"`
	// Description | 描述
	Description string `json:"description" binding:"omitempty,max=500" example:"这是一个友好的社区"`
}

// LinkItem Link item (friendly links) | 链接项（友情链接）
type LinkItem struct {
	// Link name | 链接名称
	Name string `json:"name" binding:"required,min=1,max=50" example:"示例网站"`
	// Link URL | 链接URL
	URL string `json:"url" binding:"required,url" example:"https://example.com"`
	// Link description | 链接描述
	Description string `json:"description" binding:"omitempty,max=200" example:"一个很棒的网站"`
}

// HomeSettingsRequest Home page settings request | 首页设置请求体
type HomeSettingsRequest struct {
	// Slide list | 幻灯片列表
	Slides []SlideItem `json:"slides" binding:"omitempty,dive"`
	// Friendly links list | 友情链接列表
	Links []LinkItem `json:"links" binding:"omitempty,dive"`
}

// HomeSettingsResponse Home page settings response | 首页设置响应体
type HomeSettingsResponse struct {
	// Slide list | 幻灯片列表
	Slides []SlideItem `json:"slides"`
	// Friendly links list | 友情链接列表
	Links []LinkItem `json:"links"`
}

// CommentSettingsRequest Comment settings request | 评论设置请求体
type CommentSettingsRequest struct {
	// Whether to show commenter information | 是否显示评论者信息
	ShowCommentInfo bool `json:"show_comment_info" example:"true"`
	// Whether comment approval is required | 是否需要审核评论
	RequireApproval bool `json:"require_approval" example:"false"`
	// Keyword blacklist (comma-separated) | 关键词黑名单（逗号分隔）
	KeywordBlacklist string `json:"keyword_blacklist" binding:"omitempty,max=5000" example:"垃圾,广告,spam"`
}

// CommentSettingsResponse Comment settings response | 评论设置响应体
type CommentSettingsResponse struct {
	// Whether to show commenter information | 是否显示评论者信息
	ShowCommentInfo bool `json:"show_comment_info" example:"true"`
	// Whether comment approval is required | 是否需要审核评论
	RequireApproval bool `json:"require_approval" example:"false"`
	// Keyword blacklist | 关键词黑名单
	KeywordBlacklist string `json:"keyword_blacklist" example:"垃圾,广告,spam"`
}

// SeoSettingsRequest SEO settings request | SEO设置请求体
type SeoSettingsRequest struct {
	// Website name | 网站名称
	WebSiteName string `json:"website_name" binding:"required,min=1,max=100" example:"PokeForum"`
	// Website keywords | 网站关键词
	WebSiteKeyword string `json:"website_keyword" binding:"omitempty,max=500" example:"论坛,社区,讨论"`
	// Website description | 网站描述
	WebSiteDescription string `json:"website_description" binding:"omitempty,max=1000" example:"一个友好的在线社区论坛"`
}

// SeoSettingsResponse SEO settings response | SEO设置响应体
type SeoSettingsResponse struct {
	// Website name | 网站名称
	WebSiteName string `json:"website_name" example:"PokeForum"`
	// Website keywords | 网站关键词
	WebSiteKeyword string `json:"website_keyword" example:"论坛,社区,讨论"`
	// Website description | 网站描述
	WebSiteDescription string `json:"website_description" example:"一个友好的在线社区论坛"`
}

// CodeSettingsRequest Code configuration request | 代码配置请求体
type CodeSettingsRequest struct {
	// Header code (HTML/JavaScript) | 页头代码（HTML/JavaScript）
	Header string `json:"header" binding:"omitempty,max=10000" example:"<script>console.log('header');</script>"`
	// Footer code (HTML/JavaScript) | 页脚代码（HTML/JavaScript）
	Footer string `json:"footer" binding:"omitempty,max=10000" example:"<script>console.log('footer');</script>"`
	// Custom CSS | 自定义CSS
	CustomizationCSS string `json:"customization_css" binding:"omitempty,max=50000" example:"body { background-color: #f0f0f0; }"`
}

// CodeSettingsResponse Code configuration response | 代码配置响应体
type CodeSettingsResponse struct {
	// Header code | 页头代码
	Header string `json:"header" example:"<script>console.log('header');</script>"`
	// Footer code | 页脚代码
	Footer string `json:"footer" example:"<script>console.log('footer');</script>"`
	// Custom CSS | 自定义CSS
	CustomizationCSS string `json:"customization_css" example:"body { background-color: #f0f0f0; }"`
}

// SafeSettingsRequest Security settings request | 安全设置请求体
type SafeSettingsRequest struct {
	// Whether to close registration | 是否关闭注册
	IsCloseRegister bool `json:"is_close_register" example:"false"`
	// Whether to enable email whitelist | 是否启用邮箱白名单
	IsEnableEmailWhitelist bool `json:"is_enable_email_whitelist" example:"false"`
	// Email whitelist (comma-separated domains) | 邮箱白名单（逗号分隔的域名）
	EmailWhitelist string `json:"email_whitelist" binding:"omitempty,max=5000" example:"gmail.com,qq.com,163.com"`
	// Whether email verification is required | 是否需要验证邮箱
	VerifyEmail bool `json:"verify_email" example:"true"`
}

// SafeSettingsResponse Security settings response | 安全设置响应体
type SafeSettingsResponse struct {
	// Whether to close registration | 是否关闭注册
	IsCloseRegister bool `json:"is_close_register" example:"false"`
	// Whether to enable email whitelist | 是否启用邮箱白名单
	IsEnableEmailWhitelist bool `json:"is_enable_email_whitelist" example:"false"`
	// Email whitelist | 邮箱白名单
	EmailWhitelist string `json:"email_whitelist" example:"gmail.com,qq.com,163.com"`
	// Whether email verification is required | 是否需要验证邮箱
	VerifyEmail bool `json:"verify_email" example:"true"`
}

// SigninSettingsRequest Sign-in settings request | 签到设置请求体
type SigninSettingsRequest struct {
	// Whether to enable sign-in feature | 是否启用签到功能
	IsEnable bool `json:"is_enable" example:"true"`
	// Sign-in mode: fixed, increment, random | 签到模式：fixed、increment、random
	Mode string `json:"mode" binding:"required,oneof=fixed increment random" example:"fixed"`
	// Fixed mode reward points | 固定模式奖励积分
	FixedReward int `json:"fixed_reward" binding:"omitempty,min=1,max=1000" example:"10"`
	// Increment mode base reward | 递增模式基础奖励
	IncrementBase int `json:"increment_base" binding:"omitempty,min=1,max=1000" example:"5"`
	// Increment mode step | 递增模式步长
	IncrementStep int `json:"increment_step" binding:"omitempty,min=1,max=100" example:"1"`
	// Increment cycle (days), restart increment after this period | 递增周期（天数），超过此周期后重新开始递增
	IncrementCycle int `json:"increment_cycle" binding:"omitempty,min=1,max=365" example:"7"`
	// Random mode minimum reward | 随机模式最小奖励
	RandomMin int `json:"random_min" binding:"omitempty,min=1,max=1000" example:"5"`
	// Random mode maximum reward | 随机模式最大奖励
	RandomMax int `json:"random_max" binding:"omitempty,min=1,max=1000" example:"20"`
	// Experience reward ratio | 经验值奖励比例
	ExperienceReward float64 `json:"experience_reward" binding:"omitempty,min=0,max=10" example:"1.0"`
}

// SigninSettingsResponse Sign-in settings response | 签到设置响应体
type SigninSettingsResponse struct {
	// Whether to enable sign-in feature | 是否启用签到功能
	IsEnable bool `json:"is_enable" example:"true"`
	// Sign-in mode: fixed, increment, random | 签到模式：fixed、increment、random
	Mode string `json:"mode" example:"fixed"`
	// Fixed mode reward points | 固定模式奖励积分
	FixedReward int `json:"fixed_reward" example:"10"`
	// Increment mode base reward | 递增模式基础奖励
	IncrementBase int `json:"increment_base" example:"5"`
	// Increment mode step | 递增模式步长
	IncrementStep int `json:"increment_step" example:"1"`
	// Increment cycle (days), restart increment after this period | 递增周期（天数），超过此周期后重新开始递增
	IncrementCycle int `json:"increment_cycle" example:"7"`
	// Random mode minimum reward | 随机模式最小奖励
	RandomMin int `json:"random_min" example:"5"`
	// Random mode maximum reward | 随机模式最大奖励
	RandomMax int `json:"random_max" example:"20"`
	// Experience reward ratio | 经验值奖励比例
	ExperienceReward float64 `json:"experience_reward" example:"1.0"`
}

// PublicConfigResponse Public configuration response (client-accessible configuration) | 公开配置响应体（客户端可获取的配置）
type PublicConfigResponse struct {
	Routine *RoutineSettingsResponse `json:"routine"`
	Home    *HomeSettingsResponse    `json:"home"`
	Seo     *SeoSettingsResponse     `json:"seo"`
	Safe    *SafeSettingsResponse    `json:"safe"`
	Code    *CodeSettingsResponse    `json:"code"`
	Comment *CommentSettingsResponse `json:"comment"`
}
