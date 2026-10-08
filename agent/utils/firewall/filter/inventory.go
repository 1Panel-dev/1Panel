package filter

type PositionRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type InventoryItem struct {
	Rule          FirewallRule  `json:"rule"`
	Observed      *ObservedRule `json:"observed"`
	IsWhitelist   bool          `json:"isWhitelist"`
	DescriptionID string        `json:"descriptionID"`
}
