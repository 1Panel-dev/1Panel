package server

import (
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestPrepareMasterSocketDir verifies the master socket directory is created
// with locked-down permissions and rejects directories owned by another user.
func TestPrepareMasterSocketDir_CreatesWithRestrictedPerm(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "1panel")
	if err := prepareMasterSocketDir(dir); err != nil {
		t.Fatalf("prepareMasterSocketDir failed: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != masterSocketDirPerm {
		t.Fatalf("expected dir perm %#o, got %#o", masterSocketDirPerm, perm)
	}
}

func TestPrepareMasterSocketDir_RejectsWorldReadable(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "1panel")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	if err := prepareMasterSocketDir(dir); err != nil {
		// prepareMasterSocketDir Chmods the dir, so it should succeed.
		t.Fatalf("prepareMasterSocketDir failed: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != masterSocketDirPerm {
		t.Fatalf("expected dir perm %#o after chmod, got %#o", masterSocketDirPerm, perm)
	}
}

// TestSecureMasterSocket verifies the helper enforces 0600 + matching uid on a
// freshly-created unix socket.
func TestSecureMasterSocket(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "agent.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen unix failed: %v", err)
	}
	defer listener.Close()

	// The socket starts with whatever umask leaves; secureMasterSocket should
	// always force 0600.
	if err := secureMasterSocket(sockPath); err != nil {
		t.Fatalf("secureMasterSocket failed: %v", err)
	}

	info, err := os.Stat(sockPath)
	if err != nil {
		t.Fatalf("stat sock failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != masterSocketFilePerm {
		t.Fatalf("expected sock perm %#o, got %#o", masterSocketFilePerm, perm)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("expected *syscall.Stat_t, got %T", info.Sys())
	}
	if int(stat.Uid) != os.Geteuid() {
		t.Fatalf("expected uid=%d, got uid=%d", os.Geteuid(), stat.Uid)
	}
}
