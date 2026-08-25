package docker_guard

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type restoreCall struct {
	executable string
	input      string
	args       []string
}

type recordingRunner struct {
	restoreCalls []restoreCall
	exists       map[string]bool
	chains       string
	dockerRules  string
	guardRules   string
	runErr       error
}

func (r *recordingRunner) Run(_ string, args ...string) (string, error) {
	if r.runErr != nil {
		return "", r.runErr
	}
	if reflect.DeepEqual(args, []string{"-w", "-t", "filter", "-S"}) {
		if r.chains != "" {
			return r.chains, nil
		}
		return "-N DOCKER-USER\n-N 1PANEL_DOCKER\n", nil
	}
	if reflect.DeepEqual(args, []string{"-w", "-t", "filter", "-S", DockerChain}) {
		if r.dockerRules != "" {
			return r.dockerRules, nil
		}
		return "-A DOCKER-USER -j 1PANEL_DOCKER\n", nil
	}
	if reflect.DeepEqual(args, []string{"-w", "-t", "filter", "-S", Chain}) {
		return r.guardRules, nil
	}
	return "", nil
}

func (r *recordingRunner) RunInput(executable, input string, args ...string) (string, error) {
	r.restoreCalls = append(r.restoreCalls, restoreCall{executable: executable, input: input, args: args})
	return "", nil
}

func (r *recordingRunner) Exists(executable string) bool {
	if r.exists != nil {
		return r.exists[executable]
	}
	return executable == "iptables" || executable == "iptables-restore"
}

func TestCompilePolicyUsesOriginalDestination(t *testing.T) {
	rules := compilePolicy(Policy{UUID: "id", Family: FamilyIPv4, HostIP: "192.0.2.1", HostPort: 8080, Protocol: "tcp", Mode: ModeSources, Sources: []string{"203.0.113.0/24"}})
	got := strings.Join(rules[0], " ")
	for _, want := range []string{"--ctorigdst 192.0.2.1", "--ctorigdstport 8080", "-s 203.0.113.0/24", "-j DROP"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compiled rule %q does not contain %q", got, want)
		}
	}
}

func TestCompileWildcardDoesNotMatchUnroutableWildcardAddress(t *testing.T) {
	rules := compilePolicy(Policy{UUID: "id", Family: FamilyIPv4, HostIP: "0.0.0.0", HostPort: 53, Protocol: "udp", Mode: ModeAll})
	got := strings.Join(rules[0], " ")
	if strings.Contains(got, "--ctorigdst ") {
		t.Fatalf("wildcard binding must not compile an original destination address: %s", got)
	}
	if !strings.Contains(got, "--ctorigdstport 53") {
		t.Fatalf("original destination port missing: %s", got)
	}
}

func TestCompileAllowSourcesReturnsAllowedAndDropsOthers(t *testing.T) {
	rules := compilePolicy(Policy{UUID: "id", Family: FamilyIPv4, HostIP: "0.0.0.0", HostPort: 5432, Protocol: "tcp", Mode: ModeAllow, Sources: []string{"203.0.113.10/32", "192.0.2.0/24"}})
	if len(rules) != 3 {
		t.Fatalf("compiled %d rules, want 3", len(rules))
	}
	for i, source := range []string{"203.0.113.10/32", "192.0.2.0/24"} {
		got := strings.Join(rules[i], " ")
		if !strings.Contains(got, "-s "+source) || !strings.HasSuffix(got, "-j RETURN") {
			t.Fatalf("allow rule = %q", got)
		}
	}
	if got := strings.Join(rules[2], " "); strings.Contains(got, " -s ") || !strings.HasSuffix(got, "-j DROP") {
		t.Fatalf("fallback rule = %q", got)
	}
}

