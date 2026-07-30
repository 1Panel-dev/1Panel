package service

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	forwardClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/go-playground/validator/v10"
)

type forwardingCall struct {
	rule      forwardClient.Rule
	operation string
}

type fakeForwardingAdapter struct {
	name       string
	rules      []forwardClient.Rule
	listErr    error
	operateErr error
	calls      []forwardingCall
}

func (f *fakeForwardingAdapter) Name() string { return f.name }

func (f *fakeForwardingAdapter) List() ([]forwardClient.Rule, error) {
	return append([]forwardClient.Rule(nil), f.rules...), f.listErr
}

func (f *fakeForwardingAdapter) Operate(rule forwardClient.Rule, operation string) error {
	f.calls = append(f.calls, forwardingCall{rule: rule, operation: operation})
	return f.operateErr
}

func (f *fakeForwardingAdapter) Enable() error            { return nil }
func (f *fakeForwardingAdapter) InitStatus() (bool, bool) { return true, true }
func (f *fakeForwardingAdapter) Replay() error            { return nil }

func forwardingServiceWithAdapter(adapter forwardClient.Adapter) *ForwardingService {
	return &ForwardingService{
		adapterFactory: func() (forwardClient.Adapter, error) { return adapter, nil },
		filterFactory:  firewall.NewFirewallClient,
	}
}

func TestForwardingAndFilterInterfacesAreSeparated(t *testing.T) {
	filterType := reflect.TypeOf((*firewall.FilterClient)(nil)).Elem()
	for _, method := range []string{"ListForward", "PortForward", "EnableForward"} {
		if _, ok := filterType.MethodByName(method); ok {
			t.Fatalf("filter interface still exposes %s", method)
		}
	}
	firewallServiceType := reflect.TypeOf((*IFirewallService)(nil)).Elem()
	if _, ok := firewallServiceType.MethodByName("OperateForwardRule"); ok {
		t.Fatal("firewall service still owns forwarding writes")
	}
	forwardingServiceType := reflect.TypeOf((*IForwardingService)(nil)).Elem()
	for _, method := range []string{"LoadBaseInfo", "SearchWithPage", "Operate", "Enable", "Replay"} {
		if _, ok := forwardingServiceType.MethodByName(method); !ok {
			t.Fatalf("forwarding service missing %s", method)
		}
	}
}

func TestForwardingInitRequestContract(t *testing.T) {
	req := dto.IptablesOp{Name: "1PANEL_FORWARD", Operate: "init-forward"}
	if err := validator.New().Struct(req); err != nil {
		t.Fatalf("frontend forwarding initialization request must remain valid: %v", err)
	}
}

func TestForwardingSearchPreservesAPIShapeAndPagination(t *testing.T) {
	adapter := &fakeForwardingAdapter{name: "iptables", rules: []forwardClient.Rule{
		{Num: "1", Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80", Interface: "eth0"},
		{Num: "2", Protocol: "udp", Port: "5353", TargetIP: "127.0.0.1", TargetPort: "53"},
	}}
	service := forwardingServiceWithAdapter(adapter)
	total, value, err := service.SearchWithPage(dto.ForwardRuleSearch{PageInfo: dto.PageInfo{Page: 1, PageSize: 10}, Info: "10.0.0.2"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("got total %d want 1", total)
	}
	items, ok := value.([]dto.ForwardRule)
	if !ok || len(items) != 1 || items[0].Port != "8080" {
		t.Fatalf("unexpected items: %#v", value)
	}
	data, err := json.Marshal(items[0])
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	wantFields := []string{"id", "chain", "family", "address", "port", "protocol", "strategy", "num", "targetIP", "targetPort", "interface", "usedStatus", "description"}
	for _, field := range wantFields {
		if _, ok := fields[field]; !ok {
			t.Fatalf("forward response dropped compatibility field %q: %s", field, data)
		}
	}
}

func TestForwardingOperatePreservesDuplicateAndOrderingContracts(t *testing.T) {
	existing := &fakeForwardingAdapter{name: "ufw", rules: []forwardClient.Rule{
		{Protocol: "tcp", Port: "8080", TargetIP: "127.0.0.1", TargetPort: "80"},
	}}
	service := forwardingServiceWithAdapter(existing)
	err := service.Operate(dto.ForwardRuleOperate{Rules: []dto.ForwardRuleOperation{{
		Operation: "add", Protocol: "tcp", Port: "8080", TargetPort: "80",
	}}})
	if err == nil {
		t.Fatal("duplicate forwarding rule must be rejected")
	}
	if len(existing.calls) != 0 {
		t.Fatalf("duplicate check wrote forwarding state: %#v", existing.calls)
	}

	adapter := &fakeForwardingAdapter{name: "iptables"}
	service = forwardingServiceWithAdapter(adapter)
	err = service.Operate(dto.ForwardRuleOperate{Rules: []dto.ForwardRuleOperation{
		{Operation: "add", Protocol: "tcp/udp", Port: "9000", TargetIP: "10.0.0.2", TargetPort: "90"},
		{Operation: "remove", Num: "1", Protocol: "tcp", Port: "8001", TargetIP: "10.0.0.2", TargetPort: "81"},
		{Operation: "remove", Num: "3", Protocol: "tcp", Port: "8003", TargetIP: "10.0.0.2", TargetPort: "83"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := []forwardingCall{
		{operation: "remove", rule: forwardClient.Rule{Num: "3", Protocol: "tcp", Port: "8003", TargetIP: "10.0.0.2", TargetPort: "83"}},
		{operation: "remove", rule: forwardClient.Rule{Num: "1", Protocol: "tcp", Port: "8001", TargetIP: "10.0.0.2", TargetPort: "81"}},
		{operation: "add", rule: forwardClient.Rule{Protocol: "tcp", Port: "9000", TargetIP: "10.0.0.2", TargetPort: "90"}},
		{operation: "add", rule: forwardClient.Rule{Protocol: "udp", Port: "9000", TargetIP: "10.0.0.2", TargetPort: "90"}},
	}
	if !reflect.DeepEqual(adapter.calls, want) {
		t.Fatalf("operation order changed\ngot  %#v\nwant %#v", adapter.calls, want)
	}
}

func TestForwardingSearchReturnsAdapterError(t *testing.T) {
	wantErr := errors.New("list failed")
	service := forwardingServiceWithAdapter(&fakeForwardingAdapter{name: "firewalld", listErr: wantErr})
	_, _, err := service.SearchWithPage(dto.ForwardRuleSearch{PageInfo: dto.PageInfo{Page: 1, PageSize: 20}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}
}
