package initializer

import (
	"net/http"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	// (可选项)
	// PID 限流基于实例的 CPU 使用率，通过拒绝一定比例的流量, 将实例的 CPU 使用率稳定在设定的阈值上。
	// 地址: https://github.com/bytedance/pid_limits
	// Router.Use(adaptive.PlatoMiddlewareGinDefault(0.8))

	// 存活检测
	Router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 注册Swagger
	//if config.Config.App.Mode == gin.DebugMode {
	//	Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//}

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
