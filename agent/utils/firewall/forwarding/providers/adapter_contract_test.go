package providers

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
)

type commandCall struct {
	name string
	args []string
}

type fakeFirewalldRunner struct {
	calls  []commandCall
	stdout map[string]string
	errors map[string]error
}

func (f *fakeFirewalldRunner) Run(name string, args ...string) error {
	f.calls = append(f.calls, commandCall{name: name, args: append([]string(nil), args...)})
	return f.errors[commandKey(name, args...)]
}

func (f *fakeFirewalldRunner) RunWithStdout(name string, args ...string) (string, error) {
	f.calls = append(f.calls, commandCall{name: name, args: append([]string(nil), args...)})
	key := commandKey(name, args...)
	return f.stdout[key], f.errors[key]
}

func commandKey(name string, args ...string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

func TestForwardingAdapterFactoryContract(t *testing.T) {
	for _, name := range []string{"firewalld", "ufw", "iptables", "nftables"} {
		adapter, err := New(name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if adapter.Name() != name {
			t.Fatalf("got adapter %q want %q", adapter.Name(), name)
		}
		if name == "firewalld" {
			if _, ok := adapter.(*firewalldAdapter); !ok {
				t.Fatalf("firewalld must use its native forwarding adapter, got %T", adapter)
			}
		} else if name == "nftables" {
			if _, ok := adapter.(*nftablesAdapter); !ok {
				t.Fatalf("nftables must use its native forwarding adapter, got %T", adapter)
			}
		} else if _, ok := adapter.(*iptablesNATAdapter); !ok {
			t.Fatalf("%s must use the iptables forwarding adapter, got %T", name, adapter)
		}
	}
	if _, err := New("unknown"); err == nil {
		t.Fatal("unknown forwarding provider must be rejected")
	}
}

func TestFirewalldForwardingContract(t *testing.T) {
	runner := &fakeFirewalldRunner{stdout: map[string]string{
		"firewall-cmd --zone=public --query-masquerade":   "yes\n",
		"firewall-cmd --zone=public --list-forward-ports": "port=8080:proto=tcp:toport=80:toaddr=10.0.0.2\nport=8443:proto=tcp:toport=443:toaddr=\ninvalid\n",
	}, errors: map[string]error{}}
	adapter := &firewalldAdapter{runner: runner}
	rules, err := adapter.List()
	if err != nil {
		t.Fatal(err)
	}
	wantRules := []forwarding.Rule{
		{Port: "8080", Protocol: "tcp", TargetIP: "10.0.0.2", TargetPort: "80"},
		{Port: "8443", Protocol: "tcp", TargetIP: "127.0.0.1", TargetPort: "443"},
	}
	if !reflect.DeepEqual(rules, wantRules) {
		t.Fatalf("got %#v want %#v", rules, wantRules)
	}

	remote := forwarding.Rule{Port: "8080", Protocol: "tcp", TargetIP: "10.0.0.2", TargetPort: "80"}
	if err := adapter.Operate(remote, "add"); err != nil {
		t.Fatal(err)
	}
	wantArgs := []string{"--zone=public", "--add-forward-port=port=8080:proto=tcp:toaddr=10.0.0.2:toport=80", "--permanent"}
	assertArgs(t, runner.calls[2], "firewall-cmd", wantArgs)
	assertArgs(t, runner.calls[3], "firewall-cmd", []string{"--reload"})

	localArgs := buildFirewalldForwardArgs(forwarding.Rule{Port: "8443", Protocol: "tcp", TargetIP: "127.0.0.1", TargetPort: "443"}, "remove")
	wantLocal := []string{"--zone=public", "--remove-forward-port=port=8443:proto=tcp:toport=443", "--permanent"}
	if !reflect.DeepEqual(localArgs, wantLocal) {
		t.Fatalf("got %#v want %#v", localArgs, wantLocal)
	}
}

func TestFirewalldEnableMasqueradeContract(t *testing.T) {
	query := "firewall-cmd --zone=public --query-masquerade"
	runner := &fakeFirewalldRunner{
		stdout: map[string]string{query: "no\n"},
		errors: map[string]error{query: errors.New("exit status 1")},
	}
	adapter := &firewalldAdapter{runner: runner}
	if err := adapter.Enable(); err != nil {
		t.Fatal(err)
	}
	want := []commandCall{
		{name: "firewall-cmd", args: []string{"--zone=public", "--query-masquerade"}},
		{name: "firewall-cmd", args: []string{"--zone=public", "--add-masquerade", "--permanent"}},
		{name: "firewall-cmd", args: []string{"--reload"}},
	}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("got %#v want %#v", runner.calls, want)
	}
}

type backendCall struct {
	method string
	table  string
	args   []string
}

type fakeIptablesBackend struct {
	calls   []backendCall
	stdout  map[string]string
	err     error
	saveErr error
}

func (f *fakeIptablesBackend) Run(table string, args ...string) error {
	f.calls = append(f.calls, backendCall{method: "run", table: table, args: append([]string(nil), args...)})
	return f.err
}

func (f *fakeIptablesBackend) RunWithStd(table string, args ...string) (string, error) {
	f.calls = append(f.calls, backendCall{method: "stdout", table: table, args: append([]string(nil), args...)})
	return f.stdout[commandKey(table, args...)], f.err
}

func (f *fakeIptablesBackend) AddChainWithAppend(table, parentChain, chain string) error {
	f.calls = append(f.calls, backendCall{method: "add-chain", table: table, args: []string{parentChain, chain}})
	return f.err
}

func (f *fakeIptablesBackend) SaveRulesToFile(table, chain, fileName string) error {
	f.calls = append(f.calls, backendCall{method: "save", table: table, args: []string{chain, fileName}})
	if f.saveErr != nil {
		return f.saveErr
	}
	return f.err
}

func (f *fakeIptablesBackend) LoadRulesFromFile(table, chain, fileName string) error {
	f.calls = append(f.calls, backendCall{method: "load", table: table, args: []string{chain, fileName}})
	return f.err
}

func TestIptablesOperateReturnsPersistenceError(t *testing.T) {
	wantErr := errors.New("save failed")
	backend := &fakeIptablesBackend{stdout: map[string]string{}, saveErr: wantErr}
	adapter := &iptablesNATAdapter{provider: "iptables", backend: backend, system: &fakeForwardingSystem{}}
	err := adapter.Operate(forwarding.Rule{Protocol: "tcp", Port: "8080", TargetPort: "80"}, forwarding.OperationAdd)
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}
}

