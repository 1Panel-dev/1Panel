package webhook_sender

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type responseRejection struct {
	statusCode int
	reason     string
}

func DiagnosticText(value string, input Request) string {
	return captureResponse([]byte(value), input, preparedTarget{})
}

func (e *responseRejection) Error() string {
	return fmt.Sprintf("webhook provider rejected response (HTTP %d)", e.statusCode)
}

func validateResponse(preset Preset, statusCode int, body []byte) error {
	if preset == "" {
		preset = PresetGeneric
	}
	switch preset {
	case PresetGeneric, PresetDiscord, PresetTeams, PresetSlack:
	default:
		return errors.New("unsupported webhook preset")
	}

	reason, failed := businessResponseFailure(body)
	if preset == PresetSlack && reason == "" {
		switch value := strings.TrimSpace(string(body)); value {
		case "invalid_payload", "no_text", "invalid_token", "no_service", "no_service_id",
			"no_active_hooks", "channel_not_found", "channel_is_archived",
			"action_prohibited", "posting_to_general_channel_denied",
			"team_disabled", "user_not_found", "too_many_attachments", "rollup_error":
			reason = value
		}
	}
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return &responseRejection{statusCode: statusCode, reason: reason}
	}
	if preset == PresetSlack {
		if statusCode != http.StatusOK || strings.TrimSpace(string(body)) != "ok" {
			return &responseRejection{statusCode: statusCode, reason: reason}
		}
		return nil
	}
	if failed {
		return &responseRejection{statusCode: statusCode, reason: reason}
	}
	return nil
}

func businessResponseFailure(body []byte) (string, bool) {
	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil || response == nil {
		return "", false
	}
	message := responseMessage(response)
	if message == "" {
		var errorText string
		if json.Unmarshal(response["error"], &errorText) == nil {
			message = strings.TrimSpace(errorText)
		} else {
			var nested map[string]json.RawMessage
			if json.Unmarshal(response["error"], &nested) == nil {
				message = responseMessage(nested)
			}
		}
	}
	for _, key := range []string{"errcode", "errCode", "error_code", "ErrorCode"} {
		if raw, exists := response[key]; exists {
			code, valid := responseCode(raw)
			if !valid || !isSuccessResponseCode(code, false) {
				return responseFailureReason(key, code, message), true
			}
		}
	}
	for _, key := range []string{"success", "ok"} {
		var success bool
		if raw, exists := response[key]; exists && string(raw) != "null" && json.Unmarshal(raw, &success) == nil && !success {
			return responseFailureReason(key, "false", message), true
		}
	}
	if raw, exists := response["error"]; exists {
		var value any
		if json.Unmarshal(raw, &value) == nil {
			switch typed := value.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return responseFailureReason("error", "", typed), true
				}
			case map[string]any:
				if len(typed) != 0 {
					return responseFailureReason("error", "", message), true
				}
			case bool:
				if typed {
					return responseFailureReason("error", "true", message), true
				}
			case float64:
				if typed != 0 {
					return responseFailureReason("error", strconv.FormatFloat(typed, 'g', -1, 64), message), true
				}
			}
		}
	}

	if message != "" {
		for _, key := range []string{"code", "Code", "StatusCode", "statusCode"} {
			if raw, exists := response[key]; exists {
				code, valid := responseCode(raw)
				if !valid || !isSuccessResponseCode(code, true) {
					return responseFailureReason(key, code, message), true
				}
			}
		}
	}
	return message, false
}

func responseMessage(response map[string]json.RawMessage) string {
	for _, key := range []string{"errmsg", "msg", "message", "description", "StatusMessage", "Message"} {
		var message string
		if json.Unmarshal(response[key], &message) == nil && strings.TrimSpace(message) != "" {
			return strings.TrimSpace(message)
		}
	}
	return ""
}

func responseCode(raw json.RawMessage) (string, bool) {
	var code json.Number
	if json.Unmarshal(raw, &code) == nil && code != "" {
		return code.String(), true
	}
	var text string
	if json.Unmarshal(raw, &text) == nil && strings.TrimSpace(text) != "" {
		return strings.TrimSpace(text), true
	}
	return "", false
}

func isSuccessResponseCode(code string, allowHTTP bool) bool {
	number, err := strconv.ParseFloat(code, 64)
	if err == nil {
		return number == 0 || allowHTTP && number >= 200 && number < 300
	}
	return allowHTTP && (strings.EqualFold(code, "ok") || strings.EqualFold(code, "success"))
}

func responseFailureReason(field, code, message string) string {
	reason := field
	if code != "" {
		reason += "=" + code
	}
	if message != "" {
		reason += ": " + message
	}
	return reason
}

func safeResponseError(err error, input Request, target preparedTarget) error {
	var rejection *responseRejection
	if !input.CaptureResponse || !errors.As(err, &rejection) || rejection.reason == "" {
		return err
	}
	reason := captureResponse([]byte(rejection.reason), input, target)
	reason = strings.Join(strings.Fields(reason), " ")
	if reason == "" {
		return err
	}
	return fmt.Errorf("%w: %s", rejection, truncateUTF8(reason, 512))
}
