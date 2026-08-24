package utils

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func TestFirewallTransferRetriesCleanupWithoutDuplicatingData(t *testing.T) {
	db := newFirewallTransferTestDB(t)
	cleanupCalls := 0
	transfer := &firewallTransfer{
		db: db,
		load: func() (firewallTransferSource, error) {
			legacy := legacyFirewalldForward{
				rule: forwarding.Rule{Protocol: "tcp", Port: "8443", TargetIP: "10.0.0.8", TargetPort: "443"},
				spec: "port=8443:proto=tcp:toport=443:toaddr=10.0.0.8",
			}
			return firewallTransferSource{
				rules: []forwarding.Rule{legacy.rule}, firewalld: []legacyFirewalldForward{legacy}, provider: "iptables",
				cleanupOld: func([]legacyFirewalldForward) error {
					cleanupCalls++
					if cleanupCalls == 1 {
						return errors.New("temporary cleanup failure")
					}
					return nil
				},
			}, nil
		},
	}

	if err := transfer.run(context.Background()); err == nil {
		t.Fatal("first cleanup failure was ignored")
	}
	if err := transfer.run(context.Background()); err != nil {
		t.Fatalf("retry firewall transfer: %v", err)
	}
	var count int64
	if err := db.Model(&model.ForwardingRule{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 || cleanupCalls != 2 {
		t.Fatalf("retry result count=%d cleanupCalls=%d", count, cleanupCalls)
	}
	completed, err := firewallTransferCompleted(db)
	if err != nil || !completed {
		t.Fatalf("retry completion = %v, err=%v", completed, err)
	}
}

func TestFirewallTransferEmptyInventoryMarksCompletionWithoutEnabling(t *testing.T) {
	db := newFirewallTransferTestDB(t)
	transfer := &firewallTransfer{
		db: db,
		load: func() (firewallTransferSource, error) {
			return firewallTransferSource{}, nil
		},
	}
	if err := transfer.run(context.Background()); err != nil {
		t.Fatal(err)
	}
	completed, err := firewallTransferCompleted(db)
	if err != nil || !completed {
		t.Fatalf("empty transfer completion = %v, err=%v", completed, err)
	}
	var count int64
	if err := db.Model(&model.Setting{}).Where("key IN ?", []string{"IptablesForwardStatus", "ForwardingBackend"}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("empty transfer created %d forwarding settings", count)
	}
}

func TestFirewallTransferInvalidInventoryIsAtomicAndRetryable(t *testing.T) {
	db := newFirewallTransferTestDB(t)
	transfer := &firewallTransfer{
		db: db,
		load: func() (firewallTransferSource, error) {
			return firewallTransferSource{rules: []forwarding.Rule{
				{Protocol: "tcp", Port: "8080", TargetIP: "127.0.0.1", TargetPort: "80"},
				{Protocol: "tcp", Port: "invalid", TargetIP: "127.0.0.1", TargetPort: "80"},
			}}, nil
		},
	}
	if err := transfer.run(context.Background()); err == nil {
		t.Fatal("invalid legacy forwarding inventory was accepted")
	}
	completed, err := firewallTransferCompleted(db)
	if err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&model.ForwardingRule{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if completed || count != 0 {
		t.Fatalf("invalid transfer completed=%v imported=%d", completed, count)
	}
}

func TestFirewallTransferUpdatesExistingSettings(t *testing.T) {
	db := newFirewallTransferTestDB(t)
	settings := []model.Setting{
		{Key: "IptablesForwardStatus", Value: constant.StatusDisable},
		{Key: "ForwardingBackend", Value: "nftables"},
	}
	if err := db.Create(&settings).Error; err != nil {
		t.Fatal(err)
	}
	transfer := &firewallTransfer{
		db: db,
		load: func() (firewallTransferSource, error) {
			return firewallTransferSource{
				rules:    []forwarding.Rule{{Protocol: "udp", Port: "5353", TargetIP: "127.0.0.1", TargetPort: "53"}},
				provider: "iptables",
			}, nil
		},
	}
	if err := transfer.run(context.Background()); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"IptablesForwardStatus": constant.StatusEnable,
		"ForwardingBackend":     "iptables",
	} {
		var matches []model.Setting
		if err := db.Where("key = ?", key).Find(&matches).Error; err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 || matches[0].Value != want {
			t.Fatalf("setting %s after transfer = %#v, want %q", key, matches, want)
		}
	}
}

func TestFirewallTransferValidatesDependencies(t *testing.T) {
	if err := (&firewallTransfer{}).run(context.Background()); err == nil {
		t.Fatal("nil transfer database was accepted")
	}
	db := newFirewallTransferTestDB(t)
	if err := (&firewallTransfer{db: db}).run(context.Background()); err == nil {
		t.Fatal("nil legacy loader was accepted")
	}
	transfer := &firewallTransfer{
		db: db,
		load: func() (firewallTransferSource, error) {
			legacy := legacyFirewalldForward{rule: forwarding.Rule{
				Protocol: "tcp", Port: "8080", TargetIP: "127.0.0.1", TargetPort: "80",
			}}
			return firewallTransferSource{rules: []forwarding.Rule{legacy.rule}, firewalld: []legacyFirewalldForward{legacy}}, nil
		},
	}
	if err := transfer.run(context.Background()); err == nil {
		t.Fatal("missing firewalld cleanup was accepted")
	}
	completed, err := firewallTransferCompleted(db)
	if err != nil || completed {
		t.Fatalf("dependency failure completion=%v err=%v", completed, err)
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
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "migration.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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
