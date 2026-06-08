package service

import (
	"strings"
	"testing"
)

func TestParseSupervisorRunningDetails(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantPID    string
		wantUptime string
	}{
		{
			name:       "short uptime",
			line:       "test:test_00 RUNNING pid 123, uptime 0:12:34",
			wantPID:    "123",
			wantUptime: "0:12:34",
		},
		{
			name:       "single day uptime",
			line:       "test:test_00 RUNNING pid 123, uptime 1 day, 0:12:34",
			wantPID:    "123",
			wantUptime: "1 day, 0:12:34",
		},
		{
			name:       "multiple days uptime",
			line:       "test:test_00 RUNNING pid 123, uptime 103 days, 0:12:34",
			wantPID:    "123",
			wantUptime: "103 days, 0:12:34",
		},
		{
			name:       "process name is uptime",
			line:       "uptime RUNNING pid 123, uptime 0:12:34",
			wantPID:    "123",
			wantUptime: "0:12:34",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pid, uptime := parseSupervisorRunningDetails(strings.Fields(tt.line))
			if pid != tt.wantPID {
				t.Fatalf("pid = %q, want %q", pid, tt.wantPID)
			}
			if uptime != tt.wantUptime {
				t.Fatalf("uptime = %q, want %q", uptime, tt.wantUptime)
			}
		})
	}
}
