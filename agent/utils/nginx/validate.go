package nginx

import (
	"errors"
	"net/url"
	"strings"
	"unicode"

	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

type Mode int

const (
	ModeGeneric Mode = iota
	ModeHost
	ModeURL
	ModePath
	ModeAllowMethods
)

func NginxSafeString(input string, mode Mode) (string, error) {
	if input == "" {
		return "", errors.New("empty value not allowed")
	}

	for _, r := range input {
		if unicode.IsControl(r) {
			return "", errors.New("control characters not allowed")
		}
	}

	if strings.ContainsAny(input, ";\n\r{}#`$") {
		return "", errors.New("illegal nginx syntax characters")
	}

	switch mode {
	case ModeHost:
		return validateHost(input)
	case ModeURL:
		return validateURL(input)
	case ModePath:
		return validatePath(input)
	case ModeAllowMethods:
		return validateAllowMethods(input)
	default:
		return input, nil
	}
}

func validateHost(host string) (string, error) {
	if !re.GetRegex(re.NginxHostPattern).MatchString(host) {
		return "", errors.New("invalid host format")
	}
	return host, nil
}

func validateURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("invalid url")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("unsupported scheme")
	}

	if u.Host == "" {
		return "", errors.New("missing host")
	}

	return raw, nil
}

func validatePath(p string) (string, error) {
	if !re.GetRegex(re.NginxPathPattern).MatchString(p) {
		return "", errors.New("invalid path")
	}
	return p, nil
}

var allowedMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"PATCH":   true,
	"OPTIONS": true,
	"HEAD":    true,
}

func validateAllowMethods(input string) (string, error) {
	parts := strings.Split(input, ",")
	for i, method := range parts {
		method = strings.TrimSpace(strings.ToUpper(method))
		if !allowedMethods[method] {
			return "", errors.New("invalid HTTP method: " + method)
		}
		parts[i] = method
	}
	return strings.Join(parts, ", "), nil
}
