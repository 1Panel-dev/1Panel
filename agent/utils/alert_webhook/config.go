package alert_webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/google/uuid"
)

const (
	MaskedSecret               = "******"
	DefaultGenericJSONTemplate = `{"schema_version":"1","title":"{{title}}","message":"{{message}}","type":"{{type}}","node_name":"{{nodeName}}","timestamp":"{{timestamp}}"}`

	coreKeyPrefix  = "core:v1:"
	agentKeyPrefix = "agent:v1:"

	maxDisplayNameRunes = 64
	maxURLLength        = 8192
	maxHeaders          = 64
	maxHeaderNameLength = 256
	maxHeaderValueLen   = 16 * 1024
	maxBodyLength       = 256 * 1024
	maxFormFields       = 128
	maxConfigLength     = 512 * 1024
	maxSecretLength     = 256 * 1024
)

var supportedPresets = map[string]struct{}{
	"genericJson":    {},
	"slack":          {},
	"discord":        {},
	"teamsWorkflows": {},
	"custom":         {},
}

var forbiddenHeaders = map[string]struct{}{
	"connection":          {},
	"content-length":      {},
	"content-type":        {},
	"host":                {},
	"proxy-connection":    {},
	"proxy-authorization": {},
	"te":                  {},
	"trailer":             {},
	"transfer-encoding":   {},
	"upgrade":             {},
}

var supportedTemplateVariables = map[string]struct{}{
	"{{title}}":     {},
	"{{message}}":   {},
	"{{type}}":      {},
	"{{nodeName}}":  {},
	"{{timestamp}}": {},
}

type PreparedConfig struct {
	Config       string
	SecretConfig string
	DisplayName  string
}

func Prepare(rawConfig, status string, existing *model.AlertConfig) (PreparedConfig, error) {
	if len(rawConfig) > maxConfigLength {
		return PreparedConfig{}, fmt.Errorf("custom webhook config exceeds %d bytes", maxConfigLength)
	}
	var mutation dto.AlertCustomWebhookConfig
	if err := decodeStrict(rawConfig, &mutation); err != nil {
		return PreparedConfig{}, fmt.Errorf("decode custom webhook config: %w", err)
	}
	if mutation.SchemaVersion != dto.AlertCustomWebhookSchemaVersion {
		return PreparedConfig{}, fmt.Errorf("unsupported custom webhook schemaVersion: %d", mutation.SchemaVersion)
	}
	if mutation.State != "" {
		return PreparedConfig{}, fmt.Errorf("custom webhook state is read-only")
	}

	existingSecret := dto.AlertCustomWebhookSecretConfig{
		SchemaVersion: dto.AlertCustomWebhookSchemaVersion,
		Headers:       map[string]string{},
	}
	if existing != nil {
		if existing.Type != constant.Custom {
			return PreparedConfig{}, fmt.Errorf("alert config %d is not a custom webhook", existing.ID)
		}
		if mutationNeedsExistingSecret(mutation) {
			var err error
			existingSecret, err = decryptSecret(existing.SecretConfig)
			if err != nil {
				return PreparedConfig{}, err
			}
		}
	}

	stored, secret, err := mergeAndValidate(mutation, status, existingSecret, existing != nil)
	if err != nil {
		return PreparedConfig{}, err
	}
	configData, err := json.Marshal(stored)
	if err != nil {
		return PreparedConfig{}, fmt.Errorf("encode custom webhook config: %w", err)
	}
	secretData, err := json.Marshal(secret)
	if err != nil {
		return PreparedConfig{}, fmt.Errorf("encode custom webhook secret: %w", err)
	}
	if len(configData) > maxConfigLength {
		return PreparedConfig{}, fmt.Errorf("custom webhook config exceeds %d bytes", maxConfigLength)
	}
	if len(secretData) > maxSecretLength {
		return PreparedConfig{}, fmt.Errorf("custom webhook secret exceeds %d bytes", maxSecretLength)
	}
	secretCipher, err := encryptSecretPlain(string(secretData))
	if err != nil {
		return PreparedConfig{}, fmt.Errorf("encrypt custom webhook secret: %w", err)
	}

	return PreparedConfig{
		Config:       string(configData),
		SecretConfig: secretCipher,
		DisplayName:  stored.DisplayName,
	}, nil
}

