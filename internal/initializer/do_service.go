package initializer

import (
	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/PokeForum/PokeForum/internal/pkg/cache"
	"github.com/PokeForum/PokeForum/internal/service"
	"github.com/samber/do"
)

func InjectorSrv(injector *do.Injector) {
	// 注册 CacheService
	do.Provide(injector, func(i *do.Injector) (cache.ICacheService, error) {
		return cache.NewRedisCacheService(configs.Cache, configs.Log), nil
	})

	// 注册 SettingsService
	do.Provide(injector, func(i *do.Injector) (service.ISettingsService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewSettingsService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 AuthService
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
		settingsService, err := do.Invoke[service.ISettingsService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewUserProfileService(configs.DB, cacheService, configs.Log, settingsService), nil
	})
	// 注册 RankingService
	do.Provide(injector, func(i *do.Injector) (service.IRankingService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewRankingService(configs.DB, cacheService, configs.Log), nil
	})
	// 注册 OAuthProviderService
	do.Provide(injector, func(i *do.Injector) (service.IOAuthProviderService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewOAuthProviderService(configs.DB, cacheService, configs.Log), nil
	})

	// 注册 RedisLock
	do.Provide(injector, func(i *do.Injector) (*cache.RedisLock, error) {
		return cache.NewRedisLock(configs.Cache, configs.Log), nil
	})

	// 注册 SigninAsyncTask
	do.Provide(injector, func(i *do.Injector) (*service.SigninAsyncTask, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		return service.NewSigninAsyncTask(configs.DB, cacheService, configs.Log), nil
	})

	// 注册 BlacklistService
	do.Provide(injector, func(i *do.Injector) (service.IBlacklistService, error) {
		return service.NewBlacklistService(configs.DB, configs.Log), nil
	})

	// 注册 SigninService
	do.Provide(injector, func(i *do.Injector) (service.ISigninService, error) {
		cacheService, err := do.Invoke[cache.ICacheService](injector)
		if err != nil {
			return nil, err
		}
		redisLock, err := do.Invoke[*cache.RedisLock](injector)
		if err != nil {
			return nil, err
		}
		settingsService, err := do.Invoke[service.ISettingsService](injector)
		if err != nil {
			return nil, err
		}
		asyncTask, err := do.Invoke[*service.SigninAsyncTask](injector)
		if err != nil {
			return nil, err
		}
		return service.NewSigninService(configs.DB, cacheService, redisLock, configs.Log, settingsService, asyncTask), nil
	})
}
