package dto

type LogFileRes struct {
	Lines       []string `json:"lines"`
	IsEndOfFile bool     `json:"isEndOfFile"`
	TotalPages  int      `json:"totalPages"`
	TotalLines  int      `json:"totalLines"`
}
