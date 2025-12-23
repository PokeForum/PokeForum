package initializer

import (
	_ "github.com/PokeForum/PokeForum/docs"
	"github.com/PokeForum/PokeForum/ent/user"
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/controller"
	"github.com/PokeForum/PokeForum/internal/middleware"
	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func Routers(injector *do.Injector) *gin.Engine {
	// 设置SaToken
	saManager := satoken.NewSaToken()
	saGin.SetManager(saManager)

	// 创建 Gin 插件
	saPlugin := saGin.NewPlugin(saManager)

	// 设置模式
	if !configs.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	Router := gin.New()
	Router.Use(middleware.Logger())
	Router.Use(middleware.Recovery())

	// 跨域配置
	Router.Use(cors.New(middleware.CorsConfig))

	// 安全响应头
	Router.Use(middleware.SecurityHeaders())

	// 全局速率限制（每秒100个请求）
	Router.Use(middleware.RateLimit(middleware.DefaultRateLimitConfig))

	// 注册服务到注入器
	InjectorSrv(injector)

	// 健康检查路由（不受速率限制影响，在api分组之外）
	healthCon := controller.NewHealthController()
	healthCon.HealthRouter(Router)

	if configs.Debug == true {
		Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 注册 Prometheus
	if configs.Prometheus {
		p := ginprometheus.NewPrometheus("gin")
		p.Use(Router)
	}

	api := Router.Group("/api/v1")

	// 认证校验（添加更严格的速率限制，防止暴力破解）
	AuthGroup := api.Group("/auth")
	AuthGroup.Use(middleware.RateLimit(middleware.AuthRateLimitConfig))
	AuthCon := controller.NewAuthController(injector)
	AuthCon.AuthRouter(AuthGroup)

	// 配置
	ConfigGroup := api.Group("/config")
	ConfigCon := controller.NewConfigController(injector)
	ConfigCon.ConfigRouter(ConfigGroup)

	// 添加登录校验
	AuthAPIGroup := api.Group("")
	AuthAPIGroup.Use(saPlugin.AuthMiddleware())

	// 论坛接口 - 用户侧权限校验放在Controller检查
	ForumGroup := api.Group("")
	{
		// 用户
		{
			// 个人中心
			{
				ProfileGroup := ForumGroup.Group("/profile")
				ProfileCon := controller.NewUserProfileController(injector)
				ProfileCon.UserProfileRouter(ProfileGroup)

				// 拉黑用户
				BlacklistGroup := ForumGroup.Group("/profile/blacklist")
				BlacklistCon := controller.NewBlacklistController(injector)
				BlacklistCon.BlacklistRouter(BlacklistGroup)

				// TODO 举报

				/*
					- 邀请码注册
					- 开关配置是否开启邀请码机制
					- 用户可创建邀请码数量（有限/无限）
					- 邀请码使用次数（有限/无限）
				*/
				// TODO 邀请码
			}

			// TODO 发现
			{
			}

			// 排行榜
			RankingGroup := ForumGroup.Group("/ranking")
			RankingCon := controller.NewRankingController(injector)
			RankingCon.RankingRouter(RankingGroup)

			// 版块
			CategoryGroup := ForumGroup.Group("/categories")
			CategoryCon := controller.NewCategoryController(injector)
			CategoryCon.CategoryRouter(CategoryGroup)

			// 主题帖
			PostGroup := ForumGroup.Group("/posts")
			PostCon := controller.NewPostController(injector)
			PostCon.PostRouter(PostGroup)

			// 评论
			CommentGroup := ForumGroup.Group("/comments")
			CommentCon := controller.NewCommentController(injector)
			CommentCon.CommentRouter(CommentGroup)

			// 签到系统
			SigninGroup := ForumGroup.Group("/signin")
			SigninCon := controller.NewSigninController(injector)
			SigninCon.SigninRouter(SigninGroup)
		}

		// 版主接口
		ModeratorGroup := ForumGroup.Group("/moderator")
		ModeratorGroup.Use(saGin.CheckRole(user.RoleModerator.String()))
		ModeratorCon := controller.NewModeratorController(injector)
		ModeratorCon.ModeratorRouter(ModeratorGroup)
	}

	// 管理员接口
	ManageGroup := AuthAPIGroup.Group("/manage")
	ManageGroup.Use(saGin.CheckRole(user.RoleAdmin.String()))
	{
		// 仪表盘
		{
			DashboardGroup := ManageGroup.Group("/dashboard")
			DashboardCon := controller.NewDashboardController(injector)
			DashboardCon.DashboardRouter(DashboardGroup)
		}

		// 用户管理
		{
			UserManageGroup := ManageGroup.Group("/users")
			UserManageCon := controller.NewUserManageController(injector)
			UserManageCon.UserManageRouter(UserManageGroup)
		}

		// 版块管理
		{
			CategoryManageGroup := ManageGroup.Group("/categories")
			CategoryManageCon := controller.NewCategoryManageController(injector)
			CategoryManageCon.CategoryManageRouter(CategoryManageGroup)
		}

		// 帖子管理
		{
			PostManageGroup := ManageGroup.Group("/posts")
			PostManageCon := controller.NewPostManageController(injector)
			PostManageCon.PostManageRouter(PostManageGroup)
		}

		// 评论管理
		{
			CommentManageGroup := ManageGroup.Group("/comments")
			CommentManageCon := controller.NewCommentManageController(injector)
			CommentManageCon.CommentManageRouter(CommentManageGroup)
		}

		// TODO 举报管理
	}

	// 超级管理接口
	SuperManageGroup := AuthAPIGroup.Group("/super/manage")
	SuperManageGroup.Use(saGin.CheckRole(user.RoleSuperAdmin.String()))
	{
		// 性能监控
		{
			PerformanceGroup := SuperManageGroup.Group("/performance")
			PerformanceCon := controller.NewPerformanceController(injector)
			PerformanceCon.PerformanceRouter(PerformanceGroup)
		}

		// 设置管理（统一的设置控制器，包含所有系统设置）
		SettingsGroup := SuperManageGroup.Group("/settings")
		SettingsCon := controller.NewSettingsController(injector)
		SettingsCon.SettingsRouter(SettingsGroup)

		// OAuth提供商管理
		OAuthGroup := SuperManageGroup.Group("/settings/oauth")
		{
			OAuthProviderCon := controller.NewOAuthProviderController(injector)
			OAuthProviderCon.OAuthProviderRouter(OAuthGroup)
		}

		// TODO 广告设置
	}

	return Router
}
