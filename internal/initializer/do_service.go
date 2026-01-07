package initializer

import (
	"github.com/samber/do"
	"go.uber.org/zap"

	"github.com/PokeForum/PokeForum/ent"
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/service"
)

func InjectorSrv(injector *do.Injector) {
	// Register Infrastructure | 注册基础设施
	do.ProvideValue(injector, configs.Log)
	do.ProvideValue(injector, configs.DB)
	do.ProvideValue(injector, configs.Cache)

	// Register CacheService | 注册 CacheService
	do.Provide(injector, func(i *do.Injector) (cache.ICacheService, error) {
		return cache.NewRedisCacheService(configs.Cache, configs.Log), nil
	})

	// Register SettingsService | 注册 SettingsService
	do.Provide(injector, func(i *do.Injector) (service.ISettingsService, error) {
		cacheService := do.MustInvoke[cache.ICacheService](i)
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		return service.NewSettingsService(db, cacheService, logger), nil
	})
	// Register AuthService | 注册 AuthService
	do.Provide(injector, func(i *do.Injector) (service.IAuthService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		settingsService := do.MustInvoke[service.ISettingsService](i)
		return service.NewAuthService(db, cacheService, logger, settingsService), nil
	})
	// Register UserManageService | 注册 UserManageService
	do.Provide(injector, func(i *do.Injector) (service.IUserManageService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewUserManageService(db, cacheService, logger), nil
	})
	// Register CategoryManageService | 注册 CategoryManageService
	do.Provide(injector, func(i *do.Injector) (service.ICategoryManageService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewCategoryManageService(db, cacheService, logger), nil
	})
	// Register CategoryService | 注册 CategoryService
	do.Provide(injector, func(i *do.Injector) (service.ICategoryService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewCategoryService(db, cacheService, logger), nil
	})
	// Register PostManageService | 注册 PostManageService
	do.Provide(injector, func(i *do.Injector) (service.IPostManageService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewPostManageService(db, cacheService, logger), nil
	})
	// Register CommentManageService | 注册 CommentManageService
	do.Provide(injector, func(i *do.Injector) (service.ICommentManageService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewCommentManageService(db, cacheService, logger), nil
	})
	// Register DashboardService | 注册 DashboardService
	do.Provide(injector, func(i *do.Injector) (service.IDashboardService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewDashboardService(db, cacheService, logger), nil
	})
	// Register ModeratorService | 注册 ModeratorService
	do.Provide(injector, func(i *do.Injector) (service.IModeratorService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewModeratorService(db, cacheService, logger), nil
	})
	// Register PostService | 注册 PostService
	do.Provide(injector, func(i *do.Injector) (service.IPostService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewPostService(db, cacheService, logger), nil
	})
	// Register CommentService | 注册 CommentService
	do.Provide(injector, func(i *do.Injector) (service.ICommentService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewCommentService(db, cacheService, logger), nil
	})
	// Register UserProfileService | 注册 UserProfileService
	do.Provide(injector, func(i *do.Injector) (service.IUserProfileService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		settingsService := do.MustInvoke[service.ISettingsService](i)
		userManageService := do.MustInvoke[service.IUserManageService](i)
		return service.NewUserProfileService(db, cacheService, logger, settingsService, userManageService), nil
	})
	// Register RankingService | 注册 RankingService
	do.Provide(injector, func(i *do.Injector) (service.IRankingService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewRankingService(db, cacheService, logger), nil
	})
	// Register OAuthProviderService | 注册 OAuthProviderService
	do.Provide(injector, func(i *do.Injector) (service.IOAuthProviderService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		return service.NewOAuthProviderService(db, cacheService, logger), nil
	})

	// Register RedisLock | 注册 RedisLock
	do.Provide(injector, func(i *do.Injector) (*cache.RedisLock, error) {
		return cache.NewRedisLock(configs.Cache, configs.Log), nil
	})

	// TaskManager and SigninAsyncTask are injected through do.ProvideValue in server.go | TaskManager 和 SigninAsyncTask 在 server.go 中通过 do.ProvideValue 注入

	// Register BlacklistService | 注册 BlacklistService
	do.Provide(injector, func(i *do.Injector) (service.IBlacklistService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		return service.NewBlacklistService(db, logger), nil
	})

	// Register SigninService | 注册 SigninService
	do.Provide(injector, func(i *do.Injector) (service.ISigninService, error) {
		db := do.MustInvoke[*ent.Client](i)
		logger := do.MustInvoke[*zap.Logger](i)
		cacheService := do.MustInvoke[cache.ICacheService](i)
		redisLock := do.MustInvoke[*cache.RedisLock](i)
		settingsService := do.MustInvoke[service.ISettingsService](i)
		asyncTask := do.MustInvoke[*service.SigninAsyncTask](i)
		return service.NewSigninService(db, cacheService, redisLock, logger, settingsService, asyncTask), nil
	})

	// Register PerformanceService | 注册 PerformanceService
	do.Provide(injector, func(i *do.Injector) (service.IPerformanceService, error) {
		pgDB := PgDB()
		logger := do.MustInvoke[*zap.Logger](i)
		return service.NewPerformanceService(pgDB, configs.Cache, logger), nil
	})
}
