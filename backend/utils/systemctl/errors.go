package systemctl

import "fmt"

type ServiceError struct {
	Action   string
	Service  string
	Output   string
	Wrapped  error
	ExitCode int
}

func (e ServiceError) Error() string {
	return fmt.Sprintf("service %s %s failed (exit %d): %v\n%s",
		e.Service, e.Action, e.ExitCode, e.Wrapped, e.Output)
}

func (e ServiceError) Unwrap() error {
	return e.Wrapped
}

type ServicePathError struct {
	Manager string
	Service string
}

func (e ServicePathError) Error() string {
	return fmt.Sprintf("service path resolution failed for %s/%s", e.Manager, e.Service)
}
