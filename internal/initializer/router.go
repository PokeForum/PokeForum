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
		return service.NewAuthService(configs.DB, configs.Cache), nil
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

	// 注册前端静态资源
	//SetupWebFrontend(Router)

	api := Router.Group("/api/v1")

	// 认证校验
	AuthGroup := api.Group("/auth")
	AuthCon := controller.NewAuthController(injector)
	AuthCon.AuthRouter(AuthGroup)

	// 论坛接口
	ForumGroup := api.Group("/forum")
	ForumGroup.Use(saGin.CheckRole(user.RoleUser.String()))
	{

		// 版主接口
		ModeratorGroup := ForumGroup.Group("/moderator")
		ModeratorGroup.Use(saGin.CheckRole(user.RoleModerator.String()))
		{

		}
	}

	// 管理员接口
	ManageGroup := api.Group("/manage")
	ManageGroup.Use(saGin.CheckRole(user.RoleAdmin.String()))
	{

	}

	// 超级管理接口
	SuperManageGroup := api.Group("/super/manage")
	SuperManageGroup.Use(saGin.CheckRole(user.RoleSuperAdmin.String()))
	{

	}

	return Router
}
