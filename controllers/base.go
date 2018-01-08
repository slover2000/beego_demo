package controllers

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/session"
	"github.com/casbin/beego-orm-adapter"
	"github.com/casbin/casbin"
	"github.com/gogap/logrus"
)

const (
	STATUS_OK              = 0
	STATUS_PERMISSION_DENY = -1

	ajaxTagParameter = "x-ajax-key"
)

type responseData struct {
	status  int         `json:"status"`
	message string      `json:"msg"`
	data    interface{} `json:"data"`
}

type baseController struct {
	beego.Controller
	permission bool
	userID     uint64
	userName   string
}

var enforcer *casbin.SyncedEnforcer
var globalSessions *session.Manager

func initCasbinPolicy() {
	// Initialize a Beego ORM adapter and use it in a Casbin enforcer:
	// The adapter will use the Postgres database named "casbin".
	// If it doesn't exist, the adapter will create it automatically.
	dataSource := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%d sslmode=disable",
		beego.AppConfig.String("postgres.database"),
		beego.AppConfig.String("postgres.user"),
		beego.AppConfig.String("postgres.password"),
		beego.AppConfig.String("postgres.host"),
		beego.AppConfig.DefaultInt("postgres.port", 5432))
	a := beegoormadapter.NewAdapter("postgres", dataSource, true) // Your driver and data source.
	enforcer = casbin.NewSyncedEnforcer("./conf/rbac_model.conf", a)
	// Load the policy from DB.
	enforcer.LoadPolicy()
}

func initSessionManager() {
	sessionConfig := &session.ManagerConfig{
		CookieName:      "sessionid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     3600,
		Secure:          false,
		CookieLifeTime:  3600,
		ProviderConfig:  "./tmp",
	}
	globalSessions, _ = session.NewManager("memory", sessionConfig)
	go globalSessions.GC()
}

func init() {
	initSessionManager()
	initCasbinPolicy()
}

func (c *baseController) Prepare() {
	c.Data["version"] = beego.AppConfig.String("site.version")
	c.Data["siteName"] = beego.AppConfig.String("site.name")
	if c.authenticate() {
		c.permission = true	
		c.Data["userName"] = c.userName
	}
}

func (c *baseController) authenticate() bool {
	controllerName, actionName := c.GetControllerAndAction()
	if controllerName == "HomeController" && actionName == "Login" {
		return false
	}

	req := c.Ctx.Request
	resp := c.Ctx.ResponseWriter.ResponseWriter
	sess, err := globalSessions.SessionStart(resp, req)
	if err != nil {
		c.Redirect(beego.URLFor("HomeController.Login"), 302)
		return false
	}
	defer sess.SessionRelease(resp)

	id := sess.Get("uid")
	name := sess.Get("name")
	if id != nil {
		c.userID = id.(uint64)
	}
	if name != nil {
		c.userName = name.(string)
	}

	// check permission
	if !enforcer.Enforce(c.userName, req.URL.Path, req.Method) {
		logrus.WithFields(logrus.Fields{
			"user":   c.userName,
			"path":   req.URL.Path,
			"method": req.Method,
		}).Warn("permission deny")

		isAjax := req.Header.Get(ajaxTagParameter)
		if isAjax != "" {
			c.ajaxFailure(STATUS_PERMISSION_DENY, "没有权限")
		} else {
			c.Redirect(beego.URLFor("HomeController.Login"), 302)
		}
		return false
	}
	return true
}

func (c *baseController) renderTemplate(tpl string) {
	var tplname string
	if tpl != "" {
		tplname = strings.Join([]string{tpl, "html"}, ".")
	} else {
		controllerName, actionName := c.GetControllerAndAction()
		tplname = fmt.Sprintf("%s/%s.html", controllerName, actionName)
	}
	c.Layout = "layout.html"
	c.TplName = tplname
}

func (c *baseController) ajaxSuccess(data interface{}) {
	c.Data["json"] = responseData{status: STATUS_OK, data: data}
	c.ServeJSON()
}

func (c *baseController) ajaxFailure(errno int, errmsg string) {
	c.Data["json"] = responseData{status: errno, message: errmsg}
	c.ServeJSON()
}

func (c *baseController) getClientIP() string {
	s := strings.Split(c.Ctx.Request.RemoteAddr, ":")
	return s[0]
}
