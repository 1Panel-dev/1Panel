package webhook_sender

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"
)

const MaxRenderedBodyBytes = 256 * 1024

type TemplateData struct {
	Title     string
	Message   string
	Type      string
	Timestamp time.Time
	NodeName  string
}

type FormField struct {
	Key   string
	Value string
}

type RenderRequest struct {
	Format   BodyFormat
	Template string
	Fields   []FormField
	Data     TemplateData
}

func RenderBody(input RenderRequest) ([]byte, error) {
	values := templateValues(input.Data)
	var rendered []byte
	var err error
	switch input.Format {
	case BodyJSON:
		rendered, err = renderJSONTemplate(input.Template, values)
		if err != nil {
			return nil, errors.New("rendered webhook JSON body is invalid")
		}
	case BodyText:
		var text string
		text, err = renderRestrictedTemplate(input.Template, values)
		if err != nil {
			return nil, err
		}
		rendered = []byte(text)
	case BodyForm:
		form := make(url.Values, len(input.Fields))
		for _, field := range input.Fields {
			if strings.Contains(field.Key, "{{") || strings.Contains(field.Key, "}}") {
				return nil, errors.New("webhook form field key cannot contain template variables")
			}
			value, err := renderRestrictedTemplate(field.Value, values)
			if err != nil {
				return nil, err
			}
			form.Add(field.Key, value)
		}
		rendered = []byte(form.Encode())
	default:
		return nil, errors.New("unsupported webhook body format")
	}
	if len(rendered) > MaxRenderedBodyBytes {
		return nil, errors.New("rendered webhook body exceeded size limit")
	}
	return rendered, nil
}

func ResolvePreset(value string) (Preset, error) {
	switch strings.TrimSpace(value) {
	case "", "generic", "genericJson", "custom":
		return PresetGeneric, nil
	case "slack":
		return PresetSlack, nil
	case "discord":
		return PresetDiscord, nil
	case "teams", "teamsWorkflows":
		return PresetTeams, nil
	default:
		return "", errors.New("unsupported webhook preset")
	}
}

func templateValues(data TemplateData) map[string]string {
	timestamp := data.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	return map[string]string{
		"title":     data.Title,
		"message":   data.Message,
		"type":      data.Type,
		"timestamp": timestamp.Format(time.RFC3339),
		"nodeName":  data.NodeName,
	}
}

func renderJSONTemplate(source string, values map[string]string) ([]byte, error) {
	decoder := json.NewDecoder(strings.NewReader(source))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		return nil, errors.New("invalid webhook JSON template")
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, err
	}
	rendered, err := renderJSONValue(document, values)
	if err != nil {
		return nil, err
	}
	result, err := json.Marshal(rendered)
	if err != nil {
		return nil, errors.New("marshal rendered webhook JSON body failed")
	}
	return result, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("invalid webhook JSON template")
	}
	return nil
}

func renderJSONValue(value any, values map[string]string) (any, error) {
	switch typed := value.(type) {
	case string:
		return renderRestrictedTemplate(typed, values)
	case []any:
		for index := range typed {
			rendered, err := renderJSONValue(typed[index], values)
			if err != nil {
				return nil, err
			}
			typed[index] = rendered
		}
		return typed, nil
	case map[string]any:
		for key, item := range typed {
			rendered, err := renderJSONValue(item, values)
			if err != nil {
				return nil, err
			}
			typed[key] = rendered
		}
		return typed, nil
	default:
		return value, nil
	}
}

func renderRestrictedTemplate(source string, values map[string]string) (string, error) {
	var output bytes.Buffer
	for len(source) != 0 {
		start := strings.Index(source, "{{")
		if start == -1 {
			if strings.Contains(source, "}}") {
				return "", errors.New("invalid webhook template placeholder")
			}
			output.WriteString(source)
			break
		}
		output.WriteString(source[:start])
		source = source[start+2:]
		end := strings.Index(source, "}}")
		if end == -1 {
			return "", errors.New("invalid webhook template placeholder")
		}
		name := strings.TrimSpace(source[:end])
		value, ok := values[name]
		if !ok || name == "" || strings.Contains(name, "{{") {
			return "", errors.New("unsupported webhook template variable")
		}
		output.WriteString(value)
		source = source[end+2:]
	}
	return output.String(), nil
}
