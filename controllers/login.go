package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
)

// LoginController login controller
type LoginController struct {
	baseController
}

//LoginIn TODO:XSRF过滤
func (c *LoginController) LoginIn() {
	if c.userID > 0 {
		c.Redirect("main", 302)
	}

	username := strings.TrimSpace(c.GetString("username"))
	password := strings.TrimSpace(c.GetString("password"))
	if username != "" && password != "" {
		user, err := models.AdminGetByName(username)
		fmt.Println(user)
		flash := beego.NewFlash()
		errorMsg := ""
		if err != nil || user.Password != libs.Md5([]byte(password+user.Salt)) {
			errorMsg = "帐号或密码错误"
		} else if user.Status == -1 {
			errorMsg = "该帐号已禁用"
		} else {
			user.LastLogin = time.Now().Unix()
			user.Update()
			authkey := libs.Md5([]byte(self.getClientIP() + "|" + user.Password + user.Salt))
			self.Ctx.SetCookie("auth", strconv.Itoa(user.Id)+"|"+authkey, 7*86400)

			self.redirect(beego.URLFor("HomeController.Index"))
		}
		flash.Error(errorMsg)
		flash.Store(&self.Controller)
		self.redirect(beego.URLFor("LoginController.LoginIn"))
	}
}

func (c *LoginController) LoginOut() {
	globalSessions.SessionDestroy(c.Ctx.ResponseWriter.ResponseWriter, c.Ctx.Request)
	c.Redirect("login", 302)
}

func (c *LoginController) NoAuth() {
	c.Ctx.WriteString("没有权限")
}
