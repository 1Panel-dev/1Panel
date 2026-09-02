package dto

const AlertCustomWebhookSchemaVersion = 1

type AlertConfigStatusUpdate struct {
	ID     uint   `json:"id" validate:"required"`
	Status string `json:"status" validate:"required,oneof=Enable Disable"`
}

type AlertCustomWebhookSecretMutation struct {
	Action string `json:"action,omitempty"`
	Value  string `json:"value,omitempty"`
}

type AlertCustomWebhookURL struct {
	AlertCustomWebhookSecretMutation
	Configured bool   `json:"configured"`
	Masked     string `json:"masked,omitempty"`
}

type AlertCustomWebhookBody struct {
	Type     string                        `json:"type"`
	Template string                        `json:"template,omitempty"`
	Fields   []AlertCustomWebhookFormField `json:"fields,omitempty"`
}

type AlertCustomWebhookFormField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AlertCustomWebhookHeader struct {
	UID        string `json:"uid"`
	Key        string `json:"key"`
	Secret     bool   `json:"secret"`
	Action     string `json:"action,omitempty"`
	Value      string `json:"value,omitempty"`
	Configured bool   `json:"configured,omitempty"`
	Masked     string `json:"masked,omitempty"`
}

type AlertCustomWebhookConfig struct {
	SchemaVersion int                        `json:"schemaVersion"`
	State         string                     `json:"state,omitempty"`
	DisplayName   string                     `json:"displayName"`
	Preset        string                     `json:"preset"`
	Method        string                     `json:"method"`
	URL           AlertCustomWebhookURL      `json:"url"`
	Body          AlertCustomWebhookBody     `json:"body"`
	Headers       []AlertCustomWebhookHeader `json:"headers"`
}

type AlertCustomWebhookSecretConfig struct {
	SchemaVersion int               `json:"schemaVersion"`
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers,omitempty"`
}

type AlertCustomWebhookResolvedConfig struct {
	SchemaVersion int
	DisplayName   string
	Preset        string
	Method        string
	URL           string
	Body          AlertCustomWebhookBody
	Headers       []AlertCustomWebhookResolvedHeader
}

type AlertCustomWebhookResolvedHeader struct {
	Key   string
	Value string
}

type AlertConfigTestResult struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"statusCode,omitempty"`
	Duration   int64  `json:"duration,omitempty"` // milliseconds
	Message    string `json:"message,omitempty"`
	Response   string `json:"response,omitempty"`
}
