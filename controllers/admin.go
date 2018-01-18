package controllers

import (
	"strconv"
	"strings"
	"html/template"

	"github.com/sirupsen/logrus"

	"github.com/slover2000/beego_demo/models"
)

// AdminController Operations about home
type AdminController struct {
	baseController
}

type RoleData struct {
	ID 	  uint
	Name  string
	Have  bool
}

func (c *AdminController) UserList() {
	c.Data["pageTitle"] = "用户列表"
	c.Data["xsrf_token"] = c.XSRFToken()
	c.renderNestedTemplate("admin/users")
}

func (c *AdminController) GetUsers() {
	page, err := c.GetInt("page")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user": c.userName,
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("invalid page parameter '%s'", c.GetString("page"))
		c.Abort("400")
	}
	limit, err := c.GetInt("limit")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user": c.userName,
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("invalid limit parameter '%s'", c.GetString("limit"))		
		c.Abort("400")
	}

	offset := (page - 1) * limit
	users, total := models.GetUsers(offset, limit)
	userResp := make([]models.UserResp, len(users))
	for i := range users {
		u := users[i]
		userResp[i] = models.UserResp{
			Id: u.Id,
			Name: u.Name,
			CreateTime: models.JSONTime(u.CreateTime),
			UpdateTime: models.JSONTime(u.UpdateTime),
			Profile: models.Profile{
				Gender: u.Profile2.Gender,
				Age: u.Profile2.Age,
				Address: u.Profile2.Address,
				Email: u.Profile2.Email,
			},
		}
	}
	resp := &tableData{
		Status: 0,
		Message: "ok",
		Total: total,
		Rows: userResp,
	}	
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *AdminController) GetUser() {
	tpl := ""
	roles, err := models.GetCasbinAllRoles()
	if err != nil {
		roles = make([]models.CasbinRole, 0)
	}
	id, err := c.GetInt64("id")
	if err != nil {
		c.Data["roles"] = roles
		tpl = "admin/user_add"
	} else {
		user, err := models.GetUser2(id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"id"  : id,
				"path": c.Ctx.Request.URL.Path,
			}).Errorf("user id doesn't exist")
			c.Abort("400")
		}
		
		casbinUser, err := models.GetCasbinUser(id)
		roleData := make([]RoleData, len(roles))
		for i := range roles {
			roleData[i].ID = roles[i].ID
			roleData[i].Name = roles[i].Name
			for j := range casbinUser.Roles {
				hadRoleID := casbinUser.Roles[j]
				if uint(hadRoleID) == roles[i].ID {
					roleData[i].Have = true
				}
			}
		}
		c.Data["roles"] = roleData
		c.Data["uid"] = user.Id
		c.Data["username"] = user.Name
		c.Data["age"] = user.Profile2.Age
		c.Data["gender"] = user.Profile2.Gender
		c.Data["email"] = user.Profile2.Email
		c.Data["addr"] = user.Profile2.Address
		tpl = "admin/user_edit"
	}
	
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	c.renderAjaxTemplate(tpl)
}

func (c *AdminController) SaveUser() {
	id, err := c.GetInt64("id")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path":   c.Ctx.Request.URL.Path,
		}).Errorf("can't get id parameter '%s'", c.GetString("id"))
		c.Abort("400")
	}

	age, err := c.GetInt("age")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("age parameter must be integer '%s'", c.GetString("age"))
		c.Abort("400")
	}

	var gender = "male"
	genderType, _ := c.GetInt("gender")
	if genderType == 1 {
		gender = "female"
	}

	user := &models.User2{
		Id: id,
		Name: c.GetString("name"),
		Profile2: models.Profile{
			Gender: gender,
			Age: age,
			Address: c.GetString("addr"),
			Email: c.GetString("email"),
		},
	}

	resp := &responseData{
		Status: 0,
		Message: "ok",
	}
	if err := models.SaveUser2(user); err != nil {
		resp.Status = 100
		resp.Message = "failed"
	}

	roles := make([]uint, 0)
	c.Ctx.Input.Bind(&roles, "role")	
	models.SaveCasbinUser(&models.CasbinUser{ID: user.Id, Name: user.Name}, roles)
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *AdminController) CreateUser() {
	name := c.GetString("name")
	if name == "" {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("name must be provided")
		c.Abort("400")
	}

	password := c.GetString("password")
	if password == "" {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("password must be provided")
		c.Abort("400")
	}

	age, err := c.GetInt("age")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("age must be integer '%s'", c.GetString("age"))
		c.Abort("400")
	}
	
	gender := "male"
	genderType, err := c.GetInt("gender")
	if genderType != 0 {
		gender = "female"
	}
	
	user := models.User2{
		Name: name,
		Password: password,
		Profile2: models.Profile{
			Age: age,
			Gender: gender,
			Email: c.GetString("email"),
			Address: c.GetString("addr"),
		},
	}

	resp := &responseData{
		Status: 0,
		Message: "ok",
	}	
	err = models.CreateUser2(&user)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("create user failed:%v", err)
		resp.Status = 100
		resp.Message = "用户名重复"
	}

	roles := make([]uint, 0)
	c.Ctx.Input.Bind(&roles, "role")
	models.SaveCasbinUser(&models.CasbinUser{ID: user.Id, Name: user.Name}, roles)

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *AdminController) DeleteUser() {
	id, err := c.GetInt64("id")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("can't get id parameter '%s'", c.GetString("id"))
		c.Abort("400")
	}

	resp := &responseData{
		Status: 0,
		Message: "ok",
	}	
	err = models.DeleteUser2(id)		
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("delete user failed:%v", err)
		resp.Status = 101
		resp.Message = "删除用户失败"
	}

	c.Data["json"] = resp
	c.ServeJSON()	
}