func mutationNeedsExistingSecret(mutation dto.AlertCustomWebhookConfig) bool {
	if mutation.URL.Action == "keep" {
		return true
	}
	for _, header := range mutation.Headers {
		if header.Secret && header.Action == "keep" {
			return true
		}
	}
	return false
}

type plainConfigView struct {
	SchemaVersion int                        `json:"schemaVersion"`
	State         string                     `json:"state,omitempty"`
	DisplayName   string                     `json:"displayName"`
	Preset        string                     `json:"preset"`
	Method        string                     `json:"method"`
	URL           string                     `json:"url"`
	Body          dto.AlertCustomWebhookBody `json:"body"`
	Headers       []plainConfigViewHeader    `json:"headers"`
}

type plainConfigViewHeader struct {
	UID    string `json:"uid"`
	Key    string `json:"key"`
	Secret bool   `json:"secret"`
	Value  string `json:"value"`
}

// PlainView decrypts a custom webhook only for the authenticated alert-config
// read contract. The encrypted SecretConfig remains the sole persisted copy.
func PlainView(config model.AlertConfig) (string, error) {
	if config.Type != constant.Custom {
		return config.Config, nil
	}
	stored, err := decodeStored(config.Config)
	if err != nil {
		if _, _, legacy := legacyValues(config.Config); legacy {
			return config.Config, nil
		}
		return safeFallbackView(config, "", "invalid")
	}
	secret, err := validatedStoredSecret(config, stored)
	if err != nil {
		return safeFallbackView(config, stored.DisplayName, "invalid")
	}
	view := plainConfigView{
		SchemaVersion: stored.SchemaVersion,
		DisplayName:   stored.DisplayName,
		Preset:        stored.Preset,
		Method:        stored.Method,
		URL:           secret.URL,
		Body:          stored.Body,
		Headers:       make([]plainConfigViewHeader, 0, len(stored.Headers)),
	}
	for _, header := range stored.Headers {
		value := header.Value
		if header.Secret {
			value = secret.Headers[header.UID]
		}
		view.Headers = append(view.Headers, plainConfigViewHeader{
			UID:    header.UID,
			Key:    header.Key,
			Secret: header.Secret,
			Value:  value,
		})
	}
	result, err := json.Marshal(view)
	if err != nil {
		return "", fmt.Errorf("encode custom webhook plain view: %w", err)
	}
	return string(result), nil
}

func NormalizeLegacy(rawConfig, status, fallbackName string) (PreparedConfig, string, bool, error) {
	displayName, rawURL, legacy := legacyValues(rawConfig)
	if !legacy {
		return PreparedConfig{}, status, false, nil
	}
	if displayName == "" {
		displayName = strings.TrimSpace(fallbackName)
	}
	if displayName == "" || utf8.RuneCountInString(displayName) > maxDisplayNameRunes {
		displayName = "Custom Webhook"
	}
	if status != constant.AlertEnable && status != constant.AlertDisable {
		status = constant.AlertDisable
	}
	urlMutation := dto.AlertCustomWebhookSecretMutation{Action: "clear"}
	if rawURL != "" {
		if err := validateURL(rawURL); err == nil {
			urlMutation = dto.AlertCustomWebhookSecretMutation{Action: "replace", Value: rawURL}
		} else {
			status = constant.AlertDisable
		}
	} else {
		status = constant.AlertDisable
	}
	mutation := dto.AlertCustomWebhookConfig{
		SchemaVersion: dto.AlertCustomWebhookSchemaVersion,
		DisplayName:   displayName,
		Preset:        "genericJson",
		Method:        http.MethodPost,
		URL:           dto.AlertCustomWebhookURL{AlertCustomWebhookSecretMutation: urlMutation},
		Body: dto.AlertCustomWebhookBody{
			Type:     "json",
			Template: DefaultGenericJSONTemplate,
		},
		Headers: make([]dto.AlertCustomWebhookHeader, 0),
	}
	data, err := json.Marshal(mutation)
	if err != nil {
		return PreparedConfig{}, status, true, fmt.Errorf("encode legacy custom webhook config: %w", err)
	}
	prepared, err := Prepare(string(data), status, nil)
	return prepared, status, true, err
}

