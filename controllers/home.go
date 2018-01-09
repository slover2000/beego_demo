package controllers

// HomeController Operations about home
type HomeController struct {
	baseController
}

func (c *HomeController) Index() {
	c.Data["pageTitle"] = "系统首页"
	c.renderTemplate("main")
}