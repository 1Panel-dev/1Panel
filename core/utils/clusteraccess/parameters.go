package clusteraccess

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Authorizer interface {
	AuthorizeClusterNode(*gin.Context, string, bool) error
}

func IsClusterApp(key string) bool {
	switch key {
	case "mysql-cluster", "postgresql-cluster", "redis-cluster":
		return true
	}
	return false
}

type responseFilterKey struct{}

type filterKind uint8

const (
	detailFilter filterKind = iota + 1
	searchFilter
	connectionFilter
)

func PrepareParameterResponse(c *gin.Context, node string, authorizer Authorizer) error {
	r := c.Request
	var kind filterKind
	switch {
	case r.Method == http.MethodGet &&
		(strings.HasPrefix(path.Clean(r.URL.Path), "/api/v2/apps/installed/info/") ||
			strings.HasPrefix(path.Clean(r.URL.Path), "/api/v2/apps/installed/params/")):
		kind = detailFilter
	case r.Method == http.MethodPost && path.Clean(r.URL.Path) == "/api/v2/apps/installed/search":
		kind = searchFilter
	case r.Method == http.MethodPost && path.Clean(r.URL.Path) == "/api/v2/apps/installed/conninfo":
		kind = connectionFilter
	default:
		return nil
	}
	if node == "" {
		node = "local"
	}
	if authorizer.AuthorizeClusterNode(c, node, true) == nil {
		return nil
	}
	if kind == connectionFilter {
		if r.Body == nil {
			return nil
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))
		var req struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(body, &req); err != nil || !IsClusterApp(req.Type) {
			return nil
		}
	}
	r.Header.Del("Accept-Encoding")
	c.Request = r.WithContext(context.WithValue(r.Context(), responseFilterKey{}, kind))
	return nil
}

func FilterParameterResponse(request *http.Request, response *http.Response) error {
	if request == nil {
		return nil
	}
	kind, filtered := request.Context().Value(responseFilterKey{}).(filterKind)
	if !filtered || response.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	replacement, err := redactParameterResponse(body, kind)
	if err != nil {
		return errors.New("could not read installed app response")
	}
	response.Body = io.NopCloser(bytes.NewReader(replacement))
	if bytes.Equal(body, replacement) {
		return nil
	}
	response.ContentLength = int64(len(replacement))
	response.Header.Set("Content-Length", strconv.Itoa(len(replacement)))
	response.Header.Set("Cache-Control", "no-store")
	response.Header.Del("ETag")
	return nil
}

func redactParameterResponse(body []byte, kind filterKind) ([]byte, error) {
	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if raw := response["data"]; len(raw) == 0 || string(raw) == "null" {
		return body, nil
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(response["data"], &data); err != nil {
		return nil, err
	}
	changed := false
	switch kind {
	case searchFilter:
		if raw := data["items"]; len(raw) > 0 {
			var items []map[string]json.RawMessage
			if err := json.Unmarshal(raw, &items); err != nil {
				return nil, err
			}
			for _, item := range items {
				changed = redactClusterFields(item) || changed
			}
			if changed {
				data["items"], _ = json.Marshal(items)
			}
		}
	case connectionFilter:
		data["password"] = json.RawMessage(`""`)
		changed = true
	case detailFilter:
		changed = redactClusterFields(data)
	}
	if !changed {
		return body, nil
	}
	response["data"], _ = json.Marshal(data)
	return json.Marshal(response)
}

func redactClusterFields(fields map[string]json.RawMessage) bool {
	var key string
	_ = json.Unmarshal(fields["appKey"], &key)
	if key == "" {
		var app struct {
			Key string `json:"key"`
		}
		_ = json.Unmarshal(fields["app"], &app)
		key = app.Key
	}
	if !IsClusterApp(key) {
		return false
	}
	for name, empty := range map[string]string{"env": "{}", "param": `""`, "params": "[]", "dockerCompose": `""`, "rawCompose": `""`} {
		if _, present := fields[name]; present {
			fields[name] = json.RawMessage(empty)
		}
	}
	return true
}
