package models

import (
	"time"

	"github.com/lib/pq"
)

// Enforcer interface
type Enforcer interface {
	LoadPolicy() error
	SavePolicy() error
	RefreshPolicy()
	GetAllRoles() []CasbinRole
	GetRoles(offset, limit int) ([]CasbinRole, int)
	GetRole(id uint) (*CasbinRole, error)
	CreateRole(role *CasbinRole) error	
	SaveRole(id uint, permissionIDs []uint) error
	DeleteRole(id uint) error
	GetPermissions() []CasbinPermission
	GetPermissionsWithoutEmpty() []CasbinPermission
	GetChildPermissions(parent uint) []CasbinPermission
	CreatePermission(p *CasbinPermission) error
	DeletePermission(pid uint) error	
	GetUser(id int64) (*CasbinUser, error)
	SaveUser(u *CasbinUser, roles []uint) error
	Enforce(user, resource, action string) bool
}

// Model base structure
type Model struct {
	ID        uint       `json:"ID" gorm:"primary_key"`
	CreatedAt time.Time  `json:"create_at"`
	UpdatedAt time.Time  `json:"update_at"`
	DeletedAt *time.Time `json:"-" sql:"index"`	
}

// CasbinUser represents a casbin user
type CasbinUser struct {
	ID    int64   			`gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name  string 				`gorm:"not null;unique"`
	Roles pq.Int64Array `gorm:"type:integer[]"`
}

// CasbinRole represents casbin role 
type CasbinRole struct {
	Model
	Name string `json:"name" gorm:"not null"`
	Permissions []CasbinPermission `json:"permissions" gorm:"many2many:casbin_role_permission"`
}

// CasbinPermission represents casbin permission
type CasbinPermission struct {
	Model
	Name     string `json:"name" gorm:"not null"`
	Parent   uint   `json:"parent" gorm:"index"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Children []CasbinPermission `json:"children" gorm:"-"`
}

// HasPermission check whether role having permission
func (r *CasbinRole) HasPermission(id uint) bool {
	for i := range r.Permissions {
		if r.Permissions[i].ID == id {
			return true
		}
	}
	return false
}

// GetCasbinAllRoles get all roles in database
func GetCasbinAllRoles() ([]CasbinRole, error) {
	var roles []CasbinRole
	err := gormDB.Find(&roles).Error
	return roles, err
}

// GetCasbinRoles list roles in database
func GetCasbinRoles(offset, limit int) ([]CasbinRole, int) {
	var count int	
	var roles []CasbinRole
	gormDB.Offset(offset).Limit(limit).Order("id asc").Find(&roles).Count(&count)
	return roles, count
}

// GetCasbinRole get role by id
func GetCasbinRole(id uint) (*CasbinRole, error) {
	role := &CasbinRole{}
	err := gormDB.Preload("Permissions").Order("id asc").First(role, id).Error
	return role, err
}

// GetCasbinUser get user data 
func GetCasbinUser(id int64) (*CasbinUser, error) {
	user := &CasbinUser{}
	err := gormDB.First(user, id).Error
	return user, err
}

// SaveCasbinUser get user data 
func SaveCasbinUser(u *CasbinUser, roles []uint) error {
	u.Roles = make(pq.Int64Array, len(roles))
	for i := range roles {
		u.Roles[i] = int64(roles[i])
	}
	return gormDB.Save(u).Error
}

// CreateCasbinRole create new role
func CreateCasbinRole(role *CasbinRole) error {
	return gormDB.Create(role).Error
}

// SaveCasbinRole save role
func SaveCasbinRole(id uint, permissionIDs []uint) error {
	permissions := make([]CasbinPermission, len(permissionIDs))
	for i := range permissionIDs {
		permissions[i] = CasbinPermission{Model: Model{ID: permissionIDs[i]}}
	}
	return gormDB.Model(&CasbinRole{Model: Model{ID: id}}).Association("Permissions").Replace(permissions).Error
}

// DeleteCasbinRole delete a role from database
func DeleteCasbinRole(id uint) error {
	return gormDB.Delete(&CasbinRole{Model: Model{ID: id}}).Error
}

// GetCasbinPermissions get all permissions by hierarchy mode
func GetCasbinPermissions() ([]CasbinPermission, error) {
	var permissions []CasbinPermission
	err := gormDB.Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	// find root permissions
	roots := make([]CasbinPermission, 0)
	for i := range permissions {
		if permissions[i].Parent == 0 {
			roots = append(roots, permissions[i])
		}
	}
	// build children permission of root
	for i := range roots {
		roots[i].Children = make([]CasbinPermission, 0)
		for j := range permissions {
			if permissions[j].Parent == roots[i].ID {
				roots[i].Children = append(roots[i].Children, permissions[j])
			}
		}
	}
	return roots, nil
}

// GetCasbinPermissionsWithoutEmpty load all permission groups withoud empty group
func GetCasbinPermissionsWithoutEmpty() []CasbinPermission {
	roots, err := GetCasbinPermissions()
	if err == nil {
		nonEmptyRoots := make([]CasbinPermission, 0)
		for i := range roots {
			if len(roots[i].Children) > 0 {
				nonEmptyRoots = append(nonEmptyRoots, roots[i])
			}
		}
		return nonEmptyRoots
	}
	return []CasbinPermission{}
}

// CreateCasbinRootPermission create new group
func CreateCasbinRootPermission(p *CasbinPermission) error {
	return gormDB.Create(p).Error
}

// DeleteCasbinRootPermission delete group from database
func DeleteCasbinRootPermission(root uint) error {	
	tx := gormDB.Begin()
	err := tx.Where("parent = ?", root).Delete(&CasbinPermission{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}	

	err = tx.Where("id = ?", root).Delete(&CasbinPermission{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return tx.Error
}

// AppendCasbinPermissionToRoot add a new permission to group
func AppendCasbinPermissionToRoot(root uint, p *CasbinPermission) error {
	p.Parent = root
	return gormDB.Save(p).Error	
}

// DeleteCasbinPermissionFromRoot delete a permission from group
func DeleteCasbinPermissionFromRoot(root uint, permission uint) error {
	return gormDB.Where("id = ?", permission).Delete(&CasbinPermission{}).Error
}

// GetCasbinPermissionsByRoot get permissions by root id
func GetCasbinPermissionsByRoot(root uint) []CasbinPermission {
	var permissons []CasbinPermission
	gormDB.Where("parent = ?", root).Find(&permissons)
	return permissons
}