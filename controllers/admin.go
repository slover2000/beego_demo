package controllers

import (
	"github.com/sirupsen/logrus"

	"github.com/slover2000/beego_demo/models"
)

// AdminController Operations about home
type AdminController struct {
	baseController
}

func (c *AdminController) UserList() {
	c.Data["pageTitle"] = "用户列表"
	c.renderNestedTemplate("admin/users")
}

func (c *AdminController) GetUsers() {
	page, err := c.GetInt("page")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user":   c.userName,
			"path":   c.Ctx.Request.URL.Path,
		}).Errorf("invalid page parameter '%s'", c.GetString("page"))
		c.Abort("400")
	}
	limit, err := c.GetInt("limit")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"user":   c.userName,
			"path":   c.Ctx.Request.URL.Path,
		}).Errorf("invalid limit parameter '%s'", c.GetString("limit"))		
		c.Abort("400")
	}

	offset := (page - 1) * limit
	users, total := models.GetUsers(offset, limit)
	resp := &tableData{
		Status: 0,
		Message: "ok",
		Total: total,
		Rows: users,
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *AdminController) RoleList() {
	c.Data["pageTitle"] = "角色列表"
	c.renderNestedTemplate("admin/roles")
}

func (c *AdminController) PermissionList() {
	c.Data["pageTitle"] = "权限列表"
	c.renderNestedTemplate("admin/permissions")
}