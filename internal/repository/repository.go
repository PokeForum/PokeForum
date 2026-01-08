package repository

import (
	"github.com/PokeForum/PokeForum/ent"
)

// Repositories Repository collection | 仓储集合
type Repositories struct {
	User              IUserRepository
	Post              IPostRepository
	Comment           ICommentRepository
	Category          ICategoryRepository
	Settings          ISettingsRepository
	PostAction        IPostActionRepository
	CommentAction     ICommentActionRepository
	Blacklist         IBlacklistRepository
	UserLoginLog      IUserLoginLogRepository
	UserSigninStatus  IUserSigninStatusRepository
	UserSigninLogs    IUserSigninLogsRepository
	OAuthProvider     IOAuthProviderRepository
	CategoryModerator ICategoryModeratorRepository
	UserBalanceLog    IUserBalanceLogRepository
}

// NewRepositories Create repository collection instance | 创建仓储集合实例
func NewRepositories(db *ent.Client) *Repositories {
	return &Repositories{
		User:              NewUserRepository(db),
		Post:              NewPostRepository(db),
		Comment:           NewCommentRepository(db),
		Category:          NewCategoryRepository(db),
		Settings:          NewSettingsRepository(db),
		PostAction:        NewPostActionRepository(db),
		CommentAction:     NewCommentActionRepository(db),
		Blacklist:         NewBlacklistRepository(db),
		UserLoginLog:      NewUserLoginLogRepository(db),
		UserSigninStatus:  NewUserSigninStatusRepository(db),
		UserSigninLogs:    NewUserSigninLogsRepository(db),
		OAuthProvider:     NewOAuthProviderRepository(db),
		CategoryModerator: NewCategoryModeratorRepository(db),
		UserBalanceLog:    NewUserBalanceLogRepository(db),
	}
}
