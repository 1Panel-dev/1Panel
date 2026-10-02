package cmd

import (
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
)

// holdPipeScript leaves a grandchild that survives the kill of the process
// group and keeps the output pipes open, so exec.Cmd.Wait can't return until
// it exits. It stands in for a process stuck in an uninterruptible syscall
// (statfs on a stale NFS mount), which can't be reaped after SIGKILL either.
const (
	holdPipeScript   = "setsid sleep 3 & sleep 3"
	holdPipeDuration = 3 * time.Second
)

func requireCommands(t *testing.T, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s is not available: %v", name, err)
		}
	}
}

func shortenKillWait(t *testing.T) {
	t.Helper()
	old := killWaitTimeout
	killWaitTimeout = 200 * time.Millisecond
	t.Cleanup(func() { killWaitTimeout = old })
}

// assertCmdTimeout avoids err.Error(), which needs the i18n bundle.
func assertCmdTimeout(t *testing.T, err error) {
	t.Helper()
	var busErr buserr.BusinessError
	if !errors.As(err, &busErr) || busErr.Msg != "ErrCmdTimeout" {
		t.Fatalf("err = %#v, want ErrCmdTimeout", err)
	}
}

func TestRunWithStdout(t *testing.T) {
	requireCommands(t, "sh")
	stdout, err := NewCommandMgr(WithTimeout(5*time.Second)).RunWithStdout("sh", "-c", "echo hello")
	if err != nil || stdout != "hello\n" {
		t.Fatalf("RunWithStdout() = %q, %#v, want hello", stdout, err)
	}
}

func TestRunPipe(t *testing.T) {
	requireCommands(t, "sh", "tr")
	stdout, err := NewCommandMgr(WithTimeout(5*time.Second)).RunPipe(
		PipeCommand{Name: "sh", Args: []string{"-c", "echo hello"}},
		PipeCommand{Name: "tr", Args: []string{"a-z", "A-Z"}},
	)
	if err != nil || stdout != "HELLO\n" {
		t.Fatalf("RunPipe() = %q, %#v, want HELLO", stdout, err)
	}
}

func TestRunTimeoutKillsCommand(t *testing.T) {
	requireCommands(t, "sleep")
	start := time.Now()
	_, err := NewCommandMgr(WithTimeout(300*time.Millisecond)).RunWithStdout("sleep", "5")
	assertCmdTimeout(t, err)
	// A killed process is reaped at once; the kill wait must not be added on top.
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("timed-out command returned after %s, want right after the 300ms timeout", elapsed)
	}
}

func TestRunTimeoutDoesNotWaitForHeldPipe(t *testing.T) {
	requireCommands(t, "sh", "setsid", "sleep")
	shortenKillWait(t)
	start := time.Now()
	_, err := NewCommandMgr(WithTimeout(500*time.Millisecond)).RunWithStdout("sh", "-c", holdPipeScript)
	assertCmdTimeout(t, err)
	if elapsed := time.Since(start); elapsed > holdPipeDuration-time.Second {
		t.Fatalf("timed-out command returned after %s, want it not to wait for the pipe holder", elapsed)
	}
}

// After a timed-out run has returned, output the pipe holder still produces
// must not reach the writer the caller passed in, which the caller now reads.
func TestRunTimeoutDetachesCallerWriter(t *testing.T) {
	requireCommands(t, "sh", "setsid", "sleep")
	shortenKillWait(t)
	var stderr lockedBuffer
	script := `echo early >&2; setsid sh -c "sleep 1; echo late >&2" & sleep 3`
	_, err := NewCommandMgr(WithTimeout(500*time.Millisecond), WithStderr(&stderr)).RunWithStdout("sh", "-c", script)
	assertCmdTimeout(t, err)
	if got := stderr.String(); got != "early\n" {
		t.Fatalf("stderr at return = %q, want only the output written before the timeout", got)
	}
	time.Sleep(1200 * time.Millisecond)
	if got := stderr.String(); got != "early\n" {
		t.Fatalf("stderr after the pipe holder wrote again = %q, want it unchanged", got)
	}
}

func TestRunPipeTimeoutDoesNotWaitForHeldPipe(t *testing.T) {
	requireCommands(t, "sh", "setsid", "sleep", "cat")
	shortenKillWait(t)
	start := time.Now()
	_, err := NewCommandMgr(WithTimeout(500*time.Millisecond)).RunPipe(
		PipeCommand{Name: "sh", Args: []string{"-c", holdPipeScript}},
		PipeCommand{Name: "cat"},
	)
	assertCmdTimeout(t, err)
	if elapsed := time.Since(start); elapsed > holdPipeDuration-time.Second {
		t.Fatalf("timed-out pipe returned after %s, want it not to wait for the pipe holder", elapsed)
	}
}
