package initializer

import (
	"net/http"

	_ "github.com/PokeForum/PokeForum/docs"
	"github.com/PokeForum/PokeForum/ent/user"
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
	// 注册 EmailService
	do.Provide(injector, func(i *do.Injector) (service.IEmailService, error) {
		return service.NewEmailService(configs.DB, configs.Cache, configs.Log), nil
	})
}

func Routers(injector *do.Injector) *gin.Engine {
	// 设置模式
	if !configs.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	Router := gin.New()
	Router.Use(middleware.Logger())
	Router.Use(middleware.Recovery())

	// 跨域配置
	Router.Use(cors.New(middleware.CorsConfig))

	// 设置SaToken
	saManager := satoken.NewSaToken()
	saGin.SetManager(saManager)

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

	// 论坛接口
	ForumGroup := api.Group("")
	ForumGroup.Use(saGin.CheckRole(user.RoleUser.String()))
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
		ModeratorGroup.Use(saGin.CheckRole(user.RoleModerator.String()))
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
	ManageGroup := api.Group("/manage")
	ManageGroup.Use(saGin.CheckRole(user.RoleAdmin.String()))
	{
		// 仪表盘
		{
		}

		// 用户管理
		{
		}

		// 版块管理
		{
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
	SuperManageGroup := api.Group("/super/manage")
	SuperManageGroup.Use(saGin.CheckRole(user.RoleSuperAdmin.String()))
	{
		// 常规设置
		{
			/*
				- 网站Logo
				- 网站图标ICON
				- ICP备案号
				- 公安联网备案号
				- 关闭版权显示
			*/
		}

		// 首页设置
		{
			/*
				- 幻灯片设置
				- 友情链接
				- 合作伙伴
			*/
		}

		// 评论设置
		{
			/*
				- 评论信息显示
				- 检查评论
				- 关键词黑名单
				- 外部审查API
			*/
		}

		// 功能设置
		{
			// 邮箱服务
			{
				EmailGroup := SuperManageGroup.Group("/email")
				EmailCon := controller.NewEmailController(injector)
				EmailCon.EmailRouter(EmailGroup)
			}

			/*
				- TODO 第三方登录
					- QQ
					- Telegram
					- Github
					- Apple
					- Google
			*/
		}

		// SEO设置
		{
			/*
				- 网站关键词
				- 网站描述
			*/
		}

		// TODO 广告设置

		// 代码配置
		{
			/*
				- 页头代码
				- 页脚代码
				- 自定义CSS
			*/
		}

		// 安全设置
		{
			/*
				- 注册控制
			*/
		}
	}

	return Router
}
