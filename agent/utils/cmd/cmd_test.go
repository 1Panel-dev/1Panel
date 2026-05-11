package cmd

import "testing"

// TestWhich_ExistingBinary verifies that Which() returns true for a
// binary that is guaranteed to exist on every Unix-like build host
// (`sh`). Regression test for #12605: the previous implementation
// shelled out to `which`, which is not always available on minimal
// distributions like Arch Linux. The new implementation tries
// exec.LookPath first, so this assertion holds regardless of whether
// `which` itself is on PATH.
func TestWhich_ExistingBinary(t *testing.T) {
	if !Which("sh") {
		t.Errorf("Which(\"sh\") = false, want true")
	}
}

func TestWhich_MissingBinary(t *testing.T) {
	// A binary name that is extremely unlikely to exist on any host.
	if Which("definitely-not-a-real-binary-xyzzy-1panel") {
		t.Errorf("Which(\"definitely-not-a-real-binary-xyzzy-1panel\") = true, want false")
	}
}
