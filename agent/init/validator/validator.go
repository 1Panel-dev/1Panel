package validator

import (
	"unicode"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/re"

	"github.com/go-playground/validator/v10"
)

func Init() {
	validator := validator.New()
	if err := validator.RegisterValidation("name", checkNamePattern); err != nil {
		panic(err)
	}
	if err := validator.RegisterValidation("ip", checkIpPattern); err != nil {
		panic(err)
	}
	if err := validator.RegisterValidation("password", checkPasswordPattern); err != nil {
		panic(err)
	}
	global.VALID = validator
}

func checkNamePattern(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return re.GetRegex(re.ValidatorNamePattern).MatchString(value)
}

func checkIpPattern(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return re.GetRegex(re.ValidatorIPPattern).MatchString(value)
}

func checkPasswordPattern(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if len(value) < 8 || len(value) > 30 {
		return false
	}

	hasNum := false
	hasLetter := false
	for _, r := range value {
		if unicode.IsLetter(r) && !hasLetter {
			hasLetter = true
		}
		if unicode.IsNumber(r) && !hasNum {
			hasNum = true
		}
		if hasLetter && hasNum {
			return true
		}
	}

	return false
}
