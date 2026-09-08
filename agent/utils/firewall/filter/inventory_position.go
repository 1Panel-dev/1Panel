package filter

func InventoryPositionRanges(provider Provider, items []InventoryItem) (ipv4, ipv6 PositionRange) {
	if provider == ProviderFirewalld {
		return PositionRange{Min: -32768, Max: 32767}, PositionRange{Min: -32768, Max: 32767}
	}
	for _, item := range items {
		if item.Observed == nil || item.Observed.Locator.Position == nil {
			continue
		}
		scope := item.Observed.Rule.Scope
		if scope.Provider != provider || scope.Direction != DirectionInput {
			continue
		}
		if (provider == ProviderIptables || provider == ProviderNftables) &&
			(scope.Table != "filter" || scope.Chain != IptablesInputChain) {
			continue
		}
		bounds := &ipv4
		if scope.Family == FamilyIPv6 {
			bounds = &ipv6
		} else if scope.Family != FamilyIPv4 {
			continue
		}
		position := *item.Observed.Locator.Position
		if position < 1 {
			continue
		}
		if bounds.Min == 0 || position < bounds.Min {
			bounds.Min = position
		}
		bounds.Max = max(bounds.Max, position)
		if provider != ProviderUFW {
			bounds.Min = 1
		}
	}
	return
}
