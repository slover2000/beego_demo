package controllers

import (
	"strings"

	"github.com/astaxie/beego"
	"github.com/slover2000/beego_demo/models"
)

// LoginController login controller
type LoginController struct {
	baseController
}

//Login TODO:XSRF过滤
func (c *LoginController) Login() {
	if c.userID > 0 {
		c.Redirect(beego.URLFor("HomeController.Index"), 302)
	}
	
	errorMsg := ""
	username := strings.TrimSpace(c.GetString("username"))
	password := strings.TrimSpace(c.GetString("password"))
	if username != "" && password != "" {
		user, err := models.GetAndVerifyUser(username, password)
		if err != nil {
			errorMsg = "帐号或密码错误"
		} else {
			sess, err := globalSessions.SessionStart(c.Ctx.ResponseWriter.ResponseWriter, c.Ctx.Request)
			if err == nil {
				defer sess.SessionRelease(c.Ctx.ResponseWriter.ResponseWriter)
				sess.Set("uid", user.Id)
				sess.Set("name", user.Username)				
			}
		}

		if errorMsg == "" {
			c.Redirect(beego.URLFor("HomeController.Index"), 302)
		} else {
			flash := beego.NewFlash()
			flash.Error(errorMsg)
			flash.Store(&c.Controller)
			c.Redirect(beego.URLFor("HomeController.login"), 302)
		}
	} else {
		c.Redirect(beego.URLFor("HomeController.login"), 302)
	}
}

// Logout user log out from system
func (c *LoginController) Logout() {
	globalSessions.SessionDestroy(c.Ctx.ResponseWriter.ResponseWriter, c.Ctx.Request)
	c.Redirect(beego.URLFor("HomeController.Login"), 302)
}

func (c *LoginController) NoAuth() {
	c.Ctx.WriteString("没有权限")
}


func (c *LoginController) Register() {

}