func TestCompileEmptyAllowSourcesDropsAll(t *testing.T) {
	rules := compilePolicy(Policy{UUID: "id", Family: FamilyIPv4, HostIP: "0.0.0.0", HostPort: 5432, Protocol: "tcp", Mode: ModeAllow})
	if len(rules) != 1 || !strings.HasSuffix(strings.Join(rules[0], " "), "-j DROP") {
		t.Fatalf("rules = %#v", rules)
	}
}

func TestParseIptablesDockerGuardPolicies(t *testing.T) {
	output := strings.Join([]string{
		`-A 1PANEL_DOCKER -p tcp -m conntrack --ctorigdstport 8080 -m comment --comment "1panel-docker:deny" -j DROP`,
		`-A 1PANEL_DOCKER -p tcp -m conntrack --ctorigdst 192.0.2.10/32 --ctorigdstport 5432 -s 203.0.113.1/32 -m comment --comment "1panel-docker:allow" -j RETURN`,
		`-A 1PANEL_DOCKER -p tcp -m conntrack --ctorigdst 192.0.2.10/32 --ctorigdstport 5432 -m comment --comment "1panel-docker:allow" -j DROP`,
	}, "\n")
	policies, err := parseDockerGuardPolicies(output, FamilyIPv4)
	if err != nil {
		t.Fatal(err)
	}
	if len(policies) != 2 || policies[0].UUID != "deny" || policies[0].Mode != ModeAll ||
		policies[1].UUID != "allow" || policies[1].HostIP != "192.0.2.10" || policies[1].Mode != ModeAllow ||
		!reflect.DeepEqual(policies[1].Sources, []string{"203.0.113.1/32"}) {
		t.Fatalf("policies = %#v", policies)
	}
}

func TestEffectiveJumpMustBeFirstAndUnique(t *testing.T) {
	if !hasFirstUniqueJump("-A DOCKER-USER -j 1PANEL_DOCKER\n-A DOCKER-USER -j RETURN\n") {
		t.Fatal("expected first unique jump to be effective")
	}
	if hasFirstUniqueJump("-A DOCKER-USER -j OTHER\n-A DOCKER-USER -j 1PANEL_DOCKER\n") {
		t.Fatal("jump after another rule must not be reported effective")
	}
	if hasFirstUniqueJump("-A DOCKER-USER -j 1PANEL_DOCKER\n-A DOCKER-USER -j 1PANEL_DOCKER\n") {
		t.Fatal("duplicate jumps must not be reported effective")
	}
}

func TestReconcileUsesSingleAtomicRestorePerFamily(t *testing.T) {
	runner := &recordingRunner{}
	manager := NewManagerWithRunner(runner)
	policies := []Policy{
		{UUID: "first", Family: FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080, Protocol: "tcp", Mode: ModeAll},
		{UUID: "second", Family: FamilyIPv4, HostIP: "192.0.2.10", HostPort: 5432, Protocol: "tcp", Mode: ModeAllow, Sources: []string{"203.0.113.1/32"}},
	}
	if err := manager.Reconcile(policies); err != nil {
		t.Fatal(err)
	}
	if len(runner.restoreCalls) != 1 {
		t.Fatalf("restore calls = %d, want 1", len(runner.restoreCalls))
	}
	call := runner.restoreCalls[0]
	if call.executable != "iptables-restore" || !reflect.DeepEqual(call.args, []string{"--noflush", "--wait"}) {
		t.Fatalf("restore call = %#v", call)
	}
	for _, want := range []string{
		"*filter\n",
		"-F 1PANEL_DOCKER\n",
		"-A 1PANEL_DOCKER -m conntrack --ctstate RELATED,ESTABLISHED -j RETURN\n",
		"--ctorigdstport 8080",
		"--ctorigdst 192.0.2.10 --ctorigdstport 5432",
		"-s 203.0.113.1/32",
		"-A 1PANEL_DOCKER -j RETURN\nCOMMIT\n",
	} {
		if !strings.Contains(call.input, want) {
			t.Fatalf("restore input does not contain %q:\n%s", want, call.input)
		}
	}
	if strings.Contains(call.input, "-F DOCKER-USER") {
		t.Fatalf("restore input must not flush Docker chains:\n%s", call.input)
	}
}

