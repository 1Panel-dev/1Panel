package request

type ProcessReq struct {
	PID int32 `json:"PID"  validate:"required"`
}
