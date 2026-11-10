package aihttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"unicode"

	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

func SafeError(err error, sensitiveValues ...string) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{ErrTarget, ErrResponseTooLarge, ErrResponse} {
		if errors.Is(err, known) {
			return known
		}
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		sensitiveValues = append(sensitiveValues, urlErr.URL)
		err = urlErr.Err
	}
	message := sanitizeErrorMessage(err.Error(), sensitiveValues...)
	if message == "" {
		return ErrRequest
	}
	return errors.New(message)
}

func HTTPError(statusCode int, body []byte, sensitiveValues ...string) error {
	message := fmt.Sprintf("request failed: status %d", statusCode)
	var payload struct {
		Error   json.RawMessage `json:"error"`
		Message string          `json:"message"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return errors.New(message)
	}
	var detail string
	if json.Unmarshal(payload.Error, &detail) != nil {
		var nested struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(payload.Error, &nested) == nil {
			detail = nested.Message
		}
	}
	if detail == "" {
		detail = payload.Message
	}
	if detail = sanitizeErrorMessage(detail, sensitiveValues...); detail != "" {
		message += ": " + detail
	}
	return errors.New(message)
}

func sanitizeErrorMessage(message string, sensitiveValues ...string) string {
	values := append([]string(nil), sensitiveValues...)
	for _, value := range sensitiveValues {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Host == "" {
			continue
		}
		if parsed.User != nil {
			password, _ := parsed.User.Password()
			values = append(values, parsed.User.Username(), password)
		}
		for _, queryValues := range parsed.Query() {
			values = append(values, queryValues...)
		}
	}
	for _, value := range append([]string(nil), values...) {
		encoded, _ := json.Marshal(value)
		values = append(values, url.QueryEscape(value), url.PathEscape(value), string(encoded[1:len(encoded)-1]))
	}
	sort.SliceStable(values, func(i, j int) bool { return len(values[i]) > len(values[j]) })
	replacements := make([]string, 0, len(values)*2)
	for _, value := range values {
		if value != "" {
			replacements = append(replacements, value, "[REDACTED]")
		}
	}
	if len(replacements) > 0 {
		message = strings.NewReplacer(replacements...).Replace(message)
	}
	message = re.AIErrorURLRegex.ReplaceAllString(message, "[REDACTED URL]")
	message = re.AIErrorAuthRegex.ReplaceAllString(message, "$1 [REDACTED]")
	message = re.AIErrorSecretRegex.ReplaceAllString(message, "$1[REDACTED]")
	if strings.ContainsAny(message, "<>") {
		return ""
	}
	message, _, _ = strings.Cut(strings.TrimSpace(message), "\n")
	message, _, _ = strings.Cut(message, "\r")
	message = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, message)
	runes := []rune(strings.TrimSpace(message))
	if len(runes) > 512 {
		return string(runes[:512]) + "…"
	}
	return string(runes)
}
