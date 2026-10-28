package mcpruntime

import (
	"context"
	"errors"
	"regexp"
)

type PullProgress struct {
	Status  string `json:"status,omitempty"`
	ID      string `json:"id,omitempty"`
	Current int64  `json:"current,omitempty"`
	Total   int64  `json:"total,omitempty"`
}

type PullEvent struct {
	Type string `json:"type"`
	PullProgress
	Error string `json:"error,omitempty"`
}

var pullLayerID = regexp.MustCompile(`^[a-f0-9]{12,64}$`)

func reportPullProgress(ctx context.Context, report func(PullProgress) error, progress PullProgress) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if report != nil && report(progress) != nil {
		return errors.New("MCP Docker progress connection closed")
	}
	return nil
}

func normalizePullProgress(status, id string, current, total int64) (PullProgress, bool) {
	var normalized string
	switch status {
	case "Pulling fs layer":
		normalized = "pulling"
	case "Waiting":
		normalized = "waiting"
	case "Downloading":
		normalized = "downloading"
	case "Verifying Checksum":
		normalized = "verifying_checksum"
	case "Download complete":
		normalized = "download_complete"
	case "Extracting":
		normalized = "extracting"
	case "Pull complete":
		normalized = "pull_complete"
	case "Already exists":
		normalized = "already_exists"
	default:
		return PullProgress{}, false
	}
	if !pullLayerID.MatchString(id) {
		id = ""
	}
	if current < 0 || total < 0 || (total > 0 && current > total) {
		current, total = 0, 0
	}
	return PullProgress{Status: normalized, ID: id, Current: current, Total: total}, true
}