func safeFallbackView(config model.AlertConfig, displayName, state string) (string, error) {
	if displayName == "" {
		displayName = strings.TrimSpace(config.Title)
	}
	if displayName == "" || utf8.RuneCountInString(displayName) > maxDisplayNameRunes {
		displayName = "Custom Webhook"
	}
	view := dto.AlertCustomWebhookConfig{
		SchemaVersion: dto.AlertCustomWebhookSchemaVersion,
		State:         state,
		DisplayName:   displayName,
		Preset:        "genericJson",
		Method:        http.MethodPost,
		URL:           dto.AlertCustomWebhookURL{Configured: false},
		Body: dto.AlertCustomWebhookBody{
			Type:     "json",
			Template: DefaultGenericJSONTemplate,
		},
		Headers: make([]dto.AlertCustomWebhookHeader, 0),
	}
	result, err := json.Marshal(view)
	if err != nil {
		return "", fmt.Errorf("encode custom webhook fallback view: %w", err)
	}
	return string(result), nil
}

func legacyValues(rawConfig string) (string, string, bool) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(rawConfig), &raw); err != nil || raw == nil || len(raw) != 2 {
		return "", "", false
	}
	displayRaw, hasDisplayName := raw["displayName"]
	urlRaw, hasURL := raw["url"]
	if !hasDisplayName || !hasURL {
		return "", "", false
	}
	var displayName, rawURL string
	if err := json.Unmarshal(displayRaw, &displayName); err != nil {
		return "", "", false
	}
	if err := json.Unmarshal(urlRaw, &rawURL); err != nil {
		return "", "", false
	}
	return strings.TrimSpace(displayName), strings.TrimSpace(rawURL), true
}

func Resolve(config model.AlertConfig) (dto.AlertCustomWebhookResolvedConfig, error) {
	if config.Type != constant.Custom {
		return dto.AlertCustomWebhookResolvedConfig{}, fmt.Errorf("alert config %d is not a custom webhook", config.ID)
	}
	stored, err := decodeStored(config.Config)
	if err != nil {
		return dto.AlertCustomWebhookResolvedConfig{}, err
	}
	secret, err := decryptSecret(config.SecretConfig)
	if err != nil {
		return dto.AlertCustomWebhookResolvedConfig{}, err
	}
	if err := validateURL(secret.URL); err != nil {
		return dto.AlertCustomWebhookResolvedConfig{}, err
	}

	resolved := dto.AlertCustomWebhookResolvedConfig{
		SchemaVersion: stored.SchemaVersion,
		DisplayName:   stored.DisplayName,
		Preset:        stored.Preset,
		Method:        stored.Method,
		URL:           secret.URL,
		Body:          stored.Body,
		Headers:       make([]dto.AlertCustomWebhookResolvedHeader, 0, len(stored.Headers)),
	}
	for _, header := range stored.Headers {
		value := header.Value
		if header.Secret {
			value = secret.Headers[header.UID]
			if value == "" && !header.Configured {
				continue
			}
			if value == "" {
				return dto.AlertCustomWebhookResolvedConfig{}, fmt.Errorf("secret header %q is not configured", header.Key)
			}
		}
		resolved.Headers = append(resolved.Headers, dto.AlertCustomWebhookResolvedHeader{Key: header.Key, Value: value})
	}
	return resolved, nil
}

func ValidateStored(config model.AlertConfig) error {
	if config.Type != constant.Custom {
		return fmt.Errorf("alert config %d is not a custom webhook", config.ID)
	}
	stored, err := decodeStored(config.Config)
	if err != nil {
		return err
	}
	return validateStoredSecret(config, stored)
}

func ExportSecretPlain(config model.AlertConfig) (string, error) {
	if config.Type != constant.Custom {
		return "", fmt.Errorf("alert config %d is not a custom webhook", config.ID)
	}
	plainText, err := decryptSecretPlain(config.SecretConfig)
	if err != nil {
		return "", err
	}
	secret, err := decodeAndValidateSecret(plainText)
	if err != nil {
		return "", err
	}
	canonical, err := json.Marshal(secret)
	if err != nil {
		return "", fmt.Errorf("encode custom webhook secret: %w", err)
	}
	return string(canonical), nil
}

func ImportSecretPlain(plainText string) (string, error) {
	secret, err := decodeAndValidateSecret(plainText)
	if err != nil {
		return "", err
	}
	canonical, err := json.Marshal(secret)
	if err != nil {
		return "", fmt.Errorf("encode custom webhook secret: %w", err)
	}
	return encryptSecretPlain(string(canonical))
}

func ReencryptSecret(cipherText string) (string, error) {
	plainText, err := decryptSecretPlain(cipherText)
	if err != nil {
		return "", err
	}
	return ImportSecretPlain(plainText)
}

