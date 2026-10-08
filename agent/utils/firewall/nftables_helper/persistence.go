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
	run := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second))
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		stdout, exists, err := readNftObject(func(args ...string) (string, error) {
			return run.RunWithOptionalSudoAndStdout("nft", args...)
		}, "list", "table", TableFamily(family), TableName)
		if family == filter.FamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		ruleset.WriteString(stdout)
		if !strings.HasSuffix(stdout, "\n") {
			ruleset.WriteByte('\n')
		}
	}
	return atomicWrite(filepath.Join(global.Dir.FirewallDir, RulesFile), []byte(ruleset.String()))
}

func Restore(families ...filter.Family) error {
	if len(families) == 0 {
		families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
	}
	data, err := os.ReadFile(filepath.Join(global.Dir.FirewallDir, RulesFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var script strings.Builder
	for _, family := range families {
		tableFamily := TableFamily(family)
		_, exists, err := readNftObject(run, "list", "table", tableFamily, TableName)
		if family == filter.FamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		table, err := savedTableRules(string(data), tableFamily, TableName)
		if err != nil {
			return err
		}
		script.WriteString(table)
	}
	if script.Len() == 0 {
		return nil
	}
	return RunScript(script.String())
}

func savedTableRules(ruleset, family, table string) (string, error) {
	header := "table " + family + " " + table + " {"
	lines := strings.Split(ruleset, "\n")
	for index, line := range lines {
		if line != header {
			continue
		}
		for end := index + 1; end < len(lines); end++ {
			if lines[end] == "}" {
				return strings.Join(lines[index:end+1], "\n") + "\n", nil
			}
		}
		return "", fmt.Errorf("incomplete saved nftables table %s %s", family, table)
	}
	return "", nil
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
