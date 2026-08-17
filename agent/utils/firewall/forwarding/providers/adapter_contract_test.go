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

func commandKey(name string, args ...string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

func TestForwardingAdapterFactoryContract(t *testing.T) {
	for _, name := range []string{"iptables", "nftables"} {
		adapter, err := New(name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if adapter.Name() != name {
			t.Fatalf("got adapter %q want %q", adapter.Name(), name)
		}
		if name == "nftables" {
			if _, ok := adapter.(*nftablesAdapter); !ok {
				t.Fatalf("nftables must use its native forwarding adapter, got %T", adapter)
			}
		} else if _, ok := adapter.(*iptablesNATAdapter); !ok {
			t.Fatalf("%s must use the iptables forwarding adapter, got %T", name, adapter)
		}
	}
	for _, name := range []string{"firewalld", "ufw", "unknown"} {
		if _, err := New(name); err == nil {
			t.Fatalf("unsupported forwarding provider %q must be rejected", name)
		}
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
	ipv6    bool
}

func (f *fakeIptablesBackend) IPv6Available() bool { return f.ipv6 }

func (f *fakeIptablesBackend) Run(table string, args ...string) error {
	f.calls = append(f.calls, backendCall{method: "run", table: table, args: append([]string(nil), args...)})
	return f.err
}

func (f *fakeIptablesBackend) RunWithStd(table string, args ...string) (string, error) {
	f.calls = append(f.calls, backendCall{method: "stdout", table: table, args: append([]string(nil), args...)})
	return f.stdout[commandKey(table, args...)], f.err
}

func (f *fakeIptablesBackend) RunIPv6(table string, args ...string) error {
	f.calls = append(f.calls, backendCall{method: "run6", table: table, args: append([]string(nil), args...)})
	return f.err
}

func (f *fakeIptablesBackend) RunIPv6WithStd(table string, args ...string) (string, error) {
	f.calls = append(f.calls, backendCall{method: "stdout6", table: table, args: append([]string(nil), args...)})
	return f.stdout["ipv6 "+commandKey(table, args...)], f.err
}

func (f *fakeIptablesBackend) AddChainWithAppend(table, parentChain, chain string) error {
	f.calls = append(f.calls, backendCall{method: "add-chain", table: table, args: []string{parentChain, chain}})
	return f.err
}

func (f *fakeIptablesBackend) AddIPv6ChainWithAppend(table, parentChain, chain string) error {
	f.calls = append(f.calls, backendCall{method: "add-chain6", table: table, args: []string{parentChain, chain}})
	return f.err
}

func (f *fakeIptablesBackend) SaveRulesToFile(table, chain, fileName string) error {
	f.calls = append(f.calls, backendCall{method: "save", table: table, args: []string{chain, fileName}})
	if f.saveErr != nil {
		return f.saveErr
	}
	return f.err
}

func (f *fakeIptablesBackend) SaveIPv6RulesToFile(table, chain, fileName string) error {
	f.calls = append(f.calls, backendCall{method: "save6", table: table, args: []string{chain, fileName}})
	if f.saveErr != nil {
		return f.saveErr
	}
	return f.err
}

func (f *fakeIptablesBackend) LoadRulesFromFile(table, chain, fileName string) error {
	f.calls = append(f.calls, backendCall{method: "load", table: table, args: []string{chain, fileName}})
	return f.err
}

func (f *fakeIptablesBackend) LoadIPv6RulesFromFile(table, chain, fileName string) error {
	f.calls = append(f.calls, backendCall{method: "load6", table: table, args: []string{chain, fileName}})
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
	adapter := &iptablesNATAdapter{provider: "iptables", backend: backend, system: &fakeForwardingSystem{}}
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
	system.reads["/proc/sys/net/ipv4/ip_forward"] = []byte("0\n")
	init, bind, err = adapter.InitStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !init || bind {
		t.Fatalf("disabled IP forwarding must preserve initialization without reporting a binding, got %v %v", init, bind)
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
	rules := parseIptablesRules(stdout, forwarding.FamilyIPv4)
	want := []forwarding.Rule{
		{Num: "1", Family: forwarding.FamilyIPv4, Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80", Interface: "eth0"},
		{Num: "2", Family: forwarding.FamilyIPv4, Protocol: "udp", Port: "9000-9001", TargetIP: "127.0.0.1", TargetPort: "53", Interface: "*"},
	}
	if !reflect.DeepEqual(rules, want) {
		t.Fatalf("got %#v want %#v", rules, want)
	}
}

func TestIptablesIPv6NATContract(t *testing.T) {
	backend := &fakeIptablesBackend{stdout: map[string]string{}, ipv6: true}
	adapter := &iptablesNATAdapter{provider: "iptables", backend: backend, system: &fakeForwardingSystem{}}
	rule := forwarding.Rule{Family: forwarding.FamilyIPv6, Num: "2", Protocol: "tcp", Port: "8443", TargetIP: "2001:db8::20", TargetPort: "443", Interface: "eth0"}
	if err := adapter.Operate(rule, forwarding.OperationAdd); err != nil {
		t.Fatal(err)
	}
	wantPrefix := []backendCall{
		{method: "run6", table: iptables_helper.NatTab, args: []string{"-A", forwarding.ChainPreRouting, "-i", "eth0", "-p", "tcp", "--dport", "8443", "-j", "DNAT", "--to-destination", "[2001:db8::20]:443"}},
		{method: "run6", table: iptables_helper.NatTab, args: []string{"-A", forwarding.ChainPostRouting, "-d", "2001:db8::20", "-p", "tcp", "--dport", "443", "-j", "MASQUERADE"}},
		{method: "run6", table: iptables_helper.FilterTab, args: []string{"-A", forwarding.ChainForward, "-d", "2001:db8::20", "-p", "tcp", "--dport", "443", "-j", "ACCEPT"}},
		{method: "run6", table: iptables_helper.FilterTab, args: []string{"-A", forwarding.ChainForward, "-s", "2001:db8::20", "-p", "tcp", "--sport", "443", "-j", "ACCEPT"}},
	}
	if !reflect.DeepEqual(backend.calls[:4], wantPrefix) {
		t.Fatalf("IPv6 add transcript changed\ngot  %#v\nwant %#v", backend.calls[:4], wantPrefix)
	}

	stdout := "1 0 0 DNAT tcp -- eth0 * ::/0 ::/0 tcp dpt:8443 to:[2001:db8::20]:443"
	rules := parseIptablesRules(stdout, forwarding.FamilyIPv6)
	if len(rules) != 1 || rules[0].Family != forwarding.FamilyIPv6 || rules[0].TargetIP != "2001:db8::20" || rules[0].TargetPort != "443" {
		t.Fatalf("unexpected parsed IPv6 rules: %#v", rules)
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

func TestEnableForwardingSysctlsAddsIPv6(t *testing.T) {
	content := "net.ipv4.ip_forward = 0\nnet.ipv6.conf.all.forwarding=0\n"
	want := "net.ipv4.ip_forward = 1\nnet.ipv6.conf.all.forwarding = 1\n"
	if got := enableForwardingSysctls(content, true); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func assertArgs(t *testing.T, call commandCall, name string, args []string) {
	t.Helper()
	if call.name != name || !reflect.DeepEqual(call.args, args) {
		t.Fatalf("got %#v want %s %#v", call, name, args)
	}
}
