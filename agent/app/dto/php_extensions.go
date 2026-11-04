package dto

// PHPExtensionsPush is submitted to the master agent, with an explicit source node.
type PHPExtensionsPush struct {
	SourceNode  string   `json:"sourceNode" validate:"required"`
	TargetNodes []string `json:"targetNodes" validate:"required,min=1,dive,required"`
	ID          uint     `json:"id" validate:"required"`
	TaskID      string   `json:"taskID" validate:"required,uuid"`
}