func mergeAndValidate(
	mutation dto.AlertCustomWebhookConfig,
	status string,
	existingSecret dto.AlertCustomWebhookSecretConfig,
	hasExisting bool,
) (dto.AlertCustomWebhookConfig, dto.AlertCustomWebhookSecretConfig, error) {
	mutation.DisplayName = strings.TrimSpace(mutation.DisplayName)
	mutation.Preset = strings.TrimSpace(mutation.Preset)
	if mutation.DisplayName == "" {
		return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook displayName is required")
	}
	if utf8.RuneCountInString(mutation.DisplayName) > maxDisplayNameRunes {
		return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook displayName exceeds %d characters", maxDisplayNameRunes)
	}
	if _, ok := supportedPresets[mutation.Preset]; !ok {
		return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("unsupported custom webhook preset: %s", mutation.Preset)
	}
	if mutation.Method != http.MethodPost {
		return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook method must be POST")
	}
	if err := validateBody(&mutation); err != nil {
		return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, err
	}

	secret := dto.AlertCustomWebhookSecretConfig{
		SchemaVersion: dto.AlertCustomWebhookSchemaVersion,
		Headers:       make(map[string]string),
	}
	switch mutation.URL.Action {
	case "keep":
		if !hasExisting || existingSecret.URL == "" {
			return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook URL cannot be kept because it is not configured")
		}
		secret.URL = existingSecret.URL
	case "replace":
		secret.URL = strings.TrimSpace(mutation.URL.Value)
		if err := validateURL(secret.URL); err != nil {
			return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, err
		}
	case "clear":
		secret.URL = ""
	default:
		return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook URL action must be keep, replace, or clear")
	}
	if secret.URL == "" && status != constant.AlertDisable {
		return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook URL is required while the config is enabled")
	}
	mutation.URL = dto.AlertCustomWebhookURL{Configured: secret.URL != ""}

	if len(mutation.Headers) > maxHeaders {
		return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook headers exceed the limit of %d", maxHeaders)
	}
	seenUIDs := make(map[string]struct{}, len(mutation.Headers))
	seenKeys := make(map[string]struct{}, len(mutation.Headers))
	storedHeaders := make([]dto.AlertCustomWebhookHeader, 0, len(mutation.Headers))
	for _, header := range mutation.Headers {
		header.UID = strings.TrimSpace(header.UID)
		if header.UID == "" {
			header.UID = uuid.NewString()
		} else if _, err := uuid.Parse(header.UID); err != nil {
			return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook header uid must be a UUID")
		}
		if _, exists := seenUIDs[header.UID]; exists {
			return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("duplicate custom webhook header uid: %s", header.UID)
		}
		seenUIDs[header.UID] = struct{}{}

		header.Key = http.CanonicalHeaderKey(strings.TrimSpace(header.Key))
		if err := validateHeaderName(header.Key); err != nil {
			return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, err
		}
		if isSensitiveHeaderName(header.Key) && !header.Secret {
			return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook header %q must be marked secret", header.Key)
		}
		keyIdentity := strings.ToLower(header.Key)
		if _, exists := seenKeys[keyIdentity]; exists {
			return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("duplicate custom webhook header: %s", header.Key)
		}
		seenKeys[keyIdentity] = struct{}{}

		storedHeader := dto.AlertCustomWebhookHeader{UID: header.UID, Key: header.Key, Secret: header.Secret}
		if header.Secret {
			switch header.Action {
			case "keep":
				value, ok := existingSecret.Headers[header.UID]
				if !hasExisting || !ok || value == "" {
					return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("secret header %q cannot be kept because it is not configured", header.Key)
				}
				secret.Headers[header.UID] = value
			case "replace":
				if header.Value == "" {
					return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("secret header %q value is required", header.Key)
				}
				if err := validateHeaderValue(header.Value); err != nil {
					return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("header %q value is invalid", header.Key)
				}
				secret.Headers[header.UID] = header.Value
			case "clear":
			default:
				return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("secret header %q action must be keep, replace, or clear", header.Key)
			}
			storedHeader.Configured = secret.Headers[header.UID] != ""
		} else {
			if header.Action != "replace" {
				return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("non-secret header %q action must be replace", header.Key)
			}
			if err := validateHeaderValue(header.Value); err != nil {
				return dto.AlertCustomWebhookConfig{}, dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("header %q value is invalid", header.Key)
			}
			storedHeader.Value = header.Value
		}
		storedHeaders = append(storedHeaders, storedHeader)
	}
	mutation.Headers = storedHeaders
	mutation.SchemaVersion = dto.AlertCustomWebhookSchemaVersion
	return mutation, secret, nil
}

