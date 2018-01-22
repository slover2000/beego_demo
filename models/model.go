package models

import (
	"github.com/slover2000/beego_demo/models/internal"
)

type userCache struct {
	roles       []uint
	permissions []CasbinPermission
}

// EnforcerModel ...
type EnforcerModel struct {
	autoRefresh bool
	Permissions map[uint]CasbinPermission
	Roles				map[uint][]uint
	Users       map[string]*userCache
}

func NewModel(autoRefresh bool) *EnforcerModel {
	return &EnforcerModel{
		autoRefresh: autoRefresh,
		Permissions: make(map[uint]CasbinPermission),
		Roles: make(map[uint][]uint),
		Users: make(map[string]*userCache),
	}
}

func (m *EnforcerModel) Init(users []CasbinUser, roles []CasbinRole, permissions []CasbinPermission) error {
		for _, p := range permissions {
			m.Permissions[p.ID] = p
		}

		for _, r := range roles {
			rolePermissions := make([]uint, len(r.Permissions))
			for _, p := range r.Permissions {
				rolePermissions = append(rolePermissions, p.ID)
			}
			m.Roles[r.ID] = rolePermissions
		}

		for _, user := range users {			
			roles := make([]uint, len(user.Roles))
			for _, id := range user.Roles {
				roles = append(roles, uint(id))
			}
			m.Users[user.Name] = &userCache{roles: roles, permissions: m.buildPermissions(roles)}
		}
		return nil
}

func (m *EnforcerModel) Refresh(users []CasbinUser, roles []CasbinRole, permissions []CasbinPermission) {
	m.Permissions = make(map[uint]CasbinPermission)
	m.Roles = make(map[uint][]uint)
	m.Users = make(map[string]*userCache)
	m.Init(users, roles, permissions)
}

func (m *EnforcerModel) HasPermission(user, resource, action string) bool {
	if cache, ok := m.Users[user]; ok {
		for i := range cache.permissions {
			p := cache.permissions[i]
			if internal.KeyMatch(resource, p.Resource) && internal.RegexMatch(action, p.Action) {
				return true
			}
		}
	}
	return false
}

func (m *EnforcerModel) buildPermissions(roles []uint) []CasbinPermission {
	permissions := make(map[uint]CasbinPermission)
	for i := range roles {
		if rolePermissions, ok := m.Roles[roles[i]]; ok {
			for j := range rolePermissions {
				pid := rolePermissions[j]
				if permission, ok := m.Permissions[pid]; ok {
					permissions[pid] = permission
				}
			}
		}
	}

	buildPermissions := make([]CasbinPermission, 0, len(permissions))
	for _, v := range permissions {
		buildPermissions = append(buildPermissions, v)
	}
	return buildPermissions
}
 
func (m *EnforcerModel) UpdateUser(user string, roles []uint) {
	if cache, ok := m.Users[user]; ok {
		cache.roles = roles
		if m.autoRefresh {
			cache.permissions = m.buildPermissions(roles)
		}
	}
}

func (m *EnforcerModel) RefreshAllUsers() {
	for name, cache := range m.Users {
		newPermissions := make([]CasbinPermission, 0, len(cache.permissions))
		for _, permission := range cache.permissions {
			if p, ok := m.Permissions[permission.ID]; ok {
				newPermissions = append(newPermissions, p)
			}
		}		
		if cache, ok := m.Users[name]; ok {
			cache.permissions = newPermissions
		}		 
	}	
}

func (m *EnforcerModel) RemoveUser(user string) {
	delete(m.Users, user)
}

func (m *EnforcerModel) UpdateRole(id uint, permissions []uint) {	
	m.Roles[id] = permissions
	// update impacting users
	for name, cache := range m.Users {
		for _, roleID := range cache.roles {
			if roleID == id {
				// update user permissions
				m.UpdateUser(name, cache.roles)
				break				
			}
		}
	}
}

func (m *EnforcerModel) RemoveRole(id uint) {
	delete(m.Roles, id)
	// update impacting users
	for name, cache := range m.Users {
		for i, roleID := range cache.roles {
			if roleID == id {
				// update user permissions
				newRoles := append(cache.roles[0:i], cache.roles[i+1:]...)
				m.UpdateUser(name, newRoles)
				break
			}
		}
	}
}

func (m *EnforcerModel) UpdatePermissions(id uint, permission *CasbinPermission) {
	m.Permissions[id] = *permission
	// update impacting users
	if m.autoRefresh {
		m.RefreshAllUsers()
	}	
}

func (m *EnforcerModel) RemovePermission(id uint) {
	delete(m.Permissions, id)
	// update impacting users
	if m.autoRefresh {
		m.RefreshAllUsers()
	}	
}