package sa_token

import (
	"github.com/PokeForum/PokeForum/ent/user"
)

// GetUserRole 获取用户身份权限
func GetUserRole(role string) []string {
	switch role {
	case user.RoleUser.String(): // 普通用户
		return []string{user.RoleUser.String()}
	case user.RoleModerator.String():
		return []string{
			user.RoleUser.String(),
			user.RoleModerator.String(),
		}
	case user.RoleAdmin.String():
		return []string{
			user.RoleUser.String(),
			user.RoleModerator.String(),
			user.RoleAdmin.String(),
		}
	case user.RoleSuperAdmin.String():
		return []string{
			user.RoleUser.String(),
			user.RoleModerator.String(),
			user.RoleAdmin.String(),
			user.RoleSuperAdmin.String(),
		}
	default:
		// 不存在身份用户
		return []string{}
	}
}
