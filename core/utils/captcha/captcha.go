package captcha

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
)

const (
	captchaWidth      = 160
	captchaHeight     = 48
	captchaExpiration = 10 * time.Minute
	captchaStoreLimit = 10000
)

var store = newMemoryStore(captchaStoreLimit, captchaExpiration)

func VerifyCode(codeID string, code string) string {
	answer := store.consume(codeID)
	code = strings.TrimSpace(code)
	if codeID == "" || code == "" || answer == "" || answer != code {
		return "ErrCaptchaCode"
	}
	return ""
}

func CreateCaptcha() (*dto.CaptchaResponse, error) {
	question, answer, err := generateQuestion()
	if err != nil {
		return nil, err
	}
	imagePath, err := renderQuestion(question)
	if err != nil {
		return nil, err
	}
	id := rand.Text()
	if err := store.put(id, answer); err != nil {
		return nil, err
	}
	return &dto.CaptchaResponse{CaptchaID: id, ImagePath: imagePath}, nil
}

func generateQuestion() (question, answer string, err error) {
	limits := [...]int64{3, 80, 20, 80}
	var values [4]int
	for i, limit := range limits {
		n, err := rand.Int(rand.Reader, big.NewInt(limit))
		if err != nil {
			return "", "", fmt.Errorf("generate captcha: %w", err)
		}
		values[i] = int(n.Int64())
	}
	switch values[0] {
	case 0:
		a, b := values[1]%20, values[2]
		return fmt.Sprintf("%d+%d=?", a, b), fmt.Sprint(a + b), nil
	case 1:
		a, b := values[1]+values[2], values[3]
		return fmt.Sprintf("%d-%d=?", a, b), fmt.Sprint(a - b), nil
	default:
		a, b := values[1]%10, values[2]%10
		return fmt.Sprintf("%dx%d=?", a, b), fmt.Sprint(a * b), nil
	}
}
