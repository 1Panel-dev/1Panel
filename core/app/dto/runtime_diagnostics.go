package dto

import "time"

type RuntimeDiagnosticsSummary struct {
	RSS         uint64 `json:"rss"`
	HeapAlloc   uint64 `json:"heapAlloc"`
	HeapObjects uint64 `json:"heapObjects"`
	Goroutines  int    `json:"goroutines"`
}

type RuntimeGoroutineGroup struct {
	State string   `json:"state"`
	Top   string   `json:"top"`
	Count int      `json:"count"`
	Stack []string `json:"stack"`
}

type RuntimeGoroutineSnapshot struct {
	Total      int                     `json:"total"`
	GroupCount int                     `json:"groupCount"`
	Truncated  bool                    `json:"truncated"`
	CapturedAt time.Time               `json:"capturedAt"`
	Goroutines []RuntimeGoroutineGroup `json:"goroutines"`
}

type RuntimeProfileCreate struct {
	Type     string `json:"type" validate:"required,oneof=cpu heap goroutine mutex block"`
	Duration int    `json:"duration" validate:"omitempty,min=5,max=30"`
}
