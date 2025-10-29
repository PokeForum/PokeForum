package schema

// RoutineSettingsRequest 常规设置请求体
type RoutineSettingsRequest struct {
	// 网站Logo URL
	WebSiteLogo string `json:"website_logo" binding:"omitempty,url" example:"https://example.com/logo.png"`
	// 网站Icon URL
	WebSiteIcon string `json:"website_icon" binding:"omitempty,url" example:"https://example.com/icon.ico"`
	// ICP备案号
	ICPRecord string `json:"icp_record" binding:"omitempty,max=100" example:"京ICP备12345678号"`
	// 公安联网备案号
	PublicSecurityNetwork string `json:"public_security_network" binding:"omitempty,max=100" example:"京公网安备11010802012345号"`
	// 是否关闭版权信息显示
	IsCloseCopyright bool `json:"is_close_copyright" example:"false"`
}

// RoutineSettingsResponse 常规设置响应体
type RoutineSettingsResponse struct {
	// 网站Logo URL
	WebSiteLogo string `json:"website_logo" example:"https://example.com/logo.png"`
	// 网站Icon URL
	WebSiteIcon string `json:"website_icon" example:"https://example.com/icon.ico"`
	// ICP备案号
	ICPRecord string `json:"icp_record" example:"京ICP备12345678号"`
	// 公安联网备案号
	PublicSecurityNetwork string `json:"public_security_network" example:"京公网安备11010802012345号"`
	// 是否关闭版权信息显示
	IsCloseCopyright bool `json:"is_close_copyright" example:"false"`
}

// SlideItem 幻灯片项
type SlideItem struct {
	// 图片URL
	ImageURL string `json:"image_url" binding:"required,url" example:"https://example.com/slide1.jpg"`
	// 链接URL
	LinkURL string `json:"link_url" binding:"omitempty,url" example:"https://example.com/article/1"`
	// 标题
	Title string `json:"title" binding:"omitempty,max=100" example:"欢迎来到PokeForum"`
	// 描述
	Description string `json:"description" binding:"omitempty,max=500" example:"这是一个友好的社区"`
}

// LinkItem 链接项（友情链接）
type LinkItem struct {
	// 链接名称
	Name string `json:"name" binding:"required,min=1,max=50" example:"示例网站"`
	// 链接URL
	URL string `json:"url" binding:"required,url" example:"https://example.com"`
	// 链接描述
	Description string `json:"description" binding:"omitempty,max=200" example:"一个很棒的网站"`
}

// HomeSettingsRequest 首页设置请求体
type HomeSettingsRequest struct {
	// 幻灯片列表
	Slides []SlideItem `json:"slides" binding:"omitempty,dive"`
	// 友情链接列表
	Links []LinkItem `json:"links" binding:"omitempty,dive"`
}

// HomeSettingsResponse 首页设置响应体
type HomeSettingsResponse struct {
	// 幻灯片列表
	Slides []SlideItem `json:"slides"`
	// 友情链接列表
	Links []LinkItem `json:"links"`
}

// CommentSettingsRequest 评论设置请求体
type CommentSettingsRequest struct {
	// 是否显示评论者信息
	ShowCommentInfo bool `json:"show_comment_info" example:"true"`
	// 是否需要审核评论
	RequireApproval bool `json:"require_approval" example:"false"`
	// 关键词黑名单（逗号分隔）
	KeywordBlacklist string `json:"keyword_blacklist" binding:"omitempty,max=5000" example:"垃圾,广告,spam"`
}

// CommentSettingsResponse 评论设置响应体
type CommentSettingsResponse struct {
	// 是否显示评论者信息
	ShowCommentInfo bool `json:"show_comment_info" example:"true"`
	// 是否需要审核评论
	RequireApproval bool `json:"require_approval" example:"false"`
	// 关键词黑名单
	KeywordBlacklist string `json:"keyword_blacklist" example:"垃圾,广告,spam"`
}

// SeoSettingsRequest SEO设置请求体
type SeoSettingsRequest struct {
	// 网站名称
	WebSiteName string `json:"website_name" binding:"required,min=1,max=100" example:"PokeForum"`
	// 网站关键词
	WebSiteKeyword string `json:"website_keyword" binding:"omitempty,max=500" example:"论坛,社区,讨论"`
	// 网站描述
	WebSiteDescription string `json:"website_description" binding:"omitempty,max=1000" example:"一个友好的在线社区论坛"`
}

// SeoSettingsResponse SEO设置响应体
type SeoSettingsResponse struct {
	// 网站名称
	WebSiteName string `json:"website_name" example:"PokeForum"`
	// 网站关键词
	WebSiteKeyword string `json:"website_keyword" example:"论坛,社区,讨论"`
	// 网站描述
	WebSiteDescription string `json:"website_description" example:"一个友好的在线社区论坛"`
}

// CodeSettingsRequest 代码配置请求体
type CodeSettingsRequest struct {
	// 页头代码（HTML/JavaScript）
	Header string `json:"header" binding:"omitempty,max=10000" example:"<script>console.log('header');</script>"`
	// 页脚代码（HTML/JavaScript）
	Footer string `json:"footer" binding:"omitempty,max=10000" example:"<script>console.log('footer');</script>"`
	// 自定义CSS
	CustomizationCSS string `json:"customization_css" binding:"omitempty,max=50000" example:"body { background-color: #f0f0f0; }"`
}

// CodeSettingsResponse 代码配置响应体
type CodeSettingsResponse struct {
	// 页头代码
	Header string `json:"header" example:"<script>console.log('header');</script>"`
	// 页脚代码
	Footer string `json:"footer" example:"<script>console.log('footer');</script>"`
	// 自定义CSS
	CustomizationCSS string `json:"customization_css" example:"body { background-color: #f0f0f0; }"`
}

// SafeSettingsRequest 安全设置请求体
type SafeSettingsRequest struct {
	// 是否关闭注册
	IsCloseRegister bool `json:"is_close_register" example:"false"`
	// 是否启用邮箱白名单
	IsEnableEmailWhitelist bool `json:"is_enable_email_whitelist" example:"false"`
	// 邮箱白名单（逗号分隔的域名）
	EmailWhitelist string `json:"email_whitelist" binding:"omitempty,max=5000" example:"gmail.com,qq.com,163.com"`
	// 是否需要验证邮箱
	VerifyEmail bool `json:"verify_email" example:"true"`
}

// SafeSettingsResponse 安全设置响应体
type SafeSettingsResponse struct {
	// 是否关闭注册
	IsCloseRegister bool `json:"is_close_register" example:"false"`
	// 是否启用邮箱白名单
	IsEnableEmailWhitelist bool `json:"is_enable_email_whitelist" example:"false"`
	// 邮箱白名单
	EmailWhitelist string `json:"email_whitelist" example:"gmail.com,qq.com,163.com"`
	// 是否需要验证邮箱
	VerifyEmail bool `json:"verify_email" example:"true"`
}
