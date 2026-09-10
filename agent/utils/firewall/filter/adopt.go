package filter

import "fmt"

var ErrDuplicateAdoption = fmt.Errorf("%w: duplicate firewall rules prevent adoption; manually delete duplicate rules and retry", ErrRuleOperation)

func CheckAdoptDuplicates(snapshot Snapshot, requested FirewallRule) error {
	count := 0
	for _, observed := range snapshot.Rules {
		if observed.ParseStatus != ParseStatusSupported {
			continue
		}
		same, err := SameRuleContent(observed.Rule, requested)
		if err != nil {
			return err
		}
		if same {
			count++
			if count > 1 {
				return ErrDuplicateAdoption
			}
		}
	}
	return nil
}
