package nftables_helper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

const RulesFile = "1panel_filter.nft"

func PersistRuleset(ctx context.Context) error {
	var ruleset strings.Builder
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		stdout, err := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second)).
			RunWithOptionalSudoAndStdout("nft", "list", "table", TableFamily(family), TableName)
		if err != nil {
			return err
		}
		ruleset.WriteString(stdout)
		if !strings.HasSuffix(stdout, "\n") {
			ruleset.WriteByte('\n')
		}
	}
	return atomicWrite(filepath.Join(global.Dir.FirewallDir, RulesFile), []byte(ruleset.String()))
}

func Restore() error {
	file := filepath.Join(global.Dir.FirewallDir, RulesFile)
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	existing := make([]string, 0, 2)
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		tableFamily := TableFamily(family)
		if _, err := run("list", "table", tableFamily, TableName); err == nil {
			existing = append(existing, tableFamily)
		}
	}
	if len(existing) == 2 {
		return nil
	}
	for _, tableFamily := range existing {
		if _, err := run("delete", "table", tableFamily, TableName); err != nil {
			return fmt.Errorf("remove partial nftables %s table before restore: %w", tableFamily, err)
		}
	}
	return runCommand("-f", file)
}

func atomicWrite(target string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(target), ".nft-rules-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	committed := false
	defer func() {
		_ = file.Close()
		if !committed {
			_ = os.Remove(temporary)
		}
	}()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary, target); err != nil {
		return fmt.Errorf("replace nftables rules file: %w", err)
	}
	committed = true
	return nil
}
