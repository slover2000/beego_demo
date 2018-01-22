package internal

import (
	"github.com/slover2000/beego_demo/models"
)

type userCache struct {
	roles       []uint
	permissions []models.CasbinPermission
}

// Model ...
type Model struct {
	autoRefresh bool
	Permissions map[uint]models.CasbinPermission
	Roles				map[uint][]uint
	Users       map[string]*userCache
}

func NewModel(autoRefresh bool) *Model {
	return &Model{
		autoRefresh: autoRefresh,
		Permissions: make(map[uint]models.CasbinPermission),
		Roles: make(map[uint][]uint),
		Users: make(map[string]*userCache),
	}
}

func (m *Model) Init(users []models.CasbinUser, roles []models.CasbinRole, permissions []models.CasbinPermission) error {
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

func (m *Model) Refresh(users []models.CasbinUser, roles []models.CasbinRole, permissions []models.CasbinPermission) {
	m.Permissions = make(map[uint]models.CasbinPermission)
	m.Roles = make(map[uint][]uint)
	m.Users = make(map[string]*userCache)
	m.Init(users, roles, permissions)
}

func (m *Model) HasPermission(user, resource, action string) bool {
	if cache, ok := m.Users[user]; ok {
		for i := range cache.permissions {
			p := cache.permissions[i]
			if KeyMatch(resource, p.Resource) && RegexMatch(action, p.Action) {
				return true
			}
		}
	}
	return false
}

func (m *Model) buildPermissions(roles []uint) []models.CasbinPermission {
	permissions := make(map[uint]models.CasbinPermission)
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

	buildPermissions := make([]models.CasbinPermission, 0, len(permissions))
	for _, v := range permissions {
		buildPermissions = append(buildPermissions, v)
	}
	return buildPermissions
}
 
func (m *Model) UpdateUser(user string, roles []uint) {
	if cache, ok := m.Users[user]; ok {
		cache.roles = roles
		if m.autoRefresh {
			cache.permissions = m.buildPermissions(roles)
		}
	}
}

func (m *Model) RefreshAllUsers() {
	for name, cache := range m.Users {
		newPermissions := make([]models.CasbinPermission, 0, len(cache.permissions))
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

func (m *Model) RemoveUser(user string) {
	delete(m.Users, user)
}

func (m *Model) UpdateRole(id uint, permissions []uint) {	
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

func (m *Model) RemoveRole(id uint) {
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

func (m *Model) UpdatePermissions(id uint, permission *models.CasbinPermission) {
	m.Permissions[id] = *permission
	// update impacting users
	if m.autoRefresh {
		m.RefreshAllUsers()
	}	
}

func (m *Model) RemovePermission(id uint) {
	delete(m.Permissions, id)
	// update impacting users
	if m.autoRefresh {
		m.RefreshAllUsers()
	}	
}