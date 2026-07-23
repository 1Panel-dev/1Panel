package service

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

// The rule flow is provider agnostic, so these sequences pair the real client
// expansion with a recorded apply step and show how record writes interleave.
// One intentional difference against dev-v2: ufw returned before Reload for port
// rules, the shared flow always reloads. Ufw.Reload is a no-op, so the recorded
// trailing "reload" has no effect on a real host.

type recordingFilterClient struct {
	firewall.FilterClient
	events *[]string
	err    error
}

func (r *recordingFilterClient) ApplyPortUnit(unit fireClient.PortUnit, operation string) error {
	*r.events = append(*r.events, fmt.Sprintf("apply-port %s apply(port=%s proto=%s addr=%s) record(port=%s proto=%s addr=%s) chain=%s",
		operation,
		unit.Apply.Port, unit.Apply.Protocol, unit.Apply.Address,
		unit.Record.Port, unit.Record.Protocol, unit.Record.Address,
		unit.Chain))
	return r.err
}

func (r *recordingFilterClient) ApplyAddressUnit(unit fireClient.AddressUnit, operation string) error {
	*r.events = append(*r.events, fmt.Sprintf("apply-address %s addr=%s strategy=%s chain=%s",
		operation, unit.Apply.Address, unit.Apply.Strategy, unit.Chain))
	return r.err
}

func (r *recordingFilterClient) Port(port fireClient.FireInfo, operation string) error {
	*r.events = append(*r.events, fmt.Sprintf("port %s port=%s proto=%s strategy=%s", operation, port.Port, port.Protocol, port.Strategy))
	return r.err
}

func (r *recordingFilterClient) Reload() error {
	*r.events = append(*r.events, "reload")
	return r.err
}

type recordingHostRepo struct {
	repo.IHostRepo
	events *[]string
}

func (r *recordingHostRepo) SaveFirewallRecord(record *model.Firewall) error {
	*r.events = append(*r.events, fmt.Sprintf("save-record %s chain=%s port=%s proto=%s addr=%s strategy=%s description=%s",
		record.Type, record.Chain, record.DstPort, record.Protocol, record.SrcIP, record.Strategy, record.Description))
	return nil
}

func (r *recordingHostRepo) DeleteFirewallRecordByID(id uint) error {
	*r.events = append(*r.events, fmt.Sprintf("delete-record id=%d", id))
	return nil
}

func newRecordingFlow(t *testing.T, name string) (*recordingFilterClient, *[]string) {
	t.Helper()
	var client firewall.FilterClient
	var err error
	switch name {
	case "ufw":
		client, err = fireClient.NewUfw()
	case "firewalld":
		client, err = fireClient.NewFirewalld()
	case "iptables":
		client, err = fireClient.NewIptables()
	default:
		t.Fatalf("unknown provider %s", name)
	}
	if err != nil {
		t.Fatal(err)
	}
	events := make([]string, 0)
	original := hostRepo
	hostRepo = &recordingHostRepo{events: &events}
	t.Cleanup(func() { hostRepo = original })
	return &recordingFilterClient{FilterClient: client, events: &events}, &events
}

