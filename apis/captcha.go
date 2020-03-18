package apis

import (
	"strings"

	"github.com/mojocn/base64Captcha"
	"github.com/pborman/uuid"
)

type Captcha struct {
	Base
}

// Get /api/captcha
func (c *Captcha) Get() interface{} {
	var config = base64Captcha.ConfigCharacter{
		Height:             60,
		Width:              240,
		Mode:               base64Captcha.CaptchaModeNumberAlphabet,
		IsUseSimpleFont:    true,
		ComplexOfNoiseText: 2,
		ComplexOfNoiseDot:  20,
		IsShowHollowLine:   true,
		IsShowNoiseDot:     true,
		IsShowNoiseText:    false,
		IsShowSlimeLine:    false,
		IsShowSineLine:     false,
		CaptchaLen:         6,
	}

	captchaID, captcaInterfaceInstance := base64Captcha.GenerateCaptcha(strings.Replace(uuid.NewUUID().String(), "-", "", -1), config)
	base64blob := base64Captcha.CaptchaWriteToBase64Encoding(captcaInterfaceInstance)

	return map[string]interface{}{
		"data":       base64blob,
		"captcha_id": captchaID,
		"status":     1,
	}
}
