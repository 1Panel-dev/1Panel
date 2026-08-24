package lifecycle

import "testing"

type operatorTestClient struct {
	name    string
	started bool
}

func (f *operatorTestClient) Name() string             { return f.name }
func (f *operatorTestClient) Start() error             { f.started = true; return nil }
func (f *operatorTestClient) Stop() error              { return nil }
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
