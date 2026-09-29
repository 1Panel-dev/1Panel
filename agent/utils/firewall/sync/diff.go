package sync

type Status string

const (
	StatusReady    Status = "ready"
	StatusExisting Status = "existing"
	StatusRemove   Status = "remove"
	StatusBlocked  Status = "blocked"
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
