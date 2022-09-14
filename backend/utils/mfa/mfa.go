package mfa

import (
	"bytes"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"
)

const secretLength = 16

type Otp struct {
	Secret  string `json:"secret"`
	QrImage string `json:"qrImage"`
}

func GetOtp(username string) (otp Otp, err error) {
	secret := gotp.RandomSecret(secretLength)
	otp.Secret = secret
	totp := gotp.NewDefaultTOTP(secret)
	uri := totp.ProvisioningUri(username, "1Panel")
	subImg, err := qrcode.Encode(uri, qrcode.Medium, 256)
	dist := make([]byte, 3000)
	base64.StdEncoding.Encode(dist, subImg)
	index := bytes.IndexByte(dist, 0)
	baseImage := dist[0:index]
	otp.QrImage = "data:image/png;base64," + string(baseImage)
	return
}

func ValidCode(code string, secret string) bool {
	totp := gotp.NewDefaultTOTP(secret)
	now := time.Now().Unix()
	strInt64 := strconv.FormatInt(now, 10)
	id16, _ := strconv.Atoi(strInt64)
	return totp.Verify(code, int64(id16))
}