func TestOperatePortRuleFlowGoldenSequence(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		req      dto.PortRuleOperate
		want     []string
	}{
		{
			name:     "ufw range applies colon syntax and records dash syntax",
			provider: "ufw",
			req:      dto.PortRuleOperate{Operation: "add", Port: "8000-8010", Protocol: "tcp", Strategy: "accept", Description: "web"},
			want: []string{
				"apply-port add apply(port=8000:8010 proto=tcp addr=) record(port=8000-8010 proto=tcp addr=Anywhere) chain=",
				"save-record port chain= port=8000-8010 proto=tcp addr=Anywhere strategy=accept description=web",
				"reload",
			},
		},
		{
			name:     "ufw dual protocol single port stays one rule",
			provider: "ufw",
			req:      dto.PortRuleOperate{Operation: "add", Port: "53", Protocol: "tcp/udp", Strategy: "accept", Description: "dns"},
			want: []string{
				"apply-port add apply(port=53 proto= addr=) record(port=53 proto=tcp/udp addr=Anywhere) chain=",
				"save-record port chain= port=53 proto=tcp/udp addr=Anywhere strategy=accept description=dns",
				"reload",
			},
		},
		{
			name:     "firewalld expands protocols then ports then sources",
			provider: "firewalld",
			req:      dto.PortRuleOperate{Operation: "add", Port: "80,443", Protocol: "tcp/udp", Address: "1.1.1.1,2.2.2.2", Strategy: "accept", Description: "web"},
			want: []string{
				"apply-port add apply(port=80 proto=tcp addr=1.1.1.1) record(port=80 proto=tcp addr=1.1.1.1) chain=",
				"save-record port chain= port=80 proto=tcp addr=1.1.1.1 strategy=accept description=web",
				"apply-port add apply(port=80 proto=tcp addr=2.2.2.2) record(port=80 proto=tcp addr=2.2.2.2) chain=",
				"save-record port chain= port=80 proto=tcp addr=2.2.2.2 strategy=accept description=web",
				"apply-port add apply(port=443 proto=tcp addr=1.1.1.1) record(port=443 proto=tcp addr=1.1.1.1) chain=",
				"save-record port chain= port=443 proto=tcp addr=1.1.1.1 strategy=accept description=web",
				"apply-port add apply(port=443 proto=tcp addr=2.2.2.2) record(port=443 proto=tcp addr=2.2.2.2) chain=",
				"save-record port chain= port=443 proto=tcp addr=2.2.2.2 strategy=accept description=web",
				"apply-port add apply(port=80 proto=udp addr=1.1.1.1) record(port=80 proto=udp addr=1.1.1.1) chain=",
				"save-record port chain= port=80 proto=udp addr=1.1.1.1 strategy=accept description=web",
				"apply-port add apply(port=80 proto=udp addr=2.2.2.2) record(port=80 proto=udp addr=2.2.2.2) chain=",
				"save-record port chain= port=80 proto=udp addr=2.2.2.2 strategy=accept description=web",
				"apply-port add apply(port=443 proto=udp addr=1.1.1.1) record(port=443 proto=udp addr=1.1.1.1) chain=",
				"save-record port chain= port=443 proto=udp addr=1.1.1.1 strategy=accept description=web",
				"apply-port add apply(port=443 proto=udp addr=2.2.2.2) record(port=443 proto=udp addr=2.2.2.2) chain=",
				"save-record port chain= port=443 proto=udp addr=2.2.2.2 strategy=accept description=web",
				"reload",
			},
		},
		{
			name:     "iptables records the owning chain",
			provider: "iptables",
			req:      dto.PortRuleOperate{Operation: "add", Port: "8000-8010", Protocol: "tcp", Strategy: "accept", Description: "web"},
			want: []string{
				"apply-port add apply(port=8000-8010 proto=tcp addr=) record(port=8000-8010 proto=tcp addr=) chain=1PANEL_BASIC",
				"save-record port chain=1PANEL_BASIC port=8000-8010 proto=tcp addr= strategy=accept description=web",
				"reload",
			},
		},
		{
			name:     "iptables keeps an explicit chain",
			provider: "iptables",
			req:      dto.PortRuleOperate{Operation: "add", Chain: iptables.Chain1PanelBasicBefore, Port: "80", Protocol: "tcp", Strategy: "accept", Description: "panel"},
			want: []string{
				"apply-port add apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=1PANEL_BASIC_BEFORE",
				"save-record port chain=1PANEL_BASIC_BEFORE port=80 proto=tcp addr= strategy=accept description=panel",
				"reload",
			},
		},
		{
			name:     "a rule without description writes no record",
			provider: "iptables",
			req:      dto.PortRuleOperate{Operation: "add", Port: "80", Protocol: "tcp", Strategy: "accept"},
			want: []string{
				"apply-port add apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=1PANEL_BASIC",
				"reload",
			},
		},
		{
			name:     "removing a rule deletes its record once per unit",
			provider: "firewalld",
			req:      dto.PortRuleOperate{ID: 7, Operation: "remove", Port: "80,443", Protocol: "tcp", Strategy: "accept"},
			want: []string{
				"apply-port remove apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=) chain=",
				"delete-record id=7",
				"apply-port remove apply(port=443 proto=tcp addr=) record(port=443 proto=tcp addr=) chain=",
				"delete-record id=7",
				"reload",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, events := newRecordingFlow(t, tt.provider)
			if err := (&FirewallService{}).operatePortRuleWithClient(client, tt.req, true); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*events, tt.want) {
				t.Fatalf("flow changed\ngot  %#v\nwant %#v", *events, tt.want)
			}
		})
	}
}

