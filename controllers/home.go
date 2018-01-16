package controllers

// HomeController Operations about home
type HomeController struct {
	baseController
}

func (c *HomeController) Index() {
	menuID, err := c.GetInt("menu")
	if err == nil {
		c.Data["DefaultMenu"] = menuID
	} else {
		c.Data["DefaultMenu"] = 0
	}

	c.Data["pageTitle"] = "系统首页"
	c.renderTemplate("main")
}