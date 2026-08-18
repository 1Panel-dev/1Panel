package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestFirewallTransferImportsCleansAndMarksCompletion(t *testing.T) {
	db := newFirewallTransferTestDB(t)
	loadCalls, cleanupCalls := 0, 0
	transfer := &firewallTransfer{
		db: db,
		load: func() (firewallTransferSource, error) {
			loadCalls++
			legacy := legacyFirewalldForward{
				rule: forwarding.Rule{Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80"},
				spec: "port=8080:proto=tcp:toport=80:toaddr=10.0.0.2",
			}
			return firewallTransferSource{
				rules:     []forwarding.Rule{legacy.rule, legacy.rule},
				firewalld: []legacyFirewalldForward{legacy},
				provider:  "iptables",
				cleanupOld: func(items []legacyFirewalldForward) error {
					cleanupCalls++
					if len(items) != 1 {
						return errors.New("unexpected cleanup inventory")
					}
					return nil
				},
			}, nil
		},
	}

	if err := transfer.run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if loadCalls != 1 || cleanupCalls != 1 {
		t.Fatalf("unexpected calls: load=%d cleanup=%d", loadCalls, cleanupCalls)
	}
	completed, err := firewallTransferCompleted(db)
	if err != nil || !completed {
		t.Fatalf("firewall transfer completion = %v, err=%v", completed, err)
	}
	var setting model.Setting
	if err := db.Where("key = ?", "IptablesForwardStatus").First(&setting).Error; err != nil {
		t.Fatal(err)
	}
	if setting.Value != constant.StatusEnable {
		t.Fatalf("forwarding status = %q", setting.Value)
	}
	setting = model.Setting{}
	if err := db.Where("key = ?", "ForwardingBackend").First(&setting).Error; err != nil {
		t.Fatal(err)
	}
	if setting.Value != "iptables" {
		t.Fatalf("forwarding backend = %q", setting.Value)
	}

	if err := transfer.run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if loadCalls != 1 || cleanupCalls != 1 {
		t.Fatalf("completed transfer ran again: load=%d cleanup=%d", loadCalls, cleanupCalls)
	}
}

func TestFirewallTransferFailureRemainsRetryable(t *testing.T) {
	db := newFirewallTransferTestDB(t)
	wantErr := errors.New("cleanup failed")
	transfer := &firewallTransfer{
		db: db,
		load: func() (firewallTransferSource, error) {
			legacy := legacyFirewalldForward{rule: forwarding.Rule{
				Protocol: "udp", Port: "5353", TargetIP: "127.0.0.1", TargetPort: "53",
			}}
			return firewallTransferSource{
				rules:     []forwarding.Rule{legacy.rule},
				firewalld: []legacyFirewalldForward{legacy},
				provider:  "iptables",
				cleanupOld: func([]legacyFirewalldForward) error {
					return wantErr
				},
			}, nil
		},
	}
	if err := transfer.run(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("transfer error = %v, want %v", err, wantErr)
	}
	completed, err := firewallTransferCompleted(db)
	if err != nil {
		t.Fatal(err)
	}
	if completed {
		t.Fatal("failed transfer was marked complete")
	}
	var count int64
	if err := db.Model(&model.ForwardingRule{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("retry inventory count = %d, want 1", count)
	}
}

func TestParseLegacyFirewalldForwarding(t *testing.T) {
	rules := parseLegacyFirewalldForwarding(
		"port=8080:proto=tcp:toport=80:toaddr=10.0.0.2\n" +
			"port=8443:proto=tcp:toport=443:toaddr=\ninvalid\n",
	)
	if len(rules) != 2 {
		t.Fatalf("parsed %d firewalld rules", len(rules))
	}
	if rules[0].rule.Family != forwarding.FamilyIPv4 || rules[0].rule.TargetIP != "10.0.0.2" {
		t.Fatalf("unexpected remote rule: %#v", rules[0])
	}
	if rules[1].rule.TargetIP != "127.0.0.1" || rules[1].rule.TargetPort != "443" {
		t.Fatalf("unexpected local rule: %#v", rules[1])
	}
}

func TestParseLegacyIptablesForwarding(t *testing.T) {
	stdout := "1 0 0 DNAT 6 -- eth0 * 0.0.0.0/0 0.0.0.0/0 tcp dpt:8080 to:10.0.0.2:80\n" +
		"2 0 0 REDIRECT 17 -- * * 0.0.0.0/0 0.0.0.0/0 udp dpts:5353:5354 redir ports 53\n"
	rules := parseLegacyIptablesForwarding(stdout)
	if len(rules) != 2 {
		t.Fatalf("parsed %d iptables rules", len(rules))
	}
	if rules[0].Protocol != "tcp" || rules[0].Interface != "eth0" || rules[0].Port != "8080" ||
		rules[0].TargetIP != "10.0.0.2" || rules[0].TargetPort != "80" {
		t.Fatalf("unexpected DNAT rule: %#v", rules[0])
	}
	if rules[1].Protocol != "udp" || rules[1].Port != "5353-5354" ||
		rules[1].TargetIP != "127.0.0.1" || rules[1].TargetPort != "53" {
		t.Fatalf("unexpected REDIRECT rule: %#v", rules[1])
	}
}

func newFirewallTransferTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Setting{}, &model.ForwardingRule{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE migrations (id VARCHAR(255) PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}
	return db
}