type fileWrite struct {
	name string
	data string
}

type fakeForwardingSystem struct {
	reads  map[string][]byte
	writes []fileWrite
	runs   []commandCall
}

func (f *fakeForwardingSystem) ReadFile(name string) ([]byte, error) {
	data, ok := f.reads[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f *fakeForwardingSystem) WriteFile(name string, data []byte, _ os.FileMode) error {
	f.writes = append(f.writes, fileWrite{name: name, data: string(data)})
	return nil
}

func (f *fakeForwardingSystem) RunWithOptionalSudo(name string, args ...string) error {
	f.runs = append(f.runs, commandCall{name: name, args: append([]string(nil), args...)})
	return nil
}

func TestIptablesNATAddDeleteAndPersistenceContract(t *testing.T) {
	backend := &fakeIptablesBackend{stdout: map[string]string{}}
	adapter := &iptablesNATAdapter{provider: "ufw", backend: backend, system: &fakeForwardingSystem{}}
	rule := forwarding.Rule{Num: "3", Protocol: "tcp", Port: "8080-8081", TargetIP: "10.0.0.2", TargetPort: "80-81", Interface: "eth0"}
	if err := adapter.Operate(rule, "add"); err != nil {
		t.Fatal(err)
	}
	wantAdd := []backendCall{
		{method: "run", table: iptables_helper.NatTab, args: []string{"-A", forwarding.ChainPreRouting, "-i", "eth0", "-p", "tcp", "--dport", "8080:8081", "-j", "DNAT", "--to-destination", "10.0.0.2:80-81"}},
		{method: "run", table: iptables_helper.NatTab, args: []string{"-A", forwarding.ChainPostRouting, "-d", "10.0.0.2", "-p", "tcp", "--dport", "80:81", "-j", "MASQUERADE"}},
		{method: "run", table: iptables_helper.FilterTab, args: []string{"-A", forwarding.ChainForward, "-d", "10.0.0.2", "-p", "tcp", "--dport", "80:81", "-j", "ACCEPT"}},
		{method: "run", table: iptables_helper.FilterTab, args: []string{"-A", forwarding.ChainForward, "-s", "10.0.0.2", "-p", "tcp", "--sport", "80:81", "-j", "ACCEPT"}},
		{method: "save", table: iptables_helper.FilterTab, args: []string{forwarding.ChainForward, forwarding.ForwardFile}},
		{method: "save", table: iptables_helper.NatTab, args: []string{forwarding.ChainPreRouting, forwarding.PreRoutingFile}},
		{method: "save", table: iptables_helper.NatTab, args: []string{forwarding.ChainPostRouting, forwarding.PostRoutingFile}},
	}
	if !reflect.DeepEqual(backend.calls, wantAdd) {
		t.Fatalf("add transcript changed\ngot  %#v\nwant %#v", backend.calls, wantAdd)
	}

	backend.calls = nil
	if err := adapter.Operate(rule, "remove"); err != nil {
		t.Fatal(err)
	}
	wantRemovePrefix := []backendCall{
		{method: "run", table: iptables_helper.NatTab, args: []string{"-D", forwarding.ChainPreRouting, "3"}},
		{method: "run", table: iptables_helper.NatTab, args: []string{"-D", forwarding.ChainPostRouting, "-d", "10.0.0.2", "-p", "tcp", "--dport", "80:81", "-j", "MASQUERADE"}},
		{method: "run", table: iptables_helper.FilterTab, args: []string{"-D", forwarding.ChainForward, "-d", "10.0.0.2", "-p", "tcp", "--dport", "80:81", "-j", "ACCEPT"}},
		{method: "run", table: iptables_helper.FilterTab, args: []string{"-D", forwarding.ChainForward, "-s", "10.0.0.2", "-p", "tcp", "--sport", "80:81", "-j", "ACCEPT"}},
	}
	if !reflect.DeepEqual(backend.calls[:4], wantRemovePrefix) {
		t.Fatalf("remove transcript changed\ngot  %#v\nwant %#v", backend.calls[:4], wantRemovePrefix)
	}
}

func TestIptablesNATEnableReplayAndStatusContract(t *testing.T) {
	natStatus := strings.Join([]string{
		"-N THIRD_PARTY_DNAT",
		"-A PREROUTING -j THIRD_PARTY_DNAT",
		"-N " + forwarding.ChainPreRouting,
		"-N " + forwarding.ChainPostRouting,
		"-A PREROUTING -j " + forwarding.ChainPreRouting,
		"-A POSTROUTING -j " + forwarding.ChainPostRouting,
	}, "\n")
	filterStatus := "-N " + forwarding.ChainForward + "\n-A FORWARD -j " + forwarding.ChainForward + "\n-A FORWARD -j DOCKER-USER\n"
	backend := &fakeIptablesBackend{stdout: map[string]string{
		"nat -S":    natStatus,
		"filter -S": filterStatus,
	}}
	system := &fakeForwardingSystem{reads: map[string][]byte{
		"/proc/sys/net/ipv4/ip_forward": []byte("1\n"),
		"/etc/sysctl.conf":              []byte("net.ipv4.tcp_syncookies = 1\n"),
	}}
	adapter := &iptablesNATAdapter{provider: "iptables", backend: backend, system: system}
	if err := adapter.Enable(); err != nil {
		t.Fatal(err)
	}
	if len(system.writes) != 2 || system.writes[0].name != "/proc/sys/net/ipv4/ip_forward" ||
		!strings.Contains(system.writes[1].data, "net.ipv4.ip_forward = 1") {
		t.Fatalf("sysctl writes changed: %#v", system.writes)
	}
	if !reflect.DeepEqual(system.runs, []commandCall{{name: "sysctl", args: []string{"-p"}}}) {
		t.Fatalf("sysctl transcript changed: %#v", system.runs)
	}
	init, bind, err := adapter.InitStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !init || !bind {
		t.Fatalf("expected initialized and bound, got %v %v", init, bind)
	}
	backend.calls = nil
	if err := adapter.Replay(); err != nil {
		t.Fatal(err)
	}
	wantLoads := []backendCall{
		{method: "load", table: iptables_helper.FilterTab, args: []string{forwarding.ChainForward, forwarding.ForwardFile}},
		{method: "load", table: iptables_helper.NatTab, args: []string{forwarding.ChainPreRouting, forwarding.PreRoutingFile}},
		{method: "load", table: iptables_helper.NatTab, args: []string{forwarding.ChainPostRouting, forwarding.PostRoutingFile}},
	}
	if !reflect.DeepEqual(backend.calls, wantLoads) {
		t.Fatalf("replay transcript changed: %#v", backend.calls)
	}
}

func TestIptablesListParsingContract(t *testing.T) {
	stdout := strings.Join([]string{
		"1 0 0 DNAT tcp -- eth0 * 0.0.0.0/0 0.0.0.0/0 tcp dpt:8080 to:10.0.0.2:80",
		"2 0 0 REDIRECT udp -- * * 0.0.0.0/0 0.0.0.0/0 udp dpts:9000:9001 redir ports 53",
	}, "\n")
	rules := parseIptablesRules(stdout)
	want := []forwarding.Rule{
		{Num: "1", Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80", Interface: "eth0"},
		{Num: "2", Protocol: "udp", Port: "9000-9001", TargetIP: "127.0.0.1", TargetPort: "53", Interface: "*"},
	}
	if !reflect.DeepEqual(rules, want) {
		t.Fatalf("got %#v want %#v", rules, want)
	}
}

func TestEnableIPv4ForwardingReplacesDisabledSetting(t *testing.T) {
	content := strings.Join([]string{
		"# net.ipv4.ip_forward = 0",
		"net.ipv4.ip_forward=0",
		"net.ipv4.tcp_syncookies = 1",
	}, "\n")
	want := strings.Join([]string{
		"# net.ipv4.ip_forward = 0",
		"net.ipv4.ip_forward = 1",
		"net.ipv4.tcp_syncookies = 1",
		"",
	}, "\n")
	if got := enableIPv4Forwarding(content); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func assertArgs(t *testing.T, call commandCall, name string, args []string) {
	t.Helper()
	if call.name != name || !reflect.DeepEqual(call.args, args) {
		t.Fatalf("got %#v want %s %#v", call, name, args)
	}
}
