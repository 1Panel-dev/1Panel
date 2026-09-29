package terminal_session

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
)

const CapabilityPath = "/api/v2/internal/terminal/capabilities"

func RequiresLeaseCapability(c *gin.Context) bool {
	if c == nil || c.Request == nil || !c.GetBool("API_AUTH") || c.GetString("API_AUTH_KEY_KIND") != "apiKey" {
		return false
	}
	switch c.Request.URL.Path {
	case "/api/v2/hosts/terminal/local", "/api/v2/hosts/terminal/ssh", "/api/v2/hosts/terminal/container":
		return true
	default:
		return false
	}
}

func CheckLeaseCapability(fetch func() (interface{}, error)) error {
	data, err := fetch()
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var capabilities struct {
		APIKeyLeaseVersion int `json:"apiKeyLeaseVersion"`
	}
	if err := json.Unmarshal(encoded, &capabilities); err != nil {
		return err
	}
	if capabilities.APIKeyLeaseVersion < 1 {
		return errors.New("Agent does not support API key terminal authorization leases")
	}
	return nil
}
