package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func TestNormalizeFirewallRuleUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("trims valid UUID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		value := "  managed-rule  "
		if !normalizeFirewallRuleUUID(context, &value) || value != "managed-rule" {
			t.Fatalf("normalized UUID = %q", value)
		}
	})
	t.Run("rejects blank UUID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		value := "   "
		if normalizeFirewallRuleUUID(context, &value) {
			t.Fatal("blank UUID was accepted")
		}
		var response dto.Response
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if recorder.Code != http.StatusOK || response.Code != http.StatusBadRequest {
			t.Fatalf("transport status = %d, response = %#v", recorder.Code, response)
		}
	})
}
