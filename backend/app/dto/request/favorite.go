package request

type FavoriteCreate struct {
	Path string `json:"path" validate:"required"`
}

type FavoriteDelete struct {
	ID uint `json:"id" validate:"required"`
}
