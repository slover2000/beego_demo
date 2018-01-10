// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"github.com/slover2000/beego_demo/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/object",
			beego.NSInclude(
				&controllers.ObjectController{},
			),
		),
		beego.NSNamespace("/user",
			beego.NSInclude(
				&controllers.UserController{},
			),
		),
		beego.NSNamespace("/api",
			beego.NSInclude(
				&controllers.CaptchaController{},
			),
		),
	)
	beego.AddNamespace(ns)

	beego.Router("/", &controllers.LoginController{}, "*:ShowPage")
	beego.Router("/home", &controllers.HomeController{}, "*:Index")
	beego.Router("/login", &controllers.LoginController{}, "*:Login")
	beego.Router("/logout", &controllers.LoginController{}, "*:Logout")	
	beego.Router("/admin/users", &controllers.AdminController{}, "GET:ListUsers")
	beego.Router("/admin/roles", &controllers.AdminController{}, "GET:ListRoles")
	beego.Router("/admin/permissions", &controllers.AdminController{}, "GET:ListPermissions")
}
