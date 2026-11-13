package provider

import (
	"net/url"
	"strings"
)

func BuildTextRequestURL(apiType, baseURL, model string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	path := strings.TrimRight(parsed.Path, "/")
	var endpoint string
	switch apiType {
	case "openai-completions":
		endpoint = "/chat/completions"
	case "openai-responses":
		endpoint = "/responses"
	case "anthropic-messages":
		endpoint = "/messages"
	case "gemini-generate-content":
		if strings.Contains(path, "/models/") && strings.HasSuffix(path, ":generateContent") {
			return parsed.String()
		}
		if !strings.HasSuffix(path, "/v1beta") {
			path += "/v1beta"
		}
		parsed.Path = path + "/models/" + strings.TrimSpace(model) + ":generateContent"
		parsed.RawPath = ""
		return parsed.String()
	default:
		return baseURL
	}
	if strings.HasSuffix(path, endpoint) {
		return parsed.String()
	}
	versioned := false
	for _, segment := range strings.Split(path, "/") {
		if hasAPIVersionSuffix(segment) || segment == "v1beta" {
			versioned = true
			break
		}
	}
	if (apiType == "anthropic-messages" && !strings.HasSuffix(path, "/v1")) || (apiType != "anthropic-messages" && !versioned) {
		path += "/v1"
	}
	parsed.Path = path + endpoint
	parsed.RawPath = ""
	return parsed.String()
}