func validateStoredSecret(config model.AlertConfig, stored dto.AlertCustomWebhookConfig) error {
	_, err := validatedStoredSecret(config, stored)
	return err
}

func validatedStoredSecret(config model.AlertConfig, stored dto.AlertCustomWebhookConfig) (dto.AlertCustomWebhookSecretConfig, error) {
	if config.Status != constant.AlertEnable && config.Status != constant.AlertDisable {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("stored custom webhook status is invalid")
	}
	if strings.TrimSpace(config.SecretConfig) == "" {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("stored custom webhook secret is missing")
	}
	secret, err := decryptSecret(config.SecretConfig)
	if err != nil {
		return dto.AlertCustomWebhookSecretConfig{}, err
	}
	if secret.URL != "" {
		if err := validateURL(secret.URL); err != nil {
			return dto.AlertCustomWebhookSecretConfig{}, err
		}
	}
	if stored.URL.Configured != (secret.URL != "") {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("stored custom webhook URL state does not match its secret")
	}
	if config.Status == constant.AlertEnable && secret.URL == "" {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook URL is required while the config is enabled")
	}
	usedSecrets := make(map[string]struct{}, len(stored.Headers))
	for _, header := range stored.Headers {
		if !header.Secret {
			continue
		}
		value, configured := secret.Headers[header.UID]
		if header.Configured != (configured && value != "") {
			return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("stored custom webhook secret header state is inconsistent")
		}
		if configured {
			usedSecrets[header.UID] = struct{}{}
		}
	}
	if len(usedSecrets) != len(secret.Headers) {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("stored custom webhook contains orphaned header secrets")
	}
	return secret, nil
}

func validateBody(config *dto.AlertCustomWebhookConfig) error {
	config.Body.Type = strings.TrimSpace(config.Body.Type)
	if config.Preset != "custom" && config.Body.Type != "json" {
		return fmt.Errorf("custom webhook preset %s requires a JSON body", config.Preset)
	}
	switch config.Body.Type {
	case "json":
		if config.Body.Template == "" {
			return fmt.Errorf("json custom webhook body template is required")
		}
		if len(config.Body.Template) > maxBodyLength {
			return fmt.Errorf("custom webhook body exceeds %d bytes", maxBodyLength)
		}
		if !json.Valid([]byte(config.Body.Template)) {
			return fmt.Errorf("custom webhook JSON body template must be valid JSON")
		}
		var document any
		if err := json.Unmarshal([]byte(config.Body.Template), &document); err != nil {
			return fmt.Errorf("custom webhook JSON body template must be valid JSON")
		}
		if containsTemplateJSONKey(document) {
			return fmt.Errorf("custom webhook JSON body keys cannot contain template variables")
		}
		if err := validateTemplateVariables(config.Body.Template); err != nil {
			return err
		}
		if len(config.Body.Fields) != 0 {
			return fmt.Errorf("json custom webhook body does not accept form fields")
		}
	case "form":
		if config.Body.Template != "" {
			return fmt.Errorf("form custom webhook body does not accept a template")
		}
		if len(config.Body.Fields) == 0 {
			return fmt.Errorf("form custom webhook body requires at least one field")
		}
		if len(config.Body.Fields) > maxFormFields {
			return fmt.Errorf("custom webhook form fields exceed the limit of %d", maxFormFields)
		}
		seen := make(map[string]struct{}, len(config.Body.Fields))
		for index := range config.Body.Fields {
			field := &config.Body.Fields[index]
			field.Key = strings.TrimSpace(field.Key)
			if field.Key == "" {
				return fmt.Errorf("custom webhook form field key is required")
			}
			if strings.Contains(field.Key, "{{") || strings.Contains(field.Key, "}}") {
				return fmt.Errorf("custom webhook form field keys cannot contain template variables")
			}
			if len(field.Key) > maxHeaderNameLength || len(field.Value) > maxHeaderValueLen {
				return fmt.Errorf("custom webhook form field %q is too long", field.Key)
			}
			identity := strings.ToLower(field.Key)
			if _, exists := seen[identity]; exists {
				return fmt.Errorf("duplicate custom webhook form field: %s", field.Key)
			}
			seen[identity] = struct{}{}
			if err := validateTemplateVariables(field.Value); err != nil {
				return err
			}
		}
	case "text":
		if config.Body.Template == "" {
			return fmt.Errorf("text custom webhook body template is required")
		}
		if len(config.Body.Template) > maxBodyLength {
			return fmt.Errorf("custom webhook body exceeds %d bytes", maxBodyLength)
		}
		if err := validateTemplateVariables(config.Body.Template); err != nil {
			return err
		}
		if len(config.Body.Fields) != 0 {
			return fmt.Errorf("text custom webhook body does not accept form fields")
		}
	default:
		return fmt.Errorf("custom webhook body type must be json, form, or text")
	}
	return nil
}

