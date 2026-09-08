package middleware

import (
	"net/http"
	"testing"
)

func TestVMOperationPathUsesXpackDatabase(t *testing.T) {
	const prefix = "/api/v2/core/xpack/vms"
	for _, suffix := range []string{"", "/iso/del", "/networks/update", "/storages/del", "/templates/del"} {
		if actual := normalizeOperationPath(prefix + suffix); actual != "/core/xpack/vms"+suffix {
			t.Errorf("wrong operation/database path: %s", actual)
		}
	}
	if isDemoRequestAllowed(http.MethodGet, prefix+"/console/ws") {
		t.Errorf("demo mode permits VM console: %s", prefix)
	}
}
