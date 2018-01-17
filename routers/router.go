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
	beego.Router("/admin/users", &controllers.AdminController{}, "GET:UserList")
	beego.Router("/admin/users/list", &controllers.AdminController{}, "GET:GetUsers")
	beego.Router("/admin/user", &controllers.AdminController{}, "GET:GetUser;PUT:SaveUser;POST:CreateUser;DELETE:DeleteUser")
	beego.Router("/admin/roles", &controllers.AdminController{}, "GET:RoleList")
	beego.Router("/admin/roles/list", &controllers.AdminController{}, "GET:GetRoles")
	beego.Router("/admin/role", &controllers.AdminController{}, "GET:GetRole;PUT:SaveRole;POST:CreateRole;DELETE:DeleteRole")
	beego.Router("/admin/permissions", &controllers.AdminController{}, "GET:PermissionList")
	beego.Router("/admin/permission", &controllers.AdminController{}, "GET:GetPermission;POST:CreatePermission;DELETE:DeletePermission")
	beego.Router("/admin/group", &controllers.AdminController{}, "GET:GetGroup;POST:CreateGroup;DELETE:DeleteGroup")
}
