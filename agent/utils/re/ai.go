package re

import "regexp"

var (
	AIErrorURLRegex    = regexp.MustCompile(`(?i)https?://[^\s<>"']+`)
	AIErrorAuthRegex   = regexp.MustCompile(`(?i)\b(bearer|basic)\s+[^\s,;"'<>]+`)
	AIErrorSecretRegex = regexp.MustCompile(`(?i)(["']?(?:api[-_]?key|access[-_]?token|refresh[-_]?token|token|password|secret|authorization|cookie)["']?\s*[:=]\s*)(?:"[^"]*"|'[^']*'|[^\s,;<>]+)`)

	VLLMAPIKeyFlagRegex     = regexp.MustCompile(`(^|[\s"'\[])--api-key([=\s"'\]]|$)`)
	VLLMEnvironmentKeyRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	VLLMCommandTokenRegex   = regexp.MustCompile(`^[a-zA-Z0-9_./:=,+-]+$`)
)
