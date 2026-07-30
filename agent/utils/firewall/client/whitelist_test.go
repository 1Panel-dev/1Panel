package client

import (
	"errors"
	"reflect"
	"testing"
)

type recordingNativePortWriter struct {
	active bool
	failOn string
	events []string
}

func (r *recordingNativePortWriter) Status() (bool, error) {
	r.events = append(r.events, "status")
	return r.active, nil
}

func (r *recordingNativePortWriter) Reload() error {
	r.events = append(r.events, "reload")
	if r.failOn == "reload" {
		return errors.New("reload failed")
	}
	return nil
}

func (r *recordingNativePortWriter) Port(port FireInfo, operation string) error {
	event := operation + " " + port.Port + "/" + port.Protocol
	r.events = append(r.events, event)
	if r.failOn == event {
		return errors.New("port operation failed")
	}
	return nil
}

func TestAddNativePortWhiteListSequence(t *testing.T) {
	writer := &recordingNativePortWriter{active: true}
	list := PortWhiteList{
		Configured: []PortWhiteListEntry{
			{Port: "80", Protocol: "tcp"},
			{Port: "80", Protocol: "tcp"},
		},
		Required: []PortWhiteListEntry{
			{Port: "22", Protocol: "tcp"},
			{Port: "80", Protocol: "tcp"},
		},
	}

	if err := addNativePortWhiteList(writer, list); err != nil {
		t.Fatal(err)
	}
	want := []string{"add 80/tcp", "add 22/tcp", "reload"}
	if !reflect.DeepEqual(writer.events, want) {
		t.Fatalf("got %#v want %#v", writer.events, want)
	}
}

func TestSyncNativePortWhiteListSequence(t *testing.T) {
	writer := &recordingNativePortWriter{active: true}
	list := PortWhiteList{
		Configured: []PortWhiteListEntry{{Port: "443", Protocol: "tcp"}},
		Required:   []PortWhiteListEntry{{Port: "22", Protocol: "tcp"}},
		Previous:   []PortWhiteListEntry{{Port: "80", Protocol: "tcp"}},
	}

	if err := syncNativePortWhiteList(writer, list); err != nil {
		t.Fatal(err)
	}
	want := []string{"status", "remove 80/tcp", "add 443/tcp", "reload"}
	if !reflect.DeepEqual(writer.events, want) {
		t.Fatalf("got %#v want %#v", writer.events, want)
	}
}

func TestSyncNativePortWhiteListInactiveIsNoop(t *testing.T) {
	writer := &recordingNativePortWriter{}
	list := PortWhiteList{
		Configured: []PortWhiteListEntry{{Port: "443", Protocol: "tcp"}},
		Previous:   []PortWhiteListEntry{{Port: "80", Protocol: "tcp"}},
	}

	if err := syncNativePortWhiteList(writer, list); err != nil {
		t.Fatal(err)
	}
	want := []string{"status"}
	if !reflect.DeepEqual(writer.events, want) {
		t.Fatalf("got %#v want %#v", writer.events, want)
	}
}

func TestSyncNativePortWhiteListStopsOnError(t *testing.T) {
	writer := &recordingNativePortWriter{active: true, failOn: "remove 80/tcp"}
	list := PortWhiteList{
		Configured: []PortWhiteListEntry{{Port: "443", Protocol: "tcp"}},
		Previous:   []PortWhiteListEntry{{Port: "80", Protocol: "tcp"}},
	}

	if err := syncNativePortWhiteList(writer, list); err == nil {
		t.Fatal("expected remove error")
	}
	want := []string{"status", "remove 80/tcp"}
	if !reflect.DeepEqual(writer.events, want) {
		t.Fatalf("got %#v want %#v", writer.events, want)
	}
}
