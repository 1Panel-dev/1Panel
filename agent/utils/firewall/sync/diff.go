package sync

type Status string

const (
	StatusReady    Status = "ready"
	StatusExisting Status = "existing"
	StatusRemove   Status = "remove"
	StatusBlocked  Status = "blocked"
)

type Outcome string

const (
	OutcomeApplied Outcome = "applied"
	OutcomeSkipped Outcome = "skipped"
	OutcomeRemoved Outcome = "removed"
	OutcomeFailed  Outcome = "failed"
)

type ReasonCode string

const (
	ReasonInvalidPolicy       ReasonCode = "invalid_policy"
	ReasonAlreadyExists       ReasonCode = "already_exists_in_target"
	ReasonOnlyExistsInTarget  ReasonCode = "only_exists_in_target"
	ReasonManagedOnlyInTarget ReasonCode = "managed_only_exists_in_target"
	ReasonUnsafeRemoval       ReasonCode = "unsafe_managed_rule_removal"
	ReasonReadOnlyRule        ReasonCode = "read_only_rule"
)

func ReasonMessage(code ReasonCode) string {
	switch code {
	case ReasonAlreadyExists:
		return "rule already exists in target backend"
	case ReasonOnlyExistsInTarget:
		return "rule exists only in target backend"
	case ReasonManagedOnlyInTarget:
		return "managed rule exists only in target backend"
	case ReasonUnsafeRemoval:
		return "managed runtime rule cannot be safely removed"
	case ReasonReadOnlyRule:
		return "read-only runtime rule is preserved but cannot be synchronized"
	default:
		return ""
	}
}

type Desired[T any, P any] struct {
	Value   T
	Payload P
	Err     error
}

type Item[P any] struct {
	Payload    P
	Status     Status
	ReasonCode ReasonCode
	Reason     string
}

func Diff[T any, P any](desired []Desired[T, P], actual []T, key func(T) string, actualPayload func(T) P) []Item[P] {
	items := make([]Item[P], 0, len(desired)+len(actual))
	actualByKey := make(map[string][]int, len(actual))
	for index, value := range actual {
		actualByKey[key(value)] = append(actualByKey[key(value)], index)
	}
	matched := make([]bool, len(actual))
	for _, candidate := range desired {
		item := Item[P]{Payload: candidate.Payload}
		switch {
		case candidate.Err != nil:
			item.Status, item.ReasonCode, item.Reason = StatusBlocked, ReasonInvalidPolicy, candidate.Err.Error()
		default:
			match := unmatchedIndex(actualByKey[key(candidate.Value)], matched)
			if match >= 0 {
				matched[match] = true
				item.Status, item.ReasonCode = StatusExisting, ReasonAlreadyExists
				item.Reason = ReasonMessage(item.ReasonCode)
			} else {
				item.Status = StatusReady
			}
		}
		items = append(items, item)
	}
	for index, value := range actual {
		if matched[index] {
			continue
		}
		items = append(items, Item[P]{
			Payload: actualPayload(value), Status: StatusRemove, ReasonCode: ReasonOnlyExistsInTarget,
			Reason: ReasonMessage(ReasonOnlyExistsInTarget),
		})
	}
	return items
}

func StatesEqual[T any](left, right []T, key func(T) string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, value := range left {
		counts[key(value)]++
	}
	for _, value := range right {
		valueKey := key(value)
		if counts[valueKey] == 0 {
			return false
		}
		counts[valueKey]--
	}
	return true
}

func unmatchedIndex(indices []int, matched []bool) int {
	for _, index := range indices {
		if !matched[index] {
			return index
		}
	}
	return -1
}
