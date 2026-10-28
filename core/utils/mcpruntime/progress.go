package mcpruntime

import "context"

type Progress struct {
	Status  string `json:"status,omitempty"`
	ID      string `json:"id,omitempty"`
	Current int64  `json:"current,omitempty"`
	Total   int64  `json:"total,omitempty"`
}

type progressContextKey struct{}

func WithProgress(ctx context.Context, report func(Progress)) context.Context {
	return context.WithValue(ctx, progressContextKey{}, report)
}

func ReportProgress(ctx context.Context, progress Progress) {
	if report, ok := ctx.Value(progressContextKey{}).(func(Progress)); ok {
		report(progress)
	}
}

type OperationError struct{ message string }

func (e *OperationError) Error() string { return e.message }

func operationError(message string) error { return &OperationError{message: message} }
