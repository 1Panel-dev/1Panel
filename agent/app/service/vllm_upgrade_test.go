package service

import "testing"

func TestVllmUpgradeCandidateAppStoreVersions(t *testing.T) {
	cases := []struct {
		current, target, image string
		want                   bool
	}{
		// Intel image build numbers change with every release.
		{"intel-0.14.0-b8.3.1", "intel-0.26.0-b2", "intel/llm-scaler-vllm:0.14.0-b8.3.1", true},
		{"intel-0.26.0-b2", "intel-0.26.0-b3", "intel/llm-scaler-vllm:0.26.0-b2", true},
		{"intel-0.26.0-b2", "intel-0.14.0-b8.3.1", "intel/llm-scaler-vllm:0.26.0-b2", false},
		// Hardware and model variants must stay on their own line.
		{"ascend-0.23.0-310p", "ascend-0.30.0-310p", "", true},
		{"ascend-0.30.0-310p", "ascend-0.30.0-a3", "", false},
		{"ascend-0.23.0", "ascend-0.30.0-a3", "", false},
		{"nvidia-0.32.0", "nvidia-0.33.0", "", true},
		{"nvidia-0.25-optimized-for-2-gb10-ds-v4-flash", "nvidia-0.33.0", "", false},
		{"intel-0.26.0-b2", "nvidia-0.33.0", "", false},
	}
	for _, tc := range cases {
		if got := isVllmUpgradeCandidate(tc.current, tc.target, tc.image); got != tc.want {
			t.Errorf("isVllmUpgradeCandidate(%q -> %q) = %v, want %v", tc.current, tc.target, got, tc.want)
		}
	}
}
