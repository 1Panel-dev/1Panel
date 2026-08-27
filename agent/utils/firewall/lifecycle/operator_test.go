package lifecycle

import (
	"errors"
	"testing"
)

type operatorTestClient struct {
	name    string
	started bool
	stopped bool
}

func (f *operatorTestClient) Name() string             { return f.name }
func (f *operatorTestClient) Start() error             { f.started = true; return nil }
func (f *operatorTestClient) Stop() error              { f.stopped = true; return nil }
func (f *operatorTestClient) Restart() error           { return nil }
func (f *operatorTestClient) Status() (bool, error)    { return true, nil }
func (f *operatorTestClient) Version() (string, error) { return "test", nil }
func TestOperatorDelegatesLifecycleAndPreparesStart(t *testing.T) {
	client := &operatorTestClient{name: "iptables"}
	operator := NewOperator(client)
	prepared := false
	if err := operator.Operate("start", false, func(got Client) error {
		prepared = got == client
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !client.started || !prepared {
		t.Fatalf("start was not fully coordinated: started=%v prepared=%v", client.started, prepared)
	}
}

func TestOperatorKeepsFirewallRunningWhenPostStartPreparationFails(t *testing.T) {
	client := &operatorTestClient{name: "firewalld"}
	wantErr := errors.New("sync accepted ports")

	err := NewOperator(client).Operate(OperationStart, false, func(Client) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("start returned error %v, want %v", err, wantErr)
	}
	if !client.started || client.stopped {
		t.Fatalf("post-start failure changed firewall state: started=%v stopped=%v", client.started, client.stopped)
	}
}
