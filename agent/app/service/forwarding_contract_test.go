package service

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	forwardClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/go-playground/validator/v10"
)

type fakeForwardingAdapter struct {
	name       string
	rules      []forwardClient.Rule
	listErr    error
	operateErr error
	init       bool
	initErr    error
	familyInit map[string]bool
	reconciled []forwardClient.Rule
}

func (f *fakeForwardingAdapter) Name() string { return f.name }

func (f *fakeForwardingAdapter) List() ([]forwardClient.Rule, error) {
	return append([]forwardClient.Rule(nil), f.rules...), f.listErr
}

func (f *fakeForwardingAdapter) Reconcile(rules []forwardClient.Rule) error {
	f.reconciled = append([]forwardClient.Rule(nil), rules...)
	if f.operateErr == nil {
		f.rules = append([]forwardClient.Rule(nil), rules...)
	}
	return f.operateErr
}

func (f *fakeForwardingAdapter) Enable() error  { return nil }
func (f *fakeForwardingAdapter) Cleanup() error { return nil }
func (f *fakeForwardingAdapter) FamilyStatus(family string) (bool, bool, error) {
	if f.familyInit != nil {
		initialized := f.familyInit[family]
		return initialized, initialized, f.initErr
	}
	return f.init, f.init, f.initErr
}
func (f *fakeForwardingAdapter) InitStatus() (bool, bool, error) {
	return f.init, f.init, f.initErr
}
func (f *fakeForwardingAdapter) Replay() error { return nil }

type fakeForwardingRuleRepo struct {
	rules   []model.ForwardingRule
	listErr error
}

func (r *fakeForwardingRuleRepo) List(context.Context) ([]model.ForwardingRule, error) {
	return append([]model.ForwardingRule(nil), r.rules...), r.listErr
}

func (r *fakeForwardingRuleRepo) ReplaceAll(_ context.Context, rules []model.ForwardingRule) error {
	r.rules = append([]model.ForwardingRule(nil), rules...)
	return nil
}

func forwardingServiceWithAdapter(adapter forwardClient.Adapter) *ForwardingService {
	rules, listErr := adapter.List()
	return &ForwardingService{
		managerFactory: func() (*forwardClient.Manager, error) {
			return forwardClient.NewManager(adapter, nil), nil
		},
		rules:          &fakeForwardingRuleRepo{rules: forwardingRuleModels(rules), listErr: listErr},
		enabled:        func() (bool, error) { return true, nil },
		persistBackend: func(string) error { return nil },
	}
}

func TestForwardingAndFilterInterfacesAreSeparated(t *testing.T) {
	filterType := reflect.TypeOf((*lifecycle.Client)(nil)).Elem()
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
	for _, method := range []string{"LoadBaseInfo", "SearchRules", "OperateRules", "Enable", "Restore"} {
		if _, ok := forwardingServiceType.MethodByName(method); !ok {
			t.Fatalf("forwarding service missing %s", method)
		}
	}
}

func TestForwardingInitNoLongerUsesFilterRequest(t *testing.T) {
	request := dto.FilterChainOperation{Name: "1PANEL_FORWARD", Operate: "init-forward"}
	if err := validator.New().Struct(request); err == nil {
		t.Fatal("forwarding initialization must use the dedicated forwarding endpoint")
	}
}

