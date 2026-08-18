package iptables_helper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
)

func TestBuildBaseChainsRestoreScriptBatchesPersistedRules(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		BasicBeforeFileName: "-A " + BasicBeforeChain + " -i lo -j ACCEPT\n",
		BasicFileName:       "-A " + BasicChain + " -p tcp --dport 8080 -j ACCEPT\n",
		BasicAfterFileName:  "-A " + BasicAfterChain + " -p tcp -j DROP\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	script, err := buildBaseChainsRestoreScript(dir, "9443", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"*filter\n",
		"-F " + BasicBeforeChain + "\n",
		"-F " + BasicChain + "\n",
		"-F " + BasicAfterChain + "\n",
		files[BasicBeforeFileName],
		files[BasicFileName],
		files[BasicAfterFileName],
		"-A " + BasicBeforeChain + " -p tcp -m tcp --dport 9443 -j ACCEPT\n",
		"COMMIT\n",
	} {
		if !strings.Contains(script, expected) {
			t.Fatalf("restore script does not contain %q:\n%s", expected, script)
		}
	}
	if strings.Count(script, "COMMIT\n") != 1 {
		t.Fatalf("restore script is not a single batch:\n%s", script)
	}
}

func TestBuildBaseChainsRestoreScriptUsesIPv6Files(t *testing.T) {
	dir := t.TempDir()
	want := "-A " + BasicChain + " -p ipv6-icmp -j ACCEPT\n"
	if err := os.WriteFile(filepath.Join(dir, IPv6FileName(BasicFileName)), []byte(want), 0o600); err != nil {
		t.Fatal(err)
	}
	script, err := buildBaseChainsRestoreScript(dir, "9443", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, want) {
		t.Fatalf("IPv6 persisted rule was not restored:\n%s", script)
	}
}

func TestBuildRequiredPortsRestoreScriptBatchesAddsAndDeletes(t *testing.T) {
	desired := []firewall.PortWhitelist{{Protocol: "tcp", Port: "22"}, {Protocol: "udp", Port: "53"}}
	before := []FilterRules{
		{Chain: BasicBeforeChain, Protocol: "tcp", DstPort: "22", Strategy: "accept"},
		{Chain: BasicBeforeChain, Protocol: "tcp", DstPort: "22", Strategy: "accept"},
		{Chain: BasicBeforeChain, Protocol: "tcp", DstPort: "80", Strategy: "accept"},
	}
	after := []FilterRules{{Chain: BasicAfterChain, Protocol: "udp", DstPort: "5353", Strategy: "accept"}}
	script := buildRequiredPortsRestoreScript(desired, before, after, "", "", true)

	for _, line := range []string{
		"-D 1PANEL_BASIC_BEFORE -p tcp -m tcp --dport 22 -j ACCEPT",
		"-D 1PANEL_BASIC_BEFORE -p tcp -m tcp --dport 80 -j ACCEPT",
		"-D 1PANEL_BASIC_AFTER -p udp -m udp --dport 5353 -j ACCEPT",
		"-A 1PANEL_BASIC_BEFORE -p udp -m udp --dport 53 -j ACCEPT",
		"-A 1PANEL_BASIC_AFTER -p tcp -j DROP",
		"-A 1PANEL_BASIC_AFTER -p udp -j DROP",
	} {
		if !strings.Contains(script, line+"\n") {
			t.Fatalf("batch script is missing %q:\n%s", line, script)
		}
	}
	if got := strings.Count(script, "--dport 22 -j ACCEPT"); got != 1 {
		t.Fatalf("duplicate desired port was not removed exactly once: count=%d\n%s", got, script)
	}
	if !strings.HasPrefix(script, "*filter\n") || !strings.HasSuffix(script, "COMMIT\n") {
		t.Fatalf("invalid restore transaction:\n%s", script)
	}
}

func TestBuildRequiredPortsRestoreScriptSkipsUnchangedState(t *testing.T) {
	desired := []firewall.PortWhitelist{{Protocol: "tcp", Port: "22"}}
	before := []FilterRules{{Chain: BasicBeforeChain, Protocol: "tcp", DstPort: "22", Strategy: "accept"}}
	if script := buildRequiredPortsRestoreScript(desired, before, nil, "", "", false); script != "" {
		t.Fatalf("unchanged required ports generated a restore transaction:\n%s", script)
	}
}

func TestBuildBaseChainBindingsRestoreScriptRebindsInOneTransaction(t *testing.T) {
	output := strings.Join([]string{
		"-A INPUT -j " + BasicBeforeChain,
		"-A INPUT -j external",
		"-A INPUT -j " + BasicAfterChain,
	}, "\n")
	script := buildBaseChainBindingsRestoreScript(output, true)
	for _, line := range []string{
		"-D INPUT -j " + BasicBeforeChain,
		"-D INPUT -j " + BasicAfterChain,
		"-I INPUT 1 -j " + BasicBeforeChain,
		"-I INPUT 2 -j " + BasicChain,
		"-I INPUT 3 -j " + BasicAfterChain,
	} {
		if !strings.Contains(script, line+"\n") {
			t.Fatalf("binding batch is missing %q:\n%s", line, script)
		}
	}
	if strings.Contains(script, "external") || strings.Count(script, "COMMIT\n") != 1 {
		t.Fatalf("binding batch modified an external rule or is not atomic:\n%s", script)
	}
}
