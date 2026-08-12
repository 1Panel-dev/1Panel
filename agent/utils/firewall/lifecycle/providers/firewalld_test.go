package providers

import (
	"errors"
	"testing"
)

func TestFirewalldStoppedRecognizesNormalInactiveResult(t *testing.T) {
	for _, test := range []struct {
		stdout string
		err    error
	}{
		{stdout: "not running\n"},
		{err: errors.New("stderr: FirewallD is not running, exit status 252")},
	} {
		if !firewalldStopped(test.stdout, test.err) {
			t.Fatalf("expected stopped result for stdout=%q err=%v", test.stdout, test.err)
		}
	}
}

func TestFirewalldStoppedKeepsUnexpectedFailures(t *testing.T) {
	if firewalldStopped("", errors.New("permission denied")) {
		t.Fatal("unexpected command failures must not be treated as an inactive firewall")
	}
}
