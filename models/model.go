package models

import (
	"github.com/slover2000/beego_demo/models/internal"
)

const (
	AdminRoleID   = 0
	AdminRoleName = "admin"	
)

type userCache struct {
	hasAdminRole  bool
	roles        []uint
	permissions  []CasbinPermission
}

// EnforcerModel ...
type EnforcerModel struct {
	autoRefresh     bool
	Permissions     map[uint]CasbinPermission
	RolePermissions	map[uint][]uint
	RoleNames       map[uint]string
	Users           map[string]*userCache
}

func NewModel(autoRefresh bool) *EnforcerModel {
	return &EnforcerModel{
		autoRefresh: autoRefresh,
		Permissions: make(map[uint]CasbinPermission),
		RolePermissions: make(map[uint][]uint),
		RoleNames: make(map[uint]string),
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
			m.RolePermissions[r.ID] = rolePermissions
			m.RoleNames[r.ID] = r.Name
		}

		for _, user := range users {
			hasAdminRole := false
			roles := make([]uint, len(user.Roles))
			for i, id := range user.Roles {
				roles[i] = uint(id)
				if id == AdminRoleID {
					hasAdminRole = true					
				}
			}
			m.Users[user.Name] = &userCache{hasAdminRole: hasAdminRole, roles: roles, permissions: m.buildPermissions(roles)}
		}
		return nil
}

func (m *EnforcerModel) Refresh(users []CasbinUser, roles []CasbinRole, permissions []CasbinPermission) {
	m.Permissions = make(map[uint]CasbinPermission)
	m.RolePermissions = make(map[uint][]uint)
	m.RoleNames = make(map[uint]string)
	m.Users = make(map[string]*userCache)
	m.Init(users, roles, permissions)
}

func (m *EnforcerModel) GetUserRoleNames(name string) []string {
	if cache, ok := m.Users[name]; ok {
		names := make([]string, len(cache.roles))
		for i, id := range cache.roles {
			if id == AdminRoleID {
				names[i] = AdminRoleName
				continue
			} else {
				if name, ok := m.RoleNames[id]; ok {
					names[i] = name
				}
			}
		}
		return names
	}
	return []string{}
}

func (m *EnforcerModel) HasPermission(user, resource, action string) bool {
	if cache, ok := m.Users[user]; ok {
		if cache.hasAdminRole {
			return true
		}

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
		if rolePermissions, ok := m.RolePermissions[roles[i]]; ok {
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
		hasAdminRole := false
		for i := range roles {
			if roles[i] == AdminRoleID {
				hasAdminRole = true
				break
			}
		}
		cache.hasAdminRole = hasAdminRole
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
	m.RolePermissions[id] = permissions
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
	delete(m.RolePermissions, id)
	delete(m.RoleNames, id)
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

func (m *EnforcerModel) AddPermission(id uint, permission *CasbinPermission) {
	m.Permissions[id] = *permission
}

func (m *EnforcerModel) UpdatePermission(id uint, permission *CasbinPermission) {
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