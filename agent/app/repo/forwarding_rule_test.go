package repo

import (
	"context"
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestForwardingRuleRepoReplaceAll(t *testing.T) {
	previousDB := global.DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.ForwardingRule{}); err != nil {
		t.Fatal(err)
	}
	global.DB = db
	t.Cleanup(func() { global.DB = previousDB })

	repository := NewIForwardingRuleRepo()
	first := []model.ForwardingRule{{Family: "ipv4", Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80"}}
	if err := repository.ReplaceAll(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	rules, err := repository.List(context.Background())
	if err != nil || len(rules) != 1 || rules[0].Port != "8080" {
		t.Fatalf("rules = %#v, err = %v", rules, err)
	}

	second := []model.ForwardingRule{{Family: "ipv6", Protocol: "udp", Port: "5353", TargetIP: "::1", TargetPort: "53", Interface: "eth0"}}
	if err := repository.ReplaceAll(context.Background(), second); err != nil {
		t.Fatal(err)
	}
	rules, err = repository.List(context.Background())
	if err != nil || len(rules) != 1 || rules[0].Family != "ipv6" || rules[0].Port != "5353" {
		t.Fatalf("rules = %#v, err = %v", rules, err)
	}
}
