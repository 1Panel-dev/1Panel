package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	agenti18n "github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/gin-gonic/gin"
)

func TestHandleDockerPortGuardErrorReturnsStableBusinessCode(t *testing.T) {
	agenti18n.Init()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	handleDockerPortGuardError(context, fmt.Errorf("normalize policy: %w", service.ErrDockerGuardInvalid))

	var response dto.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != http.StatusBadRequest || response.ErrorCode != "FW_DOCKER_GUARD_INVALID" {
		t.Fatalf("unexpected Docker guard error response: %#v", response)
	}
}

func TestHandleDockerPortGuardErrorLocalizesDockerUnavailable(t *testing.T) {
	agenti18n.Init()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	handleDockerPortGuardError(context, fmt.Errorf("inspect Docker: %w", service.ErrDockerUnavailable))

	var response dto.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != http.StatusServiceUnavailable || response.ErrorCode != "FW_DOCKER_UNAVAILABLE" {
		t.Fatalf("unexpected Docker unavailable response: %#v", response)
	}
	if response.Message != agenti18n.Get("ErrDockerFailed") {
		t.Fatalf("message = %q, want localized Docker failure", response.Message)
	}
}
