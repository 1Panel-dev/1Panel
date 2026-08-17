package nftables_helper

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func LoadInitStatus(tab string) (bool, bool, error) {
	if tab != "base" {
		return false, false, nil
	}
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		for _, chain := range BasicChains() {
			if _, err := run("list", "chain", TableFamily(family), TableName, chain); err != nil {
				return false, false, nil
			}
		}
	}
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		stdout, err := run("list", "chain", TableFamily(family), TableName, InputChain)
		if err != nil {
			return false, false, nil
		}
		for _, chain := range BasicChains() {
			if !strings.Contains(stdout, "jump "+chain) {
				return true, false, nil
			}
		}
	}
	return true, true, nil
}
