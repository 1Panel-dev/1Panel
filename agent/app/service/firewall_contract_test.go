package service

import (
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	fireClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
)

type firewallSettingRepoStub struct {
	repo.ISettingRepo
	value string
}

func (r *firewallSettingRepoStub) GetValueByKey(string) (string, error) {
	return r.value, nil
}

type whitelistFilterClient struct {
	firewall.FilterClient
	list fireClient.PortWhiteList
}

func (c *whitelistFilterClient) SyncPortWhiteList(list fireClient.PortWhiteList) error {
	c.list = list
	return nil
}

func TestParseFirewallPortWhiteListContract(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    []fireClient.PortWhiteListEntry
		wantErr bool
	}{
		{
			name:  "default value",
			value: constant.FirewallPortWhiteListValue,
			want: []fireClient.PortWhiteListEntry{
				{Port: "80", Protocol: "tcp"},
				{Port: "443", Protocol: "tcp"},
				{Port: "443", Protocol: "udp"},
			},
		},
		{
			name:  "dedup",
			value: "80/tcp,80/tcp,443/tcp",
			want: []fireClient.PortWhiteListEntry{
				{Port: "80", Protocol: "tcp"},
				{Port: "443", Protocol: "tcp"},
			},
		},
		{
			name:  "default protocol tcp",
			value: "8080",
			want: []fireClient.PortWhiteListEntry{
				{Port: "8080", Protocol: "tcp"},
			},
		},
		{
			name:    "invalid protocol",
			value:   "80/icmp",
			wantErr: true,
		},
		{
			name:    "invalid port",
			value:   "99999/tcp",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFirewallPortWhiteList(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v want %#v", got, tt.want)
			}
		})
	}
}

func TestSyncFirewallPortWhiteListBuildsProviderState(t *testing.T) {
	originalRepo, originalConf, originalMaster := settingRepo, global.CONF, global.IsMaster
	settingRepo = &firewallSettingRepoStub{value: "443/tcp"}
	global.IsMaster = false
	global.CONF.Base.Port = "9999"
	t.Cleanup(func() {
		settingRepo = originalRepo
		global.CONF = originalConf
		global.IsMaster = originalMaster
	})

	client := &whitelistFilterClient{}
	if err := syncFirewallPortWhiteListAfterUpdateWithClient(client, "80/tcp"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(client.list.Configured, []fireClient.PortWhiteListEntry{{Port: "443", Protocol: "tcp"}}) {
		t.Fatalf("unexpected configured list: %#v", client.list.Configured)
	}
	if !reflect.DeepEqual(client.list.Previous, []fireClient.PortWhiteListEntry{{Port: "80", Protocol: "tcp"}}) {
		t.Fatalf("unexpected previous list: %#v", client.list.Previous)
	}
	if len(client.list.Required) != 2 || client.list.Required[0] != (fireClient.PortWhiteListEntry{Port: "9999", Protocol: "tcp"}) {
		t.Fatalf("unexpected required list: %#v", client.list.Required)
	}
}

func TestCheckPortUsedContract(t *testing.T) {
	apps := []portOfApp{
		{AppName: "wordpress", HttpPort: "8080", HttpsPort: "8443"},
		{AppName: "1panel", HttpPort: "9999"},
	}
	if got := checkPortUsed("8080", "tcp", apps); got != "wordpress" {
		t.Fatalf("got %q want wordpress", got)
	}
	if got := checkPortUsed("9999", "tcp", apps); got != "1panel" {
		t.Fatalf("got %q want 1panel", got)
	}
}

func TestFirewallAPIDtoSnapshot(t *testing.T) {
	assertStructFields(t, dto.PortRuleOperate{}, []string{
		"ID", "Operation", "Chain", "Address", "Port", "Protocol", "Strategy", "Description",
	})
	assertStructFields(t, dto.AddrRuleOperate{}, []string{
		"ID", "Operation", "Address", "Strategy", "Description",
	})
	assertStructFields(t, dto.ForwardRuleOperate{}, []string{
		"ForceDelete", "Rules",
	})
	assertStructFields(t, dto.FirewallBaseInfo{}, []string{
		"Name", "IsExist", "IsActive", "IsInit", "IsBind", "Version", "PingStatus",
	})
	assertStructFields(t, dto.FirewallOperation{}, []string{
		"Operation", "WithDockerRestart",
	})
	assertStructFields(t, dto.BatchRuleOperate{}, []string{
		"Type", "Rules",
	})
	assertStructFields(t, dto.IptablesRuleOp{}, []string{
		"Operation", "ID", "Chain", "Protocol", "SrcIP", "SrcPort", "DstIP", "DstPort", "Strategy", "Description",
	})
}

func assertStructFields(t *testing.T, sample any, want []string) {
	t.Helper()
	typ := reflect.TypeOf(sample)
	if typ.Kind() != reflect.Struct {
		t.Fatalf("expected struct, got %s", typ.Kind())
	}
	got := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		got = append(got, typ.Field(i).Name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s fields changed\ngot  %#v\nwant %#v", typ.Name(), got, want)
	}
}
