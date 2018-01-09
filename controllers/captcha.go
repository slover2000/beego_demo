package controllers

import (
	"encoding/json"
	
	"github.com/mojocn/base64Captcha"
	"github.com/astaxie/beego"
)

// CaptchaController ...
type CaptchaController struct {
	beego.Controller
}

var captchaConfig *base64Captcha.ConfigCharacter

func init() {
	captchaConfig = &base64Captcha.ConfigCharacter{
		Height:             60,
		Width:              240,
		//const CaptchaModeNumber:数字,CaptchaModeAlphabet:字母,CaptchaModeArithmetic:算术,CaptchaModeNumberAlphabet:数字字母混合.
		Mode:               base64Captcha.CaptchaModeNumber,
		ComplexOfNoiseText: base64Captcha.CaptchaComplexLower,
		ComplexOfNoiseDot:  base64Captcha.CaptchaComplexLower,
		IsShowHollowLine:   false,
		IsShowNoiseDot:     false,
		IsShowNoiseText:    false,
		IsShowSlimeLine:    false,
		IsShowSineLine:     false,
		CaptchaLen:         6,		
	}
}

// @Title Generate base64 encoding image data
// @Description get base64 encoding captcha
// @Success 200 base64 encoding captcha image string
// @Failure 404 can't generate captcha image string
// @router /captcha [get]
func (u *CaptchaController) GenerateCaptcha() {	
	//GenerateCaptcha 第一个参数为空字符串,包会自动在服务器一个随机种子给你产生随机uiid.
	captchaId, captchaData := base64Captcha.GenerateCaptcha("", *captchaConfig)
	base64Png := base64Captcha.CaptchaWriteToBase64Encoding(captchaData)

	resp := u.Ctx.ResponseWriter.ResponseWriter
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	body := map[string]interface{}{"code": 0, "data": base64Png, "captchaId": captchaId, "msg": "success"}
	json.NewEncoder(resp).Encode(body)
}

func VerifyCaptcha(captchaId, captchaValue string) bool {
	return base64Captcha.VerifyCaptcha(captchaId, captchaValue)
}