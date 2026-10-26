package service

import (
	"encoding/json"
	"errors"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
)

func encodeMCPIPs(entries []string) string {
	data, _ := json.Marshal(entries)
	if entries == nil {
		return "[]"
	}
	return string(data)
}
func normalizeMCPAccessSettings(req *dto.MCPRuntimeSettings) error {
	entries, err := contract.NormalizeAllowedIPs(req.AllowedIPs)
	if err != nil {
		return err
	}
	switch req.AccessPolicy {
	case "inherit", "all":
		if len(entries) != 0 {
			return errors.New("independent MCP IP entries require restricted policy")
		}
	case "restricted":
		if len(entries) == 0 {
			return errors.New("MCP restricted policy requires at least one IP or CIDR")
		}
	default:
		return errors.New("invalid MCP access policy")
	}
	req.AllowedIPs = entries
	return nil
}
