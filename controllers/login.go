package controllers

import (
	"strings"
	"html/template"

	"github.com/astaxie/beego"
	"github.com/slover2000/beego_demo/models"
)

// LoginController login controller
type LoginController struct {
	beego.Controller
}

func (c *LoginController) Prepare() {
	c.Data["version"] = beego.AppConfig.String("site.version")
	c.Data["siteName"] = beego.AppConfig.String("site.name")
}

//Login TODO:XSRF过滤
func (c *LoginController) Login() {	
	errorMsg := ""
	username := template.HTMLEscapeString(strings.TrimSpace(c.GetString("username")))
	password := template.HTMLEscapeString(strings.TrimSpace(c.GetString("password")))
	if username != "" && password != "" {
		user, err := models.GetAndVerifyUser(username, password)
		if err != nil {
			errorMsg = "帐号或密码错误"
		} else {
			sess, err := globalSessions.SessionStart(c.Ctx.ResponseWriter.ResponseWriter, c.Ctx.Request)
			if err == nil {
				defer sess.SessionRelease(c.Ctx.ResponseWriter.ResponseWriter)
				sess.Set("uid", user.Id)
				sess.Set("name", user.Name)				
			}
		}

		if errorMsg == "" {
			c.Redirect(beego.URLFor("HomeController.Index"), 302)
		} else {
			flash := beego.NewFlash()
			flash.Error(errorMsg)
			flash.Notice(username)
			flash.Store(&c.Controller)
			c.Redirect(beego.URLFor("LoginController.ShowPage"), 302)
		}
	} else {
		flash := beego.NewFlash()
		flash.Error("please input name and password")		
		flash.Store(&c.Controller)		
		c.Redirect(beego.URLFor("LoginController.ShowPage"), 302)
	}
}

// Logout user log out from system
func (c *LoginController) Logout() {
	globalSessions.SessionDestroy(c.Ctx.ResponseWriter.ResponseWriter, c.Ctx.Request)	
	c.Redirect(beego.URLFor("LoginController.ShowPage"), 302)
}

func (c *LoginController) ShowPage() {
	beego.ReadFromRequest(&c.Controller)
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	c.TplName = "login.html"
}

func (c *LoginController) Register() {

}