func TestOperateAddressRuleFlowGoldenSequence(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		req      dto.AddrRuleOperate
		want     []string
	}{
		{
			name:     "external providers own no chain",
			provider: "ufw",
			req:      dto.AddrRuleOperate{Operation: "add", Address: "1.1.1.1,2.2.2.2", Strategy: "drop", Description: "block"},
			want: []string{
				"apply-address add addr=1.1.1.1 strategy=drop chain=",
				"save-record address chain= port= proto= addr=1.1.1.1 strategy=drop description=block",
				"apply-address add addr=2.2.2.2 strategy=drop chain=",
				"save-record address chain= port= proto= addr=2.2.2.2 strategy=drop description=block",
				"reload",
			},
		},
		{
			name:     "iptables records the basic chain",
			provider: "iptables",
			req:      dto.AddrRuleOperate{Operation: "add", Address: "10.0.0.1", Strategy: "drop", Description: "block"},
			want: []string{
				"apply-address add addr=10.0.0.1 strategy=drop chain=1PANEL_BASIC",
				"save-record address chain=1PANEL_BASIC port= proto= addr=10.0.0.1 strategy=drop description=block",
				"reload",
			},
		},
		{
			name:     "empty entries are dropped",
			provider: "firewalld",
			req:      dto.AddrRuleOperate{ID: 3, Operation: "remove", Address: "1.1.1.1,,2.2.2.2", Strategy: "drop"},
			want: []string{
				"apply-address remove addr=1.1.1.1 strategy=drop chain=",
				"delete-record id=3",
				"apply-address remove addr=2.2.2.2 strategy=drop chain=",
				"delete-record id=3",
				"reload",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, events := newRecordingFlow(t, tt.provider)
			if err := (&FirewallService{}).operateAddressRuleWithClient(client, tt.req, true); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*events, tt.want) {
				t.Fatalf("flow changed\ngot  %#v\nwant %#v", *events, tt.want)
			}
		})
	}
}

func TestOperateFirewallPortsGoldenSequence(t *testing.T) {
	want := []string{
		"port add port=2222 proto=tcp strategy=accept",
		"port remove port=22 proto=tcp strategy=accept",
		"reload",
	}
	for _, name := range []string{"ufw", "firewalld", "iptables"} {
		t.Run(name, func(t *testing.T) {
			client, events := newRecordingFlow(t, name)
			if err := operateFirewallPorts(client, []int{22}, []int{2222}); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*events, want) {
				t.Fatalf("got %#v want %#v", *events, want)
			}
		})
	}
}

func TestRuleFlowReturnsClientErrorWithoutFallback(t *testing.T) {
	nativeErr := errors.New("native firewall failure")
	client, events := newRecordingFlow(t, "ufw")
	client.err = nativeErr
	err := (&FirewallService{}).operatePortRuleWithClient(client, dto.PortRuleOperate{
		Operation:   "add",
		Port:        "80",
		Protocol:    "tcp",
		Strategy:    "accept",
		Description: "web",
	}, true)
	if !errors.Is(err, nativeErr) {
		t.Fatalf("flow must return the client error, got %v", err)
	}
	want := []string{"apply-port add apply(port=80 proto=tcp addr=) record(port=80 proto=tcp addr=Anywhere) chain="}
	if !reflect.DeepEqual(*events, want) {
		t.Fatalf("a failed apply must not write a record: %#v", *events)
	}
}

func TestLoadInitStatusExternalParity(t *testing.T) {
	for _, tt := range []struct {
		name string
		tab  string
	}{
		{name: "ufw", tab: "base"},
		{name: "ufw", tab: "port"},
		{name: "firewalld", tab: "base"},
		{name: "firewalld", tab: "advance"},
	} {
		t.Run(tt.name+"/"+tt.tab, func(t *testing.T) {
			isInit, isBind := iptables.LoadInitStatus(tt.name, tt.tab)
			if !isInit || !isBind {
				t.Fatalf("expected dev-v2 true,true, got %v,%v", isInit, isBind)
			}
		})
	}
}
