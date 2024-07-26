package mfa

import (
	"bytes"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"
)

const secretLength = 16

type Otp struct {
	Secret  string `json:"secret"`
	QrImage string `json:"qrImage"`
}

func GetOtp(username, title string, interval int) (otp Otp, err error) {
	secret := gotp.RandomSecret(secretLength)
	otp.Secret = secret
	totp := gotp.NewTOTP(secret, 6, interval, nil)
	uri := totp.ProvisioningUri(username, title)
	subImg, err := qrcode.Encode(uri, qrcode.Medium, 256)
	dist := make([]byte, 3000)
	base64.StdEncoding.Encode(dist, subImg)
	index := bytes.IndexByte(dist, 0)
	baseImage := dist[0:index]
	otp.QrImage = "data:image/png;base64," + string(baseImage)
	return
}

func ValidCode(code, intervalStr, secret string) bool {
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		global.LOG.Errorf("type conversion failed, err: %v", err)
		return false
	}
	totp := gotp.NewTOTP(secret, 6, interval, nil)
	now := time.Now().Unix()
	prevTime := now - int64(interval)
	return totp.Verify(code, now) || totp.Verify(code, prevTime)
}
