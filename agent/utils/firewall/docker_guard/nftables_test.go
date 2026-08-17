package docker_guard

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type nftRecordingRunner struct {
	objects    map[string]bool
	baseRules  map[string]string
	runCalls   [][]string
	inputCalls []restoreCall
}

func newNftRecordingRunner() *nftRecordingRunner {
	return &nftRecordingRunner{objects: map[string]bool{}, baseRules: map[string]string{}}
}

func (r *nftRecordingRunner) Run(executable string, args ...string) (string, error) {
	r.runCalls = append(r.runCalls, append([]string{executable}, args...))
	if executable != "nft" {
		return "", errors.New("unexpected executable")
	}
	plain := args
	if len(plain) > 0 && plain[0] == "-a" {
		plain = plain[1:]
	}
	if len(plain) >= 3 && plain[0] == "list" {
		key := strings.Join(plain[1:], "|")
		if !r.objects[key] {
			return "", errors.New("not found")
		}
		if plain[1] == "chain" && len(plain) == 5 && plain[4] == NftBaseChain {
			return r.baseRules[plain[2]], nil
		}
		return "", nil
	}
	if len(plain) >= 4 && plain[0] == "add" && (plain[1] == "table" || plain[1] == "chain") {
		nameLen := 3
		if plain[1] == "chain" {
			nameLen = 4
		}
		r.objects[strings.Join(plain[1:1+nameLen], "|")] = true
		return "", nil
	}
	if len(plain) >= 7 && plain[0] == "insert" && plain[1] == "rule" {
		r.baseRules[plain[2]] = "jump " + NftChain + " # handle 1\n"
		return "", nil
	}
	return "", nil
}

func (r *nftRecordingRunner) RunInput(executable, input string, args ...string) (string, error) {
	r.inputCalls = append(r.inputCalls, restoreCall{executable: executable, input: input, args: args})
	return "", nil
}

func (r *nftRecordingRunner) Exists(executable string) bool { return executable == "nft" }

func (r *nftRecordingRunner) addTable(family, table string) {
	r.objects["table|"+family+"|"+table] = true
}

func (r *nftRecordingRunner) addChain(family, table, chain string) {
	r.objects["chain|"+family+"|"+table+"|"+chain] = true
}

func TestCompileNftPolicyUsesOriginalDestination(t *testing.T) {
	rules := compileNftPolicy(Policy{UUID: "id", Family: FamilyIPv4, HostIP: "192.0.2.1", HostPort: 8080, Protocol: "tcp", Mode: ModeSources, Sources: []string{"203.0.113.0/24"}})
	got := strings.Join(rules[0], " ")
	for _, want := range []string{"meta l4proto tcp", "ct original ip daddr 192.0.2.1", "ct original proto-dst 8080", "ip saddr 203.0.113.0/24", "drop"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compiled rule %q does not contain %q", got, want)
		}
	}
}

func TestCompileNftWildcardDoesNotMatchWildcardAddress(t *testing.T) {
	rules := compileNftPolicy(Policy{UUID: "id", Family: FamilyIPv6, HostIP: "::", HostPort: 53, Protocol: "udp", Mode: ModeAll})
	got := strings.Join(rules[0], " ")
	if strings.Contains(got, "ct original ip6 daddr") {
		t.Fatalf("wildcard binding must not compile an original destination address: %s", got)
	}
	if !strings.Contains(got, "ct original proto-dst 53") {
		t.Fatalf("original destination port missing: %s", got)
	}
}

func TestCompileNftEmptyAllowSourcesDropsAll(t *testing.T) {
	rules := compileNftPolicy(Policy{UUID: "id", Family: FamilyIPv4, HostIP: "0.0.0.0", HostPort: 5432, Protocol: "tcp", Mode: ModeAllow})
	if len(rules) != 1 || !strings.HasSuffix(strings.Join(rules[0], " "), "drop") {
		t.Fatalf("rules = %#v", rules)
	}
}

