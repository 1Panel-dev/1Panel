package alert

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/webhook_sender"
)

func CustomWebhookTestResult(request webhook_sender.Request, result webhook_sender.Result, executeErr error) dto.AlertConfigTestResult {
	durationMillis := result.Duration.Milliseconds()
	if result.Duration > 0 && durationMillis == 0 {
		durationMillis = 1
	}
	testResult := dto.AlertConfigTestResult{
		Success:    executeErr == nil,
		StatusCode: result.StatusCode,
		Duration:   durationMillis,
	}
	if executeErr == nil {
		return testResult
	}
	if global.LOG != nil {
		global.LOG.Warnf("webhook test failed: statusCode=%d durationMs=%d error=%q response=%q",
			result.StatusCode, durationMillis,
			webhook_sender.DiagnosticText(executeErr.Error(), request),
			webhook_sender.DiagnosticText(result.Response, request))
	}
	return testResult
}
