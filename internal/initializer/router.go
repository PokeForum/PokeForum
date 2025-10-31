package initializer

import (
	"net/http"

	_ "github.com/PokeForum/PokeForum/docs"
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/controller"
	"github.com/PokeForum/PokeForum/internal/middleware"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
	"github.com/PokeForum/PokeForum/internal/service"
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InjectorSrv(injector *do.Injector) {
	// 注册 CacheService - 最先注册，方便所有Service使用
	do.Provide(injector, func(i *do.Injector) (cache.ICacheService, error) {
		return cache.NewRedisCacheService(configs.Cache, configs.Log), nil
	})

	// 注册 SettingsService - 需要先注册，方便其他Service依赖它
	do.Provide(injector, func(i *do.Injector) (service.ISettingsService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewSettingsService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 AuthService - 现在依赖SettingsService
	do.Provide(injector, func(i *do.Injector) (service.IAuthService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		settingsService, err := do.Invoke[service.ISettingsService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewAuthService(configs.DB, cacheService, configs.Log, settingsService), nil
	})
	// 注册 UserManageService
	do.Provide(injector, func(i *do.Injector) (service.IUserManageService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewUserManageService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 CategoryManageService
	do.Provide(injector, func(i *do.Injector) (service.ICategoryManageService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewCategoryManageService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 CategoryService
	do.Provide(injector, func(i *do.Injector) (service.ICategoryService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewCategoryService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 PostManageService
	do.Provide(injector, func(i *do.Injector) (service.IPostManageService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewPostManageService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 CommentManageService
	do.Provide(injector, func(i *do.Injector) (service.ICommentManageService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewCommentManageService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 DashboardService
	do.Provide(injector, func(i *do.Injector) (service.IDashboardService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewDashboardService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 ModeratorService
	do.Provide(injector, func(i *do.Injector) (service.IModeratorService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewModeratorService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 PostService
	do.Provide(injector, func(i *do.Injector) (service.IPostService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewPostService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 CommentService
	do.Provide(injector, func(i *do.Injector) (service.ICommentService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewCommentService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 UserProfileService
	do.Provide(injector, func(i *do.Injector) (service.IUserProfileService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewUserProfileService(configs.DB, cacheService, configs.Log), nil
	})
}

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

	// 注册服务到注入器
	InjectorSrv(injector)

	// 存活检测
	Router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 注册Swagger
	if configs.Debug == true {
		Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := Router.Group("/api/v1")

	// 认证校验
	AuthGroup := api.Group("/auth")
	AuthCon := controller.NewAuthController(injector)
	AuthCon.AuthRouter(AuthGroup)

	// 添加登录校验
	AuthAPIGroup := api.Group("")
	AuthAPIGroup.Use(saPlugin.AuthMiddleware())

	// 论坛接口
	ForumGroup := AuthAPIGroup.Group("")
	ForumGroup.Use()
	{
		// 用户
		{
			// 个人中心
			{
				ProfileGroup := ForumGroup.Group("/profile")
				ProfileCon := controller.NewUserProfileController(injector)
				ProfileCon.UserProfileRouter(ProfileGroup)

				/*
					- 屏蔽用户可阅览楼主帖子
					- 禁止站内信楼主
					- 禁止评论楼主帖子
				*/
				// TODO 屏蔽用户
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

			// TODO 排行榜
			{
				// 阅读榜(总榜/月榜/周榜)
				// 主题数
				// 评论数
				// 积分榜
				// TODO 邀请数
				// TODO 财富榜
				// TODO 活跃榜(签到榜)
			}

			// 版块
			CategoryGroup := ForumGroup.Group("/categories")
			CategoryCon := controller.NewCategoryController(injector)
			CategoryCon.CategoryRouter(CategoryGroup)

			// 主题帖
			/*
				发帖前需要验证邮箱, 可通过开关配置
			*/
			// 发布新帖
			// 保存草稿
			// 编辑帖子(每三分钟可操作一次)
			// 私有帖子(双向, 每三日可操作一次)
			// 点赞帖子(单向, 不可取消点赞)
			// 点踩帖子(单向, 不可取消点赞)
			// 收藏帖子
			PostGroup := ForumGroup.Group("/posts")
			PostCon := controller.NewPostController(injector)
			PostCon.PostRouter(PostGroup)

			// 评论
			/*
				发评论前需要验证邮箱, 可通过开关配置
			*/
			// 发布评论
			// 编辑评论
			// 点赞评论(单向, 不可取消点赞)
			// 点踩评论(单向, 不可取消点赞)
			// 获取评论列表
			// 获取评论详情
			CommentGroup := ForumGroup.Group("/comments")
			CommentCon := controller.NewCommentController(injector)
			CommentCon.CommentRouter(CommentGroup)
		}

		// 版主接口
		ModeratorGroup := ForumGroup.Group("/moderator")
		ModeratorCon := controller.NewModeratorController(injector)
		ModeratorCon.ModeratorRouter(ModeratorGroup)
	}

	// 管理员接口
	ManageGroup := AuthAPIGroup.Group("/manage")
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
	{
		// 设置管理（统一的设置控制器，包含所有系统设置）
		SettingsGroup := SuperManageGroup.Group("/settings")
		SettingsCon := controller.NewSettingsController(injector)
		SettingsCon.SettingsRouter(SettingsGroup)

		// 功能设置
		{
			/*
				- TODO 第三方登录
					- QQ
					- Telegram
					- Github
					- Apple
					- Google
			*/
		}

		// TODO 广告设置
	}

	return Router
}
