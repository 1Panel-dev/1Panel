package service

import "testing"

func TestParseSSHLogLineKeepsSessionKeyAcrossSSHDProcesses(t *testing.T) {
	accepted, ok := parseSSHLogLine("2026-08-07T14:13:01+08:00 localhost sshd[2204243]: Accepted publickey for root from 10.128.128.21 port 54995 ssh2")
	if !ok {
		t.Fatal("expected accepted SSH log line to parse")
	}

	disconnect, ok := parseSSHLogLine("2026-08-07T14:13:10+08:00 localhost sshd[2204261]: Received disconnect from 10.128.128.21 port 54995:11: disconnected by user")
	if !ok {
		t.Fatal("expected disconnect SSH log line to parse")
	}

	const sessionKey = "10.128.128.21:54995"
	if accepted.SessionKey != sessionKey {
		t.Fatalf("expected accepted log session key %q, got %q", sessionKey, accepted.SessionKey)
	}
	if disconnect.SessionKey != sessionKey {
		t.Fatalf("expected disconnect log session key %q, got %q", sessionKey, disconnect.SessionKey)
	}

	filtered := filterRedundantSSHSessionItems(
		[]sshParsedLog{accepted, disconnect},
		map[string]bool{accepted.SessionKey: true},
	)
	if len(filtered) != 1 {
		t.Fatalf("expected normal disconnect to be removed after matching accepted log, got %d items", len(filtered))
	}
	if filtered[0].History.Status != accepted.History.Status {
		t.Fatalf("expected accepted log to remain, got status %q", filtered[0].History.Status)
	}
}
