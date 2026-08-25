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
		initialized, bound, err := loadFamilyInitStatus(family)
		if err != nil || !initialized {
			return false, false, err
		}
		if !bound {
			return true, false, nil
		}
	}
	return true, true, nil
}

func LoadFamilyInitStatus(family filter.Family, tab string) (bool, bool, error) {
	if tab != "base" {
		return false, false, nil
	}
	if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
		return false, false, nil
	}
	return loadFamilyInitStatus(family)
}

func LoadFamilyBindStatus(family filter.Family) (bool, error) {
	if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
		return false, nil
	}
	stdout, err := run("list", "chain", TableFamily(family), TableName, InputChain)
	if err != nil {
		return false, err
	}
	return hasBaseChainBinding(stdout), nil
}

func hasBaseChainBinding(output string) bool {
	for _, chain := range BasicChains() {
		if strings.Contains(output, "jump "+chain) {
			return true
		}
	}
	return false
}

func loadFamilyInitStatus(family filter.Family) (bool, bool, error) {
	for _, chain := range BasicChains() {
		if _, err := run("list", "chain", TableFamily(family), TableName, chain); err != nil {
			return false, false, nil
		}
	}
	stdout, err := run("list", "chain", TableFamily(family), TableName, InputChain)
	if err != nil {
		return false, false, nil
	}
	for _, chain := range BasicChains() {
		if !strings.Contains(stdout, "jump "+chain) {
			return true, false, nil
		}
	}
	return true, true, nil
}
