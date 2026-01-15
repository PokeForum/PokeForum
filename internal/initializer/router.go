package initializer

import (
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	_ "github.com/PokeForum/PokeForum/docs"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/controller"
	"github.com/PokeForum/PokeForum/internal/middleware"
	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
	"github.com/PokeForum/PokeForum/internal/service"
)

func Routers(injector *do.Injector) *gin.Engine {
	// Set up SaToken | 设置SaToken
	saManager := satoken.NewSaToken()
	saGin.SetManager(saManager)

	// Create Gin plugin | 创建 Gin 插件
	saPlugin := saGin.NewPlugin(saManager)

	// Set mode | 设置模式
	if !configs.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	Router := gin.New()
	Router.Use(middleware.Logger())
	Router.Use(middleware.Recovery())

	// CORS configuration | 跨域配置
	Router.Use(cors.New(middleware.CorsConfig))

	// Security response headers | 安全响应头
	Router.Use(middleware.SecurityHeaders())

	// Global rate limiting (100 requests per second) | 全局速率限制（每秒100个请求）
	Router.Use(middleware.RateLimit(middleware.DefaultRateLimitConfig))

	// Register services to the injector | 注册服务到注入器
	InjectorSrv(injector)

	// Services
	healthService := do.MustInvoke[service.IHealthService](injector)
	authService := do.MustInvoke[service.IAuthService](injector)
	settingsService := do.MustInvoke[service.ISettingsService](injector)
	userProfileService := do.MustInvoke[service.IUserProfileService](injector)
	blacklistService := do.MustInvoke[service.IBlacklistService](injector)
	rankingService := do.MustInvoke[service.IRankingService](injector)
	categoryService := do.MustInvoke[service.ICategoryService](injector)
	postService := do.MustInvoke[service.IPostService](injector)
	commentService := do.MustInvoke[service.ICommentService](injector)
	signinService := do.MustInvoke[service.ISigninService](injector)
	moderatorService := do.MustInvoke[service.IModeratorService](injector)
	dashboardService := do.MustInvoke[service.IDashboardService](injector)
	userManageService := do.MustInvoke[service.IUserManageService](injector)
	categoryManageService := do.MustInvoke[service.ICategoryManageService](injector)
	postManageService := do.MustInvoke[service.IPostManageService](injector)
	commentManageService := do.MustInvoke[service.ICommentManageService](injector)
	oauthProviderService := do.MustInvoke[service.IOAuthProviderService](injector)
	oauthService := do.MustInvoke[service.IOAuthService](injector)

	// Health check route (not affected by rate limiting, outside of api group) | 健康检查路由（不受速率限制影响，在api分组之外）
	healthCon := controller.NewHealthController(healthService)
	healthCon.HealthRouter(Router)

	if configs.Debug {
		Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Register Prometheus | 注册 Prometheus
	if configs.Prometheus {
		p := ginprometheus.NewPrometheus("gin")
		p.Use(Router)
	}

	api := Router.Group("/api/v1")

	// Authentication verification (add stricter rate limiting to prevent brute force attacks) | 认证校验（添加更严格的速率限制，防止暴力破解）
	AuthGroup := api.Group("/auth")
	AuthGroup.Use(middleware.RateLimit(middleware.AuthRateLimitConfig))
	AuthCon := controller.NewAuthController(authService)
	AuthCon.AuthRouter(AuthGroup)

	// OAuth login public routes (no login required) | OAuth登录公开路由（无需登录）
	OAuthPublicGroup := AuthGroup.Group("/oauth")
	OAuthCon := controller.NewOAuthController(oauthService)
	OAuthCon.OAuthPublicRouter(OAuthPublicGroup)

	// Configuration | 配置
	ConfigGroup := api.Group("/config")
	ConfigCon := controller.NewConfigController(settingsService)
	ConfigCon.ConfigRouter(ConfigGroup)

	// Add login verification | 添加登录校验
	AuthAPIGroup := api.Group("")
	AuthAPIGroup.Use(saPlugin.AuthMiddleware())

	// Forum interfaces - User-side permission checks are performed in the Controller | 论坛接口 - 用户侧权限校验放在Controller检查
	ForumGroup := api.Group("")
	{
		// User | 用户
		{
			// User Profile | 个人中心
			{
				ProfileGroup := ForumGroup.Group("/profile")
				ProfileCon := controller.NewUserProfileController(userProfileService)
				ProfileCon.UserProfileRouter(ProfileGroup)

				// Blacklist | 拉黑用户
				BlacklistGroup := ForumGroup.Group("/profile/blacklist")
				BlacklistCon := controller.NewBlacklistController(blacklistService)
				BlacklistCon.BlacklistRouter(BlacklistGroup)

				// OAuth user routes | OAuth用户路由
				OAuthUserGroup := AuthAPIGroup.Group("/user/oauth")
				OAuthCon.OAuthUserRouter(OAuthUserGroup)

				// TODO Report | 举报

				/*
					- Invitation code registration | 邀请码注册
					- Toggle configuration for invitation code mechanism | 开关配置是否开启邀请码机制
					- Number of invitation codes a user can create (limited/unlimited) | 用户可创建邀请码数量（有限/无限）
					- Number of times an invitation code can be used (limited/unlimited) | 邀请码使用次数（有限/无限）
				*/
				// TODO Invitation Code | 邀请码
			}

			// TODO Discovery | 发现
			{
			}

			// Ranking | 排行榜
			RankingGroup := ForumGroup.Group("/ranking")
			RankingCon := controller.NewRankingController(rankingService)
			RankingCon.RankingRouter(RankingGroup)

			// Category | 版块
			CategoryGroup := ForumGroup.Group("/categories")
			CategoryCon := controller.NewCategoryController(categoryService)
			CategoryCon.CategoryRouter(CategoryGroup)

			// Post | 主题帖
			PostGroup := ForumGroup.Group("/posts")
			PostCon := controller.NewPostController(postService)
			PostCon.PostRouter(PostGroup)

			// Comment | 评论
			CommentGroup := ForumGroup.Group("/comments")
			CommentCon := controller.NewCommentController(commentService)
			CommentCon.CommentRouter(CommentGroup)

			// Sign-in System | 签到系统
			SigninGroup := ForumGroup.Group("/signin")
			SigninCon := controller.NewSigninController(signinService)
			SigninCon.SigninRouter(SigninGroup)
		}

		// Moderator Interface | 版主接口
		ModeratorGroup := ForumGroup.Group("/moderator")
		ModeratorGroup.Use(saGin.CheckRole(user.RoleModerator.String()))
		ModeratorCon := controller.NewModeratorController(moderatorService)
		ModeratorCon.ModeratorRouter(ModeratorGroup)
	}

	// Administrator Interface | 管理员接口
	ManageGroup := AuthAPIGroup.Group("/manage")
	ManageGroup.Use(saGin.CheckRole(user.RoleAdmin.String()))
	{
		// Dashboard | 仪表盘
		{
			DashboardGroup := ManageGroup.Group("/dashboard")
			DashboardCon := controller.NewDashboardController(dashboardService)
			DashboardCon.DashboardRouter(DashboardGroup)
		}

		// User Management | 用户管理
		{
			UserManageGroup := ManageGroup.Group("/users")
			UserManageCon := controller.NewUserManageController(userManageService)
			UserManageCon.UserManageRouter(UserManageGroup)
		}

		// Category Management | 版块管理
		{
			CategoryManageGroup := ManageGroup.Group("/categories")
			CategoryManageCon := controller.NewCategoryManageController(categoryManageService)
			CategoryManageCon.CategoryManageRouter(CategoryManageGroup)
		}

		// Post Management | 帖子管理
		{
			PostManageGroup := ManageGroup.Group("/posts")
			PostManageCon := controller.NewPostManageController(postManageService)
			PostManageCon.PostManageRouter(PostManageGroup)
		}

		// Comment Management | 评论管理
		{
			CommentManageGroup := ManageGroup.Group("/comments")
			CommentManageCon := controller.NewCommentManageController(commentManageService)
			CommentManageCon.CommentManageRouter(CommentManageGroup)
		}

		// TODO Report Management | 举报管理
	}

	// Super Administrator Interface | 超级管理接口
	SuperManageGroup := AuthAPIGroup.Group("/super/manage")
	SuperManageGroup.Use(saGin.CheckRole(user.RoleSuperAdmin.String()))
	{
		// Settings Management (Unified settings controller, includes all system settings) | 设置管理（统一的设置控制器，包含所有系统设置）
		SettingsGroup := SuperManageGroup.Group("/settings")
		SettingsCon := controller.NewSettingsController(settingsService)
		SettingsCon.SettingsRouter(SettingsGroup)

		// OAuth Provider Management | OAuth提供商管理
		OAuthGroup := SuperManageGroup.Group("/settings/oauth")
		{
			OAuthProviderCon := controller.NewOAuthProviderController(oauthProviderService)
			OAuthProviderCon.OAuthProviderRouter(OAuthGroup)
		}

		// TODO Advertisement Settings | 广告设置
	}

	return Router
}
