package v2

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	agenti18n "github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/gin-gonic/gin"
)

func TestHandleFirewallRuleErrorReturnsStableBusinessCode(t *testing.T) {
	agenti18n.Init()
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		err  error
		code string
	}{
		{name: "stale rule", err: filter.ErrRuleStale, code: "FW_RULE_STALE"},
		{name: "check required", err: filter.ErrRuleCheckRequired, code: "FW_RULE_CHECK_REQUIRED"},
		{name: "revision conflict", err: fmt.Errorf("persist rule: %w", repo.ErrFirewallRuleRevisionConflict), code: "FW_RULE_REVISION_CONFLICT"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			handleFirewallRuleError(context, test.err)
			var response dto.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Code != 409 || response.ErrorCode != test.code {
				t.Fatalf("unexpected firewall error response: %#v", response)
			}
		})
	}
}
