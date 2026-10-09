package re

import (
	"fmt"
	"regexp"
)

const (
	OrderByValidationPattern = `^[a-zA-Z_][a-zA-Z0-9_]*$`
)

var regexMap = make(map[string]*regexp.Regexp)

func Init() {
	patterns := []string{
		OrderByValidationPattern,
	}

	for _, pattern := range patterns {
		regexMap[pattern] = regexp.MustCompile(pattern)
	}
}

func GetRegex(pattern string) *regexp.Regexp {
	regex, exists := regexMap[pattern]
	if !exists {
		panic(fmt.Sprintf("regex pattern not found: %s", pattern))
	}
	return regex
}

func RegisterRegex(pattern string) {
	regexMap[pattern] = regexp.MustCompile(pattern)
}