func containsTemplateJSONKey(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if strings.Contains(key, "{{") || strings.Contains(key, "}}") || containsTemplateJSONKey(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsTemplateJSONKey(child) {
				return true
			}
		}
	}
	return false
}

func validateTemplateVariables(template string) error {
	remainder := template
	for {
		start := strings.Index(remainder, "{{")
		if start < 0 {
			if strings.Contains(remainder, "}}") {
				return fmt.Errorf("custom webhook body contains a malformed template variable")
			}
			return nil
		}
		if strings.Contains(remainder[:start], "}}") {
			return fmt.Errorf("custom webhook body contains a malformed template variable")
		}
		endOffset := strings.Index(remainder[start+2:], "}}")
		if endOffset < 0 {
			return fmt.Errorf("custom webhook body contains a malformed template variable")
		}
		end := start + 2 + endOffset + 2
		variable := remainder[start:end]
		if _, supported := supportedTemplateVariables[variable]; !supported {
			return fmt.Errorf("custom webhook body contains an unsupported template variable")
		}
		remainder = remainder[end:]
	}
}

func ContentTypeForBodyType(bodyType string) (string, error) {
	switch bodyType {
	case "json":
		return "application/json", nil
	case "form":
		return "application/x-www-form-urlencoded", nil
	case "text":
		return "text/plain", nil
	default:
		return "", fmt.Errorf("custom webhook body type must be json, form, or text")
	}
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("custom webhook URL is required")
	}
	if len(rawURL) > maxURLLength {
		return fmt.Errorf("custom webhook URL exceeds %d bytes", maxURLLength)
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("invalid custom webhook URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("custom webhook URL scheme must be http or https")
	}
	hostname := strings.TrimSuffix(strings.TrimSpace(parsed.Hostname()), ".")
	if parsed.Host == "" || hostname == "" || strings.Contains(hostname, "%") {
		return fmt.Errorf("custom webhook URL host is required")
	}
	if strings.HasSuffix(parsed.Host, ":") {
		return fmt.Errorf("custom webhook URL port is invalid")
	}
	if port := parsed.Port(); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil || value < 1 || value > 65535 {
			return fmt.Errorf("custom webhook URL port is invalid")
		}
	}
	if parsed.User != nil {
		return fmt.Errorf("custom webhook URL must not contain user info")
	}
	if parsed.Fragment != "" {
		return fmt.Errorf("custom webhook URL must not contain a fragment")
	}
	return nil
}

func validateHeaderName(name string) error {
	if name == "" {
		return fmt.Errorf("custom webhook header name is required")
	}
	if len(name) > maxHeaderNameLength {
		return fmt.Errorf("custom webhook header name is too long")
	}
	for index := 0; index < len(name); index++ {
		if !isHeaderTokenByte(name[index]) {
			return fmt.Errorf("invalid custom webhook header name: %s", name)
		}
	}
	if _, forbidden := forbiddenHeaders[strings.ToLower(name)]; forbidden {
		return fmt.Errorf("custom webhook header %q is managed by the sender", name)
	}
	return nil
}

func validateHeaderValue(value string) error {
	if len(value) > maxHeaderValueLen || strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("invalid custom webhook header value")
	}
	return nil
}

func isSensitiveHeaderName(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	compact := strings.NewReplacer("-", "", "_", "").Replace(normalized)
	return normalized == "authorization" ||
		normalized == "cookie" ||
		strings.Contains(compact, "token") ||
		strings.Contains(compact, "secret") ||
		strings.Contains(compact, "signature") ||
		strings.Contains(compact, "apikey") ||
		strings.HasSuffix(normalized, "-key") ||
		strings.HasSuffix(normalized, "_key")
}

