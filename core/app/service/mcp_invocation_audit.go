package service

import (
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
)

// SQLite stores time.Time with the server's offset. Compare absolute seconds
// and integer nanoseconds so offset changes and repeated local hours remain
// chronological without changing the stored timestamps or losing precision.
const mcpInvocationSecond = "CAST(strftime('%s', substr(started_at, 1, 19) || substr(started_at, -6)) AS INTEGER)"
const mcpInvocationNanosecond = "CASE WHEN substr(started_at, 20, 1) = '.' THEN CAST(substr(substr(started_at, 21, length(started_at) - 26) || '000000000', 1, 9) AS INTEGER) ELSE 0 END"

func (s *MCPService) finishInvocationAudit(record model.MCPInvocation) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 50 * time.Millisecond)
		}
		// Only compensate the audit write. The business operation must never be replayed.
		err = s.DB.Model(&model.MCPInvocation{}).Where("id = ? AND status = ?", record.ID, "running").
			Updates(map[string]any{"status": record.Status, "code": record.Code, "finished_at": record.FinishedAt, "duration_ms": record.DurationMs}).Error
		if err == nil {
			return nil
		}
	}
	return err
}

// RecoverInterruptedInvocations runs before any MCP ingress starts. The fixed
// process-start boundary prevents a later invocation from being marked interrupted.
func (s *MCPService) RecoverInterruptedInvocations(processStartedAt time.Time) error {
	now := s.Now()
	return s.DB.Model(&model.MCPInvocation{}).
		Where("status = ? AND finished_at IS NULL", "running").
		Where("("+mcpInvocationSecond+" < ? OR ("+mcpInvocationSecond+" = ? AND "+mcpInvocationNanosecond+" < ?))", processStartedAt.Unix(), processStartedAt.Unix(), processStartedAt.Nanosecond()).
		Updates(map[string]any{"status": "failed", "code": "MCP_OUTCOME_UNKNOWN", "finished_at": now}).Error
}
