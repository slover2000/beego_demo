package internal

import (
	"sync"
	"github.com/lib/pq"
	"github.com/jinzhu/gorm"
	"github.com/slover2000/beego_demo/models"
)

// SyncedEnforcer goroutine safed enforcer
type SyncedEnforcer struct {
	db 				*gorm.DB
	model     *Model
	lock 			sync.RWMutex	
}

// NewSyncedEnforcer create a SyncedEnforcer object 
func NewSyncedEnforcer(db *gorm.DB, autoLoad bool) models.Enforcer {
	return &SyncedEnforcer{db: db, model:NewModel(autoLoad)}
}

func (e *SyncedEnforcer) LoadPolicy() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	users := e.GetAllUsers()
	roles := e.GetAllRoles()
	permissions := e.GetAllChildPermissions()
	return e.model.Init(users, roles, permissions)	
}

func (e *SyncedEnforcer) SavePolicy() error {
	return nil
}

func (e *SyncedEnforcer) RefreshPolicy() {
	e.lock.Lock()
	defer e.lock.Unlock()
	
	users := e.GetAllUsers()
	roles := e.GetAllRoles()
	permissions := e.GetAllChildPermissions()
	e.model.Refresh(users, roles, permissions)
}

func (e *SyncedEnforcer) GetAllRoles() []models.CasbinRole {
	var roles []models.CasbinRole
	if err := e.db.Find(&roles).Error; err == nil {
		return roles
	}
	return []models.CasbinRole{}
}

func (e *SyncedEnforcer) GetRoles(offset, limit int) ([]models.CasbinRole, int) {
	var count int	
	var roles []models.CasbinRole
	e.db.Offset(offset).Limit(limit).Order("id asc").Find(&roles).Count(&count)
	return roles, count
}

func (e *SyncedEnforcer) GetRole(id uint) (*models.CasbinRole, error) {
	role := &models.CasbinRole{}
	err := e.db.Preload("Permissions").Order("id asc").First(role, id).Error
	return role, err
}

func (e *SyncedEnforcer) CreateRole(role *models.CasbinRole) error	{
	e.lock.Lock()
	defer e.lock.Unlock()
			
	if err := e.db.Create(role).Error; err != nil {
		return err
	}
	if len(role.Permissions) > 0 {
		permissions := make([]uint, len(role.Permissions))
		for _, p := range role.Permissions {
			permissions = append(permissions, p.ID)
		}
		e.model.UpdateRole(role.ID, permissions)
	}	
	return nil
}

func (e *SyncedEnforcer) SaveRole(id uint, permissionIDs []uint) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	permissions := make([]models.CasbinPermission, len(permissionIDs))
	for i := range permissionIDs {
		permissions[i] = models.CasbinPermission{Model: models.Model{ID: permissionIDs[i]}}
	}
	err := e.db.Model(&models.CasbinRole{Model: models.Model{ID: id}}).Association("Permissions").Replace(permissions).Error
	if err == nil {
		e.model.UpdateRole(id, permissionIDs)
	}
	return err
}

func (e *SyncedEnforcer) DeleteRole(id uint) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	err := e.db.Delete(&models.CasbinRole{Model: models.Model{ID: id}}).Error
	if err == nil {
		e.model.RemoveRole(id)
	}
	return err
}

func (e *SyncedEnforcer) GetPermissions() []models.CasbinPermission {
	var permissions []models.CasbinPermission
	err := e.db.Find(&permissions).Error
	if err != nil {
		return permissions
	}
	// find root permissions
	roots := make([]models.CasbinPermission, 0)
	for i := range permissions {
		if permissions[i].Parent == 0 {
			roots = append(roots, permissions[i])
		}
	}
	// build children permission of root
	for i := range roots {
		roots[i].Children = make([]models.CasbinPermission, 0)
		for j := range permissions {
			if permissions[j].Parent == roots[i].ID {
				roots[i].Children = append(roots[i].Children, permissions[j])
			}
		}
	}
	return roots
}

func (e *SyncedEnforcer) GetPermissionsWithoutEmpty() []models.CasbinPermission {
	roots := e.GetPermissions()
	nonEmptyRoots := make([]models.CasbinPermission, 0)
	for i := range roots {
		if len(roots[i].Children) > 0 {
			nonEmptyRoots = append(nonEmptyRoots, roots[i])
		}
	}
	return nonEmptyRoots
}

func (e *SyncedEnforcer) GetAllChildPermissions() []models.CasbinPermission {
	var permissons []models.CasbinPermission
	e.db.Where("parent <> 0").Find(&permissons)
	return permissons
}

func (e *SyncedEnforcer) GetChildPermissions(parent uint) []models.CasbinPermission {
	var permissons []models.CasbinPermission
	e.db.Where("parent = ?", parent).Find(&permissons)
	return permissons
}

func (e *SyncedEnforcer) CreatePermission(p *models.CasbinPermission) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	err := e.db.Create(p).Error
	if err == nil {
		e.model.UpdatePermissions(p.ID, p)
	}
	return err
}

func (e *SyncedEnforcer) DeletePermission(pid uint) error	{
	e.lock.Lock()
	defer e.lock.Unlock()

	tx := e.db.Begin()
	err := tx.Where("parent = ?", pid).Delete(&models.CasbinPermission{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}	

	err = tx.Where("id = ?", pid).Delete(&models.CasbinPermission{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	if tx.Error == nil {
		e.model.RemovePermission(pid)
	}
	return tx.Error
}

func (e *SyncedEnforcer) GetAllUsers() []models.CasbinUser {
	users := make([]models.CasbinUser, 0)
	if err := e.db.Find(&users).Error; err == nil {
		return users
	}
	return []models.CasbinUser{}
}

func (e *SyncedEnforcer) GetUser(id int64) (*models.CasbinUser, error) {
	user := &models.CasbinUser{}
	err := e.db.First(user, id).Error
	return user, err
}

func (e *SyncedEnforcer) SaveUser(u *models.CasbinUser, roles []uint) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	u.Roles = make(pq.Int64Array, len(roles))
	for i := range roles {
		u.Roles[i] = int64(roles[i])
	}
	err := e.db.Save(u).Error
	if err == nil {
		e.model.UpdateUser(u.Name, roles)
	}
	return err
}

func (e *SyncedEnforcer) Enforce(user, resource, action string) bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.model.HasPermission(user, resource, action)
}