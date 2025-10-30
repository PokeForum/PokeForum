package initializer

import (
	"net/http"

	_ "github.com/PokeForum/PokeForum/docs"
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/controller"
	"github.com/PokeForum/PokeForum/internal/middleware"
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
	// 注册 AuthService
	do.Provide(injector, func(i *do.Injector) (service.IAuthService, error) {
		return service.NewAuthService(configs.DB, configs.Cache, configs.Log), nil
	})
	// 注册 SettingsService
	do.Provide(injector, func(i *do.Injector) (service.ISettingsService, error) {
		return service.NewSettingsService(configs.DB, configs.Cache, configs.Log), nil
	})
	// 注册 UserManageService
	do.Provide(injector, func(i *do.Injector) (service.IUserManageService, error) {
		return service.NewUserManageService(configs.DB, configs.Cache, configs.Log), nil
	})
	// 注册 CategoryManageService
	do.Provide(injector, func(i *do.Injector) (service.ICategoryManageService, error) {
		return service.NewCategoryManageService(configs.DB, configs.Cache, configs.Log), nil
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
				// 个人信息/概览
				// 主题贴 / TODO 已购
				// 评论
				// 收藏
				// TODO 粉丝
				// 修改密码
				// 修改头像
				// 修改用户名(每七日可操作一次)
				// 修改邮箱

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

			// 主题帖
			{
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
			}

			// 评论
			{
				/*
					发评论前需要验证邮箱, 可通过开关配置
				*/
				// 发布评论
				// 编辑评论
				// 点赞评论
				// 点踩评论
			}
		}

		// 版主接口
		ModeratorGroup := ForumGroup.Group("/moderator")
		ModeratorGroup.Use()
		{
			// 主题帖
			{
				// 封禁帖子
				// 编辑帖子
				// 移动帖子(迁移板块)
				// 精华帖子
				// 锁定帖子
				// 指定帖子

			}

			// 版块
			{
				// 版块编辑
				// 版块公告
			}
		}
	}

	// 管理员接口
	ManageGroup := AuthAPIGroup.Group("/manage")
	{
		// 仪表盘
		{
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
		}

		// 评论管理
		{
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
