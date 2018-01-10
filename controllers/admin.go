package controllers

import (

)

// AdminController Operations about home
type AdminController struct {
	baseController
}

func (c *AdminController) ListUsers() {
	c.renderNestedTemplate("admin/users")
}

func (c *AdminController) ListRoles() {
	c.renderNestedTemplate("admin/roles")
}

func (c *AdminController) ListPermissions() {
	c.renderNestedTemplate("admin/persmissions")
}