func (c *AdminController) RoleList() {
	c.Data["pageTitle"] = "角色列表"
	c.Data["xsrf_token"] = c.XSRFToken()
	c.renderNestedTemplate("admin/roles")
}

func (c *AdminController) GetRole() {
	tpl := ""
	id, err := c.GetUint32("id")
	if err != nil {
		tpl = "admin/role_add"
	} else {
		groups := models.GetCasbinGroupsWithoutEmpty()		
		role, err := models.GetCasbinRole(uint(id))
		if err == nil {
			hadPermissions := make([]uint, 0)
			for i := range groups {
				haveAll := true
				group := groups[i]
				for j := range group.Permissions {
					pid := group.Permissions[j].ID
					if role.HasPermission(pid) {
						hadPermissions = append(hadPermissions, pid)
					} else {
						haveAll = false
					}
				}
				if haveAll {
					hadPermissions = append(hadPermissions, group.ID * 10000)
				}
				groups[i].ID = group.ID * 10000
			}
			c.Data["HadPermissions"] = hadPermissions
		} else {
			logrus.WithFields(logrus.Fields{
				"role": id,
				"path": c.Ctx.Request.URL.Path,
			}).Errorf("load role data failed:%v", err)
			c.Abort("400")
		}
		c.Data["id"] = id
		c.Data["PermissionGroups"] = groups
		tpl = "admin/role_edit"
	}
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	c.renderAjaxTemplate(tpl)
}

func (c *AdminController) GetRoles() {
	page, err := c.GetInt("page")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user": c.userName,
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("invalid page parameter '%s'", c.GetString("page"))
		c.Abort("400")
	}
	limit, err := c.GetInt("limit")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user": c.userName,
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("invalid limit parameter '%s'", c.GetString("limit"))		
		c.Abort("400")
	}

	offset := (page - 1) * limit
	roles, total := models.GetCasbinRoles(offset, limit)
	roleResp := make([]models.CasbinRoleResp, len(roles))
	for i := range roles {
		r := roles[i]
		roleResp[i] = models.CasbinRoleResp{
			ID: r.ID,
			Name: r.Name,
			CreateTime: models.JSONTime(r.CreatedAt),
		}
	}
	resp := &tableData{
		Status: 0,
		Message: "ok",
		Total: total,
		Rows: roleResp,
	}	
	c.Data["json"] = resp
	c.ServeJSON()	
}

func (c *AdminController) SaveRole() {
	roleid, err := c.GetUint32("id")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("role id must be provided '%s'", c.GetString("id"))
		c.Abort("400")
	}

	checkedArray := c.GetString("checked")
	idArray := strings.Split(checkedArray, ",")
	ids := make([]uint, len(idArray))
	for i := range idArray {
		if id, err := strconv.Atoi(idArray[i]); err == nil {
			ids[i] = uint(id)
		}
	}

	resp := &responseData{
		Status: 0,
		Message: "ok",
	}		
	err = models.SaveCasbinRole(uint(roleid), ids)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("save role's permissions failed:%v", err)
		resp.Status = 100
		resp.Message = "保存角色权限失败"
	}
	c.Data["json"] = resp
	c.ServeJSON()	
}

