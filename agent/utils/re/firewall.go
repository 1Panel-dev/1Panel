package re

import "regexp"

var (
	UFWNumberedRulePrefixRegex = regexp.MustCompile(`^\s*\[\s*([0-9]+)\]\s+(.+?)\s*$`)
	UFWNumberedRuleRegex       = regexp.MustCompile(`^\s*\[\s*([0-9]+)\]\s+(.+?)\s+(ALLOW|DENY|REJECT|LIMIT)(?:\s+(IN|OUT|FWD))?\s+(.+?)\s*$`)
	ForwardInterfaceRegex      = regexp.MustCompile(`^[A-Za-z0-9_.:@-]{1,15}$`)
)