func isHeaderTokenByte(char byte) bool {
	if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' {
		return true
	}
	return strings.ContainsRune("!#$%&'*+-.^_`|~", rune(char))
}

func decodeStored(rawConfig string) (dto.AlertCustomWebhookConfig, error) {
	if len(rawConfig) > maxConfigLength {
		return dto.AlertCustomWebhookConfig{}, fmt.Errorf("stored custom webhook config exceeds %d bytes", maxConfigLength)
	}
	var stored dto.AlertCustomWebhookConfig
	if err := decodeStrict(rawConfig, &stored); err != nil {
		return dto.AlertCustomWebhookConfig{}, fmt.Errorf("decode stored custom webhook config: %w", err)
	}
	if stored.SchemaVersion != dto.AlertCustomWebhookSchemaVersion {
		return dto.AlertCustomWebhookConfig{}, fmt.Errorf("unsupported stored custom webhook schemaVersion: %d", stored.SchemaVersion)
	}
	if stored.State != "" {
		return dto.AlertCustomWebhookConfig{}, fmt.Errorf("stored custom webhook contains a view state")
	}
	if stored.URL.Action != "" || stored.URL.Value != "" {
		return dto.AlertCustomWebhookConfig{}, fmt.Errorf("stored custom webhook URL contains a mutation")
	}
	for _, header := range stored.Headers {
		if header.Action != "" || header.Secret && header.Value != "" {
			return dto.AlertCustomWebhookConfig{}, fmt.Errorf("stored custom webhook header %q contains a mutation", header.Key)
		}
	}
	if err := validateStoredConfig(&stored); err != nil {
		return dto.AlertCustomWebhookConfig{}, err
	}
	return stored, nil
}

func validateStoredConfig(config *dto.AlertCustomWebhookConfig) error {
	config.DisplayName = strings.TrimSpace(config.DisplayName)
	if config.DisplayName == "" || utf8.RuneCountInString(config.DisplayName) > maxDisplayNameRunes {
		return fmt.Errorf("stored custom webhook displayName is invalid")
	}
	if _, ok := supportedPresets[config.Preset]; !ok {
		return fmt.Errorf("unsupported stored custom webhook preset: %s", config.Preset)
	}
	if config.Method != http.MethodPost {
		return fmt.Errorf("stored custom webhook method must be POST")
	}
	if err := validateBody(config); err != nil {
		return err
	}
	seenUIDs := make(map[string]struct{}, len(config.Headers))
	seenKeys := make(map[string]struct{}, len(config.Headers))
	for index := range config.Headers {
		header := &config.Headers[index]
		if header.UID == "" {
			return fmt.Errorf("stored custom webhook header uid is required")
		}
		if _, err := uuid.Parse(header.UID); err != nil {
			return fmt.Errorf("stored custom webhook header uid must be a UUID")
		}
		if _, exists := seenUIDs[header.UID]; exists {
			return fmt.Errorf("duplicate stored custom webhook header uid: %s", header.UID)
		}
		seenUIDs[header.UID] = struct{}{}
		header.Key = http.CanonicalHeaderKey(strings.TrimSpace(header.Key))
		if err := validateHeaderName(header.Key); err != nil {
			return err
		}
		if isSensitiveHeaderName(header.Key) && !header.Secret {
			return fmt.Errorf("stored custom webhook header %q must be marked secret", header.Key)
		}
		identity := strings.ToLower(header.Key)
		if _, exists := seenKeys[identity]; exists {
			return fmt.Errorf("duplicate stored custom webhook header: %s", header.Key)
		}
		seenKeys[identity] = struct{}{}
		if !header.Secret {
			if err := validateHeaderValue(header.Value); err != nil {
				return fmt.Errorf("stored header %q value is invalid", header.Key)
			}
		}
	}
	return nil
}

func decryptSecret(cipherText string) (dto.AlertCustomWebhookSecretConfig, error) {
	if strings.TrimSpace(cipherText) == "" {
		return dto.AlertCustomWebhookSecretConfig{
			SchemaVersion: dto.AlertCustomWebhookSchemaVersion,
			Headers:       map[string]string{},
		}, nil
	}
	plainText, err := decryptSecretPlain(cipherText)
	if err != nil {
		return dto.AlertCustomWebhookSecretConfig{}, err
	}
	return decodeAndValidateSecret(plainText)
}

