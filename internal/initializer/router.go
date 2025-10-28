package initializer

import (
	"net/http"

	_ "github.com/PokeForum/PokeForum/docs"
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/middleware"
	satoken "github.com/PokeForum/PokeForum/internal/pkg/sa-token"
	saGin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Routers() *gin.Engine {
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

	//api := Router.Group("/api")

	//// 健康检查
	//HealthyGroup := api.Group("/")
	//HealthyCon := controller.NewHealthyController()
	//HealthyCon.HealthyRouter(HealthyGroup)
	//
	//// 公开服务（应用令牌桶限速）
	//OpenGroup := api.Group("/open")
	//OpenGroup.Use(middleware.OpenAPIRateLimit()) // 应用令牌桶限速中间件
	//OpenCon := controller.NewOpenController()
	//OpenCon.OpenRouter(OpenGroup)
	//
	//// 认证校验
	//AuthGroup := api.Group("/auth")
	//AuthCon := controller.NewAuthController()
	//AuthCon.AuthRouter(AuthGroup)
	//
	//// 认证接口
	//authAPI := api.Group("")
	//authAPI.Use(middleware.JWTAuth()) // 校验请求认证
	//{
	//	// 认证
	//	AuthRequiredGroup := authAPI.Group("/auth")
	//	AuthRequiredCon := controller.NewAuthRequiredController()
	//	AuthRequiredCon.AuthRequiredRouter(AuthRequiredGroup)
	//
	//	// 仪表盘
	//	DashboardGroup := authAPI.Group("/dashboard")
	//	DashboardCon := controller.NewDashboardController()
	//	DashboardCon.DashboardRouter(DashboardGroup)
	//}

	return Router
}