func TestNftInitializeCreatesOwnedChainsBeforeDockerForwardRules(t *testing.T) {
	runner := newNftRecordingRunner()
	runner.addTable("ip", dockerNftTable)
	manager := NewNftablesManagerWithRunner(runner)
	if err := manager.Initialize(nil); err != nil {
		t.Fatal(err)
	}
	joined := make([]string, 0, len(runner.runCalls))
	for _, call := range runner.runCalls {
		joined = append(joined, strings.Join(call, " "))
	}
	all := strings.Join(joined, "\n")
	for _, want := range []string{
		"nft add table ip " + NftTable,
		"nft add chain ip " + NftTable + " " + NftBaseChain + " { type filter hook forward priority filter - 1 ; policy accept ; }",
		"nft add chain ip " + NftTable + " " + NftChain,
		"nft insert rule ip " + NftTable + " " + NftBaseChain + " jump " + NftChain,
	} {
		if !strings.Contains(all, want) {
			t.Fatalf("nft calls do not contain %q:\n%s", want, all)
		}
	}
	if len(runner.inputCalls) != 1 {
		t.Fatalf("restore calls = %d, want IPv4 only", len(runner.inputCalls))
	}
}

func TestNftReconcileUsesSingleAtomicScriptPerFamily(t *testing.T) {
	runner := newNftRecordingRunner()
	runner.addChain("ip", NftTable, NftChain)
	manager := NewNftablesManagerWithRunner(runner)
	policies := []Policy{
		{UUID: "first", Family: FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080, Protocol: "tcp", Mode: ModeAll},
		{UUID: "second", Family: FamilyIPv4, HostIP: "192.0.2.10", HostPort: 5432, Protocol: "tcp", Mode: ModeAllow, Sources: []string{"203.0.113.1/32"}},
	}
	if err := manager.Reconcile(policies); err != nil {
		t.Fatal(err)
	}
	if len(runner.inputCalls) != 1 {
		t.Fatalf("restore calls = %d, want 1", len(runner.inputCalls))
	}
	call := runner.inputCalls[0]
	if call.executable != "nft" || !reflect.DeepEqual(call.args, []string{"-f", "-"}) {
		t.Fatalf("restore call = %#v", call)
	}
	for _, want := range []string{
		"flush chain ip " + NftTable + " " + NftChain,
		"ct state { established,related } return",
		"ct original proto-dst 8080 comment \"1panel-docker:first\" drop",
		"ct original ip daddr 192.0.2.10 ct original proto-dst 5432 ip saddr 203.0.113.1/32",
		"add rule ip " + NftTable + " " + NftChain + " return",
	} {
		if !strings.Contains(call.input, want) {
			t.Fatalf("restore input does not contain %q:\n%s", want, call.input)
		}
	}
}

func TestNftFamilyStatusExplainsIncompleteStep(t *testing.T) {
	runner := newNftRecordingRunner()
	runner.addTable("ip", dockerNftTable)
	runner.addTable("ip", NftTable)
	runner.addChain("ip", NftTable, NftBaseChain)
	runner.addChain("ip", NftTable, NftChain)
	runner.baseRules["ip"] = "counter packets 0 bytes 0 # handle 1\njump " + NftChain + " # handle 2\n"
	status := NewNftablesManagerWithRunner(runner).Status(FamilyIPv4)
	if status.State != StatusNotEffective || status.Reason != ReasonJumpNotFirst || !status.Initialized {
		t.Fatalf("status = %#v", status)
	}
	runner.baseRules["ip"] = "jump " + NftChain + " # handle 2\n"
	status = NewNftablesManagerWithRunner(runner).Status(FamilyIPv4)
	if status.State != StatusEffective || !status.Initialized || !status.Bound || !status.Effective {
		t.Fatalf("status = %#v", status)
	}
}

func TestBuildNftScriptRejectsUnsafeTokens(t *testing.T) {
	if _, err := buildNftScript([][]string{{"add", "rule", "unsafe value"}}); err == nil {
		t.Fatal("expected unsafe token to be rejected")
	}
}