func TestForwardingRestoreHonorsPersistedState(t *testing.T) {
	adapter := &fakeForwardingAdapter{name: "iptables", rules: []forwardClient.Rule{{
		Family: forwardClient.FamilyIPv4, Protocol: "tcp", Port: "8080", TargetIP: "127.0.0.1", TargetPort: "80",
	}}}
	service := forwardingServiceWithAdapter(adapter)
	service.enabled = func() (bool, error) { return false, nil }
	if err := service.Restore(context.Background()); err != nil {
		t.Fatal(err)
	}
	if adapter.reconciled != nil {
		t.Fatalf("disabled forwarding was restored: %#v", adapter.reconciled)
	}

	service.enabled = func() (bool, error) { return true, nil }
	if err := service.Restore(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(adapter.reconciled) != 1 || adapter.reconciled[0].Port != "8080" {
		t.Fatalf("unexpected restored rules: %#v", adapter.reconciled)
	}
}

func TestForwardingBackendSelection(t *testing.T) {
	nft := &fakeForwardingAdapter{name: "nftables"}
	iptables := &fakeForwardingAdapter{name: "iptables"}

	manager, err := selectForwardingManager([]forwardingCandidate{{adapter: nft}, {adapter: iptables}})
	if err != nil {
		t.Fatal(err)
	}
	if manager.Name() != "nftables" {
		t.Fatalf("uninitialized backends selected %q, want nftables", manager.Name())
	}

	iptables.init = true
	manager, err = selectForwardingManager([]forwardingCandidate{{adapter: nft}, {adapter: iptables}})
	if err != nil {
		t.Fatal(err)
	}
	if manager.Name() != "iptables" {
		t.Fatalf("initialized backend selected %q, want iptables", manager.Name())
	}

	nft.init = true
	if _, err := selectForwardingManager([]forwardingCandidate{{adapter: nft}, {adapter: iptables}}); !errors.Is(err, errForwardingBackendConflict) {
		t.Fatalf("both initialized backends returned %v, want conflict", err)
	}
}

func TestForwardingBackendSelectionRejectsSplitFamilies(t *testing.T) {
	nft := &fakeForwardingAdapter{name: "nftables", familyInit: map[string]bool{forwardClient.FamilyIPv4: true}}
	iptables := &fakeForwardingAdapter{name: "iptables", familyInit: map[string]bool{forwardClient.FamilyIPv6: true}}
	if _, err := selectForwardingManager([]forwardingCandidate{{adapter: nft}, {adapter: iptables}}); !errors.Is(err, errForwardingBackendConflict) {
		t.Fatalf("split-family backends returned %v, want conflict", err)
	}
}

func TestForwardingBackendSelectionReturnsStatusError(t *testing.T) {
	wantErr := errors.New("status failed")
	_, err := selectForwardingManager([]forwardingCandidate{{adapter: &fakeForwardingAdapter{name: "nftables", initErr: wantErr}}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}
}

func TestForwardingDisplayName(t *testing.T) {
	for backend, want := range map[string]string{
		"iptables": "iptables-forward",
		"nftables": "nftables-forward",
		"unknown":  "unknown",
	} {
		if got := forwardingDisplayName(backend); got != want {
			t.Fatalf("forwardingDisplayName(%q) = %q, want %q", backend, got, want)
		}
	}
}

func TestForwardingSearchPreservesAPIShapeAndPagination(t *testing.T) {
	adapter := &fakeForwardingAdapter{name: "iptables", rules: []forwardClient.Rule{
		{Num: "1", Family: forwardClient.FamilyIPv6, Protocol: "tcp", Port: "8080", TargetIP: "2001:db8::2", TargetPort: "80", Interface: "eth0"},
		{Num: "2", Protocol: "udp", Port: "5353", TargetIP: "127.0.0.1", TargetPort: "53"},
	}}
	service := forwardingServiceWithAdapter(adapter)
	service.rules.(*fakeForwardingRuleRepo).rules[0].ID = 42
	total, value, err := service.SearchRules(dto.ForwardRuleSearch{PageInfo: dto.PageInfo{Page: 1, PageSize: 10}, Info: "2001:db8"})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("got total %d want 1", total)
	}
	items, ok := value.([]dto.ForwardRule)
	if !ok || len(items) != 1 || items[0].ID != 42 || items[0].Port != "8080" || items[0].Family != forwardClient.FamilyIPv6 {
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

func TestForwardingDuplicateIdentityIncludesAddressFamily(t *testing.T) {
	adapter := &fakeForwardingAdapter{name: "nftables", rules: []forwardClient.Rule{{
		Family: forwardClient.FamilyIPv4, Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80",
	}}}
	service := forwardingServiceWithAdapter(adapter)
	err := service.OperateRules(dto.ForwardRuleOperate{Rules: []dto.ForwardRuleOperation{{
		Operation: "add", Family: forwardClient.FamilyIPv6, Protocol: "tcp", Port: "8080", TargetIP: "2001:db8::2", TargetPort: "80",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(adapter.reconciled) != 2 || adapter.reconciled[1].Family != forwardClient.FamilyIPv6 {
		t.Fatalf("unexpected reconciled rules: %#v", adapter.reconciled)
	}
}

func TestForwardingOperatePreservesDuplicateAndOrderingContracts(t *testing.T) {
	existing := &fakeForwardingAdapter{name: "iptables", rules: []forwardClient.Rule{
		{Protocol: "tcp", Port: "8080", TargetIP: "127.0.0.1", TargetPort: "80"},
	}}
	service := forwardingServiceWithAdapter(existing)
	err := service.OperateRules(dto.ForwardRuleOperate{Rules: []dto.ForwardRuleOperation{{
		Operation: "add", Protocol: "tcp", Port: "8080", TargetPort: "80",
	}}})
	if err == nil {
		t.Fatal("duplicate forwarding rule must be rejected")
	}
	adapter := &fakeForwardingAdapter{name: "iptables"}
	service = forwardingServiceWithAdapter(adapter)
	err = service.OperateRules(dto.ForwardRuleOperate{Rules: []dto.ForwardRuleOperation{
		{Operation: "add", Protocol: "tcp/udp", Port: "9000", TargetIP: "10.0.0.2", TargetPort: "90"},
		{Operation: "remove", Num: "1", Protocol: "tcp", Port: "8001", TargetIP: "10.0.0.2", TargetPort: "81"},
		{Operation: "remove", Num: "3", Protocol: "tcp", Port: "8003", TargetIP: "10.0.0.2", TargetPort: "83"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := []forwardClient.Rule{
		{Family: forwardClient.FamilyIPv4, Protocol: "tcp", Port: "9000", TargetIP: "10.0.0.2", TargetPort: "90"},
		{Family: forwardClient.FamilyIPv4, Protocol: "udp", Port: "9000", TargetIP: "10.0.0.2", TargetPort: "90"},
	}
	if !reflect.DeepEqual(adapter.reconciled, want) {
		t.Fatalf("desired forwarding rules changed\ngot  %#v\nwant %#v", adapter.reconciled, want)
	}
}

func TestForwardingSearchReturnsAdapterError(t *testing.T) {
	wantErr := errors.New("list failed")
	service := forwardingServiceWithAdapter(&fakeForwardingAdapter{name: "nftables", listErr: wantErr})
	_, _, err := service.SearchRules(dto.ForwardRuleSearch{PageInfo: dto.PageInfo{Page: 1, PageSize: 20}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}
}

func TestForwardingOperateReturnsAdapterListError(t *testing.T) {
	wantErr := errors.New("list failed")
	service := forwardingServiceWithAdapter(&fakeForwardingAdapter{name: "iptables", listErr: wantErr})
	err := service.OperateRules(dto.ForwardRuleOperate{Rules: []dto.ForwardRuleOperation{{
		Operation: "remove", Protocol: "tcp", Port: "8080", TargetPort: "80",
	}}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}
}

func TestForwardingForceDeleteOnlySuppressesRemoveErrors(t *testing.T) {
	wantErr := errors.New("operate failed")
	removeAdapter := &fakeForwardingAdapter{name: "iptables", operateErr: wantErr}
	removeService := forwardingServiceWithAdapter(removeAdapter)
	err := removeService.OperateRules(dto.ForwardRuleOperate{ForceDelete: true, Rules: []dto.ForwardRuleOperation{{
		Operation: "remove", Protocol: "tcp", Port: "8080", TargetPort: "80",
	}}})
	if err != nil {
		t.Fatalf("forced remove returned %v", err)
	}

	addAdapter := &fakeForwardingAdapter{name: "iptables", operateErr: wantErr}
	addService := forwardingServiceWithAdapter(addAdapter)
	err = addService.OperateRules(dto.ForwardRuleOperate{ForceDelete: true, Rules: []dto.ForwardRuleOperation{{
		Operation: "add", Protocol: "tcp", Port: "8080", TargetPort: "80",
	}}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}
}

func TestForwardingOperateKeepsDesiredStateWhenRuntimeReconcileFails(t *testing.T) {
	wantErr := errors.New("runtime reconcile failed")
	adapter := &fakeForwardingAdapter{name: "iptables", operateErr: wantErr}
	rules := &fakeForwardingRuleRepo{}
	service := forwardingServiceWithAdapter(adapter)
	service.rules = rules

	err := service.OperateRules(dto.ForwardRuleOperate{Rules: []dto.ForwardRuleOperation{{
		Operation: "add", Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80",
	}}})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v want %v", err, wantErr)
	}
	if len(rules.rules) != 1 || rules.rules[0].Port != "8080" {
		t.Fatalf("database desired state was lost after runtime failure: %#v", rules.rules)
	}

}