func TestDockerGuardLifecycleRulesBatchCreateAndRebind(t *testing.T) {
	output := strings.Join([]string{
		"-N " + DockerChain,
		"-A " + DockerChain + " -j " + Chain,
		"-A " + DockerChain + " -j " + Chain,
	}, "\n")
	rules := dockerGuardLifecycleRules(output, true, true)
	script, err := buildRestoreScript(rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{
		"-N " + Chain,
		"-D " + DockerChain + " -j " + Chain,
		"-I " + DockerChain + " 1 -j " + Chain,
	} {
		if !strings.Contains(script, line+"\n") {
			t.Fatalf("lifecycle restore is missing %q:\n%s", line, script)
		}
	}
	if strings.Count(script, "-D "+DockerChain+" -j "+Chain+"\n") != 2 || strings.Count(script, "COMMIT\n") != 1 {
		t.Fatalf("duplicate jumps were not removed in one transaction:\n%s", script)
	}
}

func TestReconcileReturnsChainInspectionError(t *testing.T) {
	manager := NewManagerWithRunner(&recordingRunner{runErr: errors.New("inspect failed")})
	err := manager.Reconcile(nil)
	if err == nil {
		t.Fatal("expected chain inspection error")
	}
	var familyErr *FamilyError
	if !errors.As(err, &familyErr) || familyErr.Family != FamilyIPv4 {
		t.Fatalf("error = %#v, want IPv4 FamilyError", err)
	}
}

func TestBuildRestoreScriptRejectsUnsafeTokens(t *testing.T) {
	if _, err := buildRestoreScript([][]string{{"-A", Chain, "--comment", "unsafe value"}}); err == nil {
		t.Fatal("expected unsafe token to be rejected")
	}
}

func TestFamilyStatusExplainsIncompleteStep(t *testing.T) {
	tests := []struct {
		name        string
		runner      *recordingRunner
		wantState   string
		wantReason  string
		initialized bool
		bound       bool
	}{
		{name: "command missing", runner: &recordingRunner{exists: map[string]bool{}}, wantState: StatusDisabled, wantReason: ReasonCommandMissing},
		{name: "Docker chain missing", runner: &recordingRunner{chains: "-N 1PANEL_DOCKER\n"}, wantState: StatusDisabled, wantReason: ReasonDockerChainMissing},
		{name: "guard chain missing", runner: &recordingRunner{chains: "-N DOCKER-USER\n"}, wantState: StatusDisabled, wantReason: ReasonGuardChainMissing},
		{name: "jump missing", runner: &recordingRunner{dockerRules: "-A DOCKER-USER -j RETURN\n"}, wantState: StatusNotEffective, wantReason: ReasonJumpMissing, initialized: true},
		{name: "jump not first", runner: &recordingRunner{dockerRules: "-A DOCKER-USER -j OTHER\n-A DOCKER-USER -j 1PANEL_DOCKER\n"}, wantState: StatusNotEffective, wantReason: ReasonJumpNotFirst, initialized: true},
		{name: "jump duplicate", runner: &recordingRunner{dockerRules: "-A DOCKER-USER -j 1PANEL_DOCKER\n-A DOCKER-USER -j 1PANEL_DOCKER\n"}, wantState: StatusNotEffective, wantReason: ReasonJumpDuplicate, initialized: true},
		{name: "effective", runner: &recordingRunner{}, wantState: StatusEffective, initialized: true, bound: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := NewManagerWithRunner(test.runner).Status(FamilyIPv4)
			if status.State != test.wantState || status.Reason != test.wantReason || status.Initialized != test.initialized || status.Bound != test.bound {
				t.Fatalf("status = %#v", status)
			}
		})
	}
}
