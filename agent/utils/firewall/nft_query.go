package firewall

import (
	"fmt"
	"slices"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func ReadNftObject(run func(...string) (string, error), args ...string) (string, bool, error) {
	output, err := run(args...)
	if err == nil {
		return output, true, nil
	}
	message := err.Error()
	if slices.Contains(args, "ip6") && (strings.Contains(message, "Address family not supported") || strings.Contains(message, "Protocol not supported")) {
		return "", false, fmt.Errorf("%w: %v", filter.ErrFamilyUnavailable, err)
	}
	if strings.Contains(message, "Error:") && strings.Contains(message, "No such file or directory") &&
		!strings.Contains(message, "Operation not permitted") && !strings.Contains(message, "Permission denied") {
		return "", false, nil
	}
	return "", false, err
}
