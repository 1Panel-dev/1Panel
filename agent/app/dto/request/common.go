package request

type ReqWithID struct {
	ID uint `json:"id" validate:"required"`
}
