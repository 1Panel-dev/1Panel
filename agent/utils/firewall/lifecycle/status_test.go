package lifecycle

import (
	"errors"
	"testing"
)

type statusTestClient struct {
	active     bool
	statusErr  error
	versionErr error
	versioned  *bool
}

func (statusTestClient) Name() string            { return "ufw" }
func (statusTestClient) Start() error            { return nil }
func (statusTestClient) Stop() error             { return nil }
func (statusTestClient) Restart() error          { return nil }
func (c statusTestClient) Status() (bool, error) { return c.active, c.statusErr }
func (c statusTestClient) Version() (string, error) {
	if c.versioned != nil {
		*c.versioned = true
	}
	return "1.0", c.versionErr
}

func TestLoadStatusAggregatesClientState(t *testing.T) {
	status, err := LoadStatus(statusTestClient{active: true})
	if err != nil {
		t.Fatal(err)
	}
	if status.Name != "ufw" || status.Version != "1.0" || !status.IsActive {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestLoadStatusReturnsClientErrors(t *testing.T) {
	statusErr := errors.New("status failed")
	versionErr := errors.New("version failed")
	_, err := LoadStatus(statusTestClient{active: true, statusErr: statusErr, versionErr: versionErr})
	if !errors.Is(err, statusErr) || !errors.Is(err, versionErr) {
		t.Fatalf("got %v, want joined status and version errors", err)
	}
}

func TestLoadStatusIgnoresVersionFailureForInactiveFirewall(t *testing.T) {
	versioned := false
	status, err := LoadStatus(statusTestClient{versionErr: errors.New("firewall is stopped"), versioned: &versioned})
	if err != nil {
		t.Fatal(err)
	}
	if status.Name != "ufw" || status.Version != "-" || status.IsActive || !versioned {
		t.Fatalf("unexpected inactive status: %#v versioned=%v", status, versioned)
	}
}
