package models

import (
	"time"

	"github.com/lib/pq"
	"github.com/jinzhu/gorm"
)

type Model struct {
	ID        uint       `json:"ID" gorm:"primary_key"`
	CreatedAt time.Time  `json:"create_at"`
	UpdatedAt time.Time  `json:"update_at"`
	DeletedAt *time.Time `json:"-" sql:"index"`	
}

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

type CasbinGroup struct {
	gorm.Model
	Name     string `json:"name" gorm:"not null"`
	Permissions []CasbinPermission `json:"permissions" gorm:"many2many:casbin_group_permission"`
}

type CasbinPermission struct {
	Model
	Name     string `json:"name" gorm:"not null"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type CasbinRoleResp struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	CreateTime JSONTime `json:"create_time"`
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

// GetCasbinGroups load all permission groups
func GetCasbinGroups() []CasbinGroup {
	var groups []CasbinGroup
	gormDB.Preload("Permissions").Order("id asc").Find(&groups)
	return groups
}

// GetCasbinGroupsWithoutEmpty load all permission groups withoud empty group
func GetCasbinGroupsWithoutEmpty() []CasbinGroup {
	var groups []CasbinGroup
	gormDB.Preload("Permissions").Order("id asc").Find(&groups)
	if len(groups) > 0 {
		filtered := make([]CasbinGroup, 0)
		for i := range groups {
			if len(groups[i].Permissions) > 0 {
				filtered = append(filtered, groups[i])
			}
		}
		return filtered
	}
	return groups
}

// CreateCasbinGroup create new group
func CreateCasbinGroup(group *CasbinGroup) error {
	return gormDB.Create(group).Error
}

// DeleteCasbinGroup delete group from database
func DeleteCasbinGroup(group uint) error {
	g := &CasbinGroup{Model: gorm.Model{ID: group}}
	var permissions []CasbinPermission
	gormDB.Model(g).Association("Permissions").Find(&permissions)
	tx := gormDB.Begin()
	err := tx.Model(g).Association("Permissions").Clear().Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Delete(g).Error
	if err != nil {
		tx.Rollback()
		return err
	}	

	if len(permissions) > 0 {
		pids := make([]uint, len(permissions))
		for i := range permissions {
			pids[i] = permissions[i].ID
		}
		tx.Where("id IN (?)", pids).Delete(CasbinPermission{})
	}
	tx.Commit()
	return tx.Error
}

// AppendCasbinPermissionToGroup add a new permission to group
func AppendCasbinPermissionToGroup(group uint, p *CasbinPermission) error {
	return gormDB.Model(&CasbinGroup{Model: gorm.Model{ID: group}}).Association("Permissions").Append(*p).Error
}

// DeleteCasbinPermissionFromGroup delete a permission from group
func DeleteCasbinPermissionFromGroup(group uint, permission uint) error {
	err := gormDB.Model(&CasbinGroup{Model: gorm.Model{ID: group}}).Association("Permissions").Delete(CasbinPermission{Model: Model{ID: permission}}).Error
	if err != nil {
		return err
	}
	return gormDB.Delete(&CasbinPermission{Model: Model{ID: permission}}).Error
}

// SaveCasbinGroup save group with permission associations
func SaveCasbinGroup(group uint, permissionIDs []uint) error {
	permissions := make([]CasbinPermission, len(permissionIDs))
	for i := range permissionIDs {
		permissions[i] = CasbinPermission{Model: Model{ID: permissionIDs[i]}}
	}
	return gormDB.Model(&CasbinGroup{Model: gorm.Model{ID: group}}).Association("Permissions").Replace(permissions).Error
}

// GetCasbinPermissionsByGroup get permissions by group id
func GetCasbinPermissionsByGroup(group uint) []CasbinPermission {
	var permissons []CasbinPermission
	gormDB.First(&CasbinGroup{Model: gorm.Model{ID: group}}).Association("Permissions").Find(&permissons)
	return permissons
}

// CreateCasbinPermission create a new permission
func CreateCasbinPermission(p *CasbinPermission) error {
	return gormDB.Create(p).Error
}

// DeleteCasbinPermission delete a permission from database
func DeleteCasbinPermission(id uint) error {
	return gormDB.Delete(&CasbinPermission{Model: Model{ID: id}}).Error
}