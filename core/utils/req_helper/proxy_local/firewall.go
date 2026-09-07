package proxy_local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func UpdatePanelPort(oldPort string, newPort uint) error {
	previousPort, err := strconv.ParseUint(oldPort, 10, 16)
	if err != nil || previousPort == 0 {
		return fmt.Errorf("invalid old panel port %q", oldPort)
	}
	if newPort == 0 || newPort > 65535 {
		return fmt.Errorf("invalid new panel port %d", newPort)
	}
	body, err := json.Marshal(struct {
		OldPort uint `json:"oldPort"`
		NewPort uint `json:"newPort"`
	}{OldPort: uint(previousPort), NewPort: newPort})
	if err != nil {
		return err
	}
	_, err = NewLocalClient("/api/v2/hosts/firewall/port", http.MethodPost, bytes.NewReader(body), nil)
	return err
}