func (c *AdminController) CreateRole() {
	name := c.GetString("name")
	if name == "" {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("role name must be provided")
		c.Abort("400")
	}

	role := models.CasbinRole{
		Name: name,
	}
	resp := &responseData{
		Status: 0,
		Message: "ok",
	}	
	err := models.CreateCasbinRole(&role)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("create role failed:%v", err)
		resp.Status = 100
		resp.Message = "角色名重复"
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *AdminController) DeleteRole() {
	id, err := c.GetUint32("id")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("can't get role id parameter '%s'", c.GetString("id"))
		c.Abort("400")
	}

	resp := &responseData{
		Status: 0,
		Message: "ok",
	}	
	err = models.DeleteCasbinRole(uint(id))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("delete role failed:%v", err)
		resp.Status = 101
		resp.Message = "删除角色失败"
	}

	c.Data["json"] = resp
	c.ServeJSON()	
}

func (c *AdminController) PermissionList() {
	groups := models.GetCasbinGroups()
	c.Data["permissGroup"] = groups
	c.Data["pageTitle"] = "权限列表"
	c.Data["xsrf_token"] = c.XSRFToken()
	c.renderNestedTemplate("admin/permissions")
}

func (c *AdminController) GetPermission() {
	tpl := "admin/permission_add"
	gid, err := c.GetInt64("group")
	if err != nil {
		gid = 0
	}
	c.Data["groupID"] = gid
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	c.renderAjaxTemplate(tpl)
}

func (c *AdminController) CreatePermission() {
	name := c.GetString("name")
	if name == "" {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("name must be provided")
		c.Abort("400")
	}

	resource := c.GetString("resource")
	if resource == "" {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("resource must be provided")
		c.Abort("400")
	}

	actionID, err := c.GetInt("action")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("action must be provided")
		c.Abort("400")
	}

	gid, err := c.GetUint32("group")
	if err != nil || gid == 0 {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("group id must be provided")
		c.Abort("400")
	}

	// convert action id to name
	action := ""
	switch actionID {
	case 0:
		action = "GET"
	case 1:
		action = "POST"
	case 2:
		action = "PUT"
	case 3:
		action = "DELETE"
	case 4:
		action = "*"
	}
	
	permission := models.CasbinPermission{
		Name: name,
		Resource: resource,
		Action: action,
	}	
	resp := &responseData{
		Status: 0,
		Message: "ok",
	}	
	err = models.AppendCasbinPermissionToGroup(uint(gid), &permission)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("create permission failed:%v", err)
		resp.Status = 100
		resp.Message = "权限名重复"
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *AdminController) DeletePermission() {
	id, err := c.GetUint32("id")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("can't get permission id")
		c.Abort("400")
	}

	gid, err := c.GetUint32("group")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("group id must be provided")
		c.Abort("400")
	}	

	resp := &responseData{
		Status: 0,
		Message: "ok",
	}	
	err = models.DeleteCasbinPermissionFromGroup(uint(gid), uint(id))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("delete permission failed:%v", err)
		resp.Status = 101
		resp.Message = "删除权限失败"
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *AdminController) GetGroup() {
	gid, err := c.GetUint32("group")
	if err != nil {
		tpl := "admin/group_add"
		c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
		c.renderAjaxTemplate(tpl)
	} else {
		permissions := models.GetCasbinPermissionsByGroup(uint(gid))
		resp := &tableData{
			Status: 0,
			Message: "ok",
			Total: len(permissions),
			Rows: permissions,
		}
		c.Data["json"] = resp
		c.ServeJSON()
	}
}

func (c *AdminController) CreateGroup() {
	name := c.GetString("name")
	if name == "" {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("name must be provided")
		c.Abort("400")
	}
	
	resp := &responseData{
		Status: 0,
		Message: "ok",
	}	
	err := models.CreateCasbinGroup(&models.CasbinGroup{Name: name})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("create group failed:%v", err)
		resp.Status = 100
		resp.Message = "组名重复"
	}

	c.Data["json"] = resp
	c.ServeJSON()	
}

func (c *AdminController) DeleteGroup() {
	group, err := c.GetUint32("group")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("group id must be provided")
		c.Abort("400")
	}

	err = models.DeleteCasbinGroup(uint(group))
	resp := &responseData{
		Status: 0,
		Message: "ok",
	}
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"path": c.Ctx.Request.URL.Path,
		}).Errorf("deletre group failed:%v", err)
		resp.Status = 100
		resp.Message = "删除权限组失败"
	}

	c.Data["json"] = resp
	c.ServeJSON()		
}