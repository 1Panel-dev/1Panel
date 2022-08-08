package captcha

import (
	"strings"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

func VerifyCode(codeID string, code string) error {
	if !global.CONF.Captcha.Enable {
		return nil
	}
	if codeID == "" {
		return constant.ErrCaptchaCode
	}
	vv := store.Get(codeID, true)
	vv = strings.TrimSpace(vv)
	code = strings.TrimSpace(code)

	if strings.EqualFold(vv, code) {
		return nil
	}
	return constant.ErrCaptchaCode
}

func CreateCaptcha() (*dto.CaptchaResponse, error) {
	var driverString base64Captcha.DriverString
	driverString.Source = global.CONF.Captcha.Source
	driverString.Width = global.CONF.Captcha.ImgWidth
	driverString.Height = global.CONF.Captcha.ImgHeight
	driverString.NoiseCount = global.CONF.Captcha.NoiseCount
	driverString.Length = global.CONF.Captcha.Length
	driverString.Fonts = []string{"wqy-microhei.ttc"}
	driver := driverString.ConvertFonts()
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, err := c.Generate()
	if err != nil {
		return nil, err
	}
	return &dto.CaptchaResponse{
		CaptchaID: id,
		ImagePath: b64s,
	}, nil
}