func encryptSecretPlain(plainText string) (string, error) {
	coreKey, coreAvailable, err := loadCoreEncryptKey()
	if err != nil {
		return "", err
	}
	if coreAvailable {
		cipherText, err := encrypt.StringEncryptWithKey(plainText, coreKey)
		if err != nil {
			return "", fmt.Errorf("encrypt custom webhook secret with core key: %w", err)
		}
		return coreKeyPrefix + cipherText, nil
	}
	key, err := loadAgentEncryptKey()
	if err != nil {
		return "", err
	}
	cipherText, err := encrypt.StringEncryptWithKey(plainText, key)
	if err != nil {
		return "", fmt.Errorf("encrypt custom webhook secret with agent key: %w", err)
	}
	return agentKeyPrefix + cipherText, nil
}

func decryptSecretPlain(cipherText string) (string, error) {
	if strings.TrimSpace(cipherText) == "" {
		return "", fmt.Errorf("custom webhook secret is empty")
	}
	var (
		payload string
		key     string
		err     error
	)
	switch {
	case strings.HasPrefix(cipherText, coreKeyPrefix):
		payload = strings.TrimPrefix(cipherText, coreKeyPrefix)
		var available bool
		key, available, err = loadCoreEncryptKey()
		if err != nil {
			return "", err
		}
		if !available {
			return "", fmt.Errorf("decrypt custom webhook secret: core encrypt key is unavailable")
		}
	case strings.HasPrefix(cipherText, agentKeyPrefix):
		payload = strings.TrimPrefix(cipherText, agentKeyPrefix)
		key, err = loadAgentEncryptKey()
		if err != nil {
			return "", err
		}
	default:
		payload = cipherText
		key, err = loadAgentEncryptKey()
		if err != nil {
			return "", err
		}
	}
	if payload == "" {
		return "", fmt.Errorf("decrypt custom webhook secret: ciphertext is empty")
	}
	plainText, err := encrypt.StringDecryptWithKey(payload, key)
	if err != nil {
		return "", fmt.Errorf("decrypt custom webhook secret: %w", err)
	}
	return plainText, nil
}

func loadCoreEncryptKey() (string, bool, error) {
	if global.CoreDB == nil {
		return "", false, nil
	}
	var setting model.Setting
	if err := global.CoreDB.Where("key = ?", "EncryptKey").First(&setting).Error; err != nil {
		return "", true, fmt.Errorf("custom webhook core encrypt key is unavailable")
	}
	key := strings.TrimSpace(setting.Value)
	if key == "" {
		return "", true, fmt.Errorf("custom webhook core encrypt key is empty")
	}
	return key, true, nil
}

func loadAgentEncryptKey() (string, error) {
	if key := strings.TrimSpace(global.CONF.Base.EncryptKey); key != "" {
		return key, nil
	}
	if global.DB != nil {
		var setting model.Setting
		if err := global.DB.Where("key = ?", "EncryptKey").First(&setting).Error; err == nil {
			if key := strings.TrimSpace(setting.Value); key != "" {
				return key, nil
			}
		}
	}
	return "", fmt.Errorf("custom webhook agent encrypt key is empty")
}

func decodeAndValidateSecret(plainText string) (dto.AlertCustomWebhookSecretConfig, error) {
	if len(plainText) > maxSecretLength {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook secret exceeds %d bytes", maxSecretLength)
	}
	var secret dto.AlertCustomWebhookSecretConfig
	if err := decodeStrict(plainText, &secret); err != nil {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("decode custom webhook secret: %w", err)
	}
	if secret.SchemaVersion != dto.AlertCustomWebhookSchemaVersion {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("unsupported custom webhook secret schemaVersion: %d", secret.SchemaVersion)
	}
	if secret.URL != "" {
		if err := validateURL(secret.URL); err != nil {
			return dto.AlertCustomWebhookSecretConfig{}, err
		}
	}
	if secret.Headers == nil {
		secret.Headers = map[string]string{}
	}
	if len(secret.Headers) > maxHeaders {
		return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook secret headers exceed the limit of %d", maxHeaders)
	}
	for uid, value := range secret.Headers {
		if _, err := uuid.Parse(strings.TrimSpace(uid)); err != nil {
			return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook secret header uid must be a UUID")
		}
		if value == "" {
			return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook secret header %q is empty", uid)
		}
		if err := validateHeaderValue(value); err != nil {
			return dto.AlertCustomWebhookSecretConfig{}, fmt.Errorf("custom webhook secret header %q is invalid", uid)
		}
	}
	return secret, nil
}

func decodeStrict(raw string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}
