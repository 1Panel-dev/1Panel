package lifecycle

import (
	"errors"
	"testing"
)

type statusTestClient struct {
	statusErr  error
	versionErr error
}

func (statusTestClient) Name() string               { return "ufw" }
func (statusTestClient) Start() error               { return nil }
func (statusTestClient) Stop() error                { return nil }
func (statusTestClient) Restart() error             { return nil }
func (c statusTestClient) Status() (bool, error)    { return true, c.statusErr }
func (c statusTestClient) Version() (string, error) { return "1.0", c.versionErr }

func TestLoadStatusAggregatesClientState(t *testing.T) {
	status, err := LoadStatus(statusTestClient{})
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
	_, err := LoadStatus(statusTestClient{statusErr: statusErr, versionErr: versionErr})
	if !errors.Is(err, statusErr) || !errors.Is(err, versionErr) {
		t.Fatalf("got %v, want joined status and version errors", err)
	}
}
