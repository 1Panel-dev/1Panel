package cmd

import (
	"testing"

	"github.com/1Panel-dev/1Panel/core/constant"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		complexity string
		want       passwordValidation
	}{
		{
			name:       "complexity enabled rejects digits only",
			password:   "123456",
			complexity: constant.StatusEnable,
			want:       passwordTooWeak,
		},
		{
			name:       "complexity enabled rejects letters only",
			password:   "abcdefgh",
			complexity: constant.StatusEnable,
			want:       passwordTooWeak,
		},
		{
			name:       "complexity enabled rejects short mixed password",
			password:   "Ab1!",
			complexity: constant.StatusEnable,
			want:       passwordTooWeak,
		},
		{
			name:       "complexity enabled accepts letters and digits",
			password:   "abcd1234",
			complexity: constant.StatusEnable,
			want:       passwordValid,
		},
		{
			name:       "complexity enabled accepts letters digits and special",
			password:   "Abc123!x",
			complexity: constant.StatusEnable,
			want:       passwordValid,
		},
		{
			name:       "complexity disabled accepts six digits",
			password:   "123456",
			complexity: constant.StatusDisable,
			want:       passwordValid,
		},
		{
			name:       "complexity disabled accepts six letters",
			password:   "abcdef",
			complexity: constant.StatusDisable,
			want:       passwordValid,
		},
		{
			name:       "complexity disabled rejects five characters",
			password:   "12345",
			complexity: constant.StatusDisable,
			want:       passwordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validatePassword(tt.password, tt.complexity); got != tt.want {
				t.Fatalf("validatePassword(%q, %q) = %v, want %v", tt.password, tt.complexity, got, tt.want)
			}
		})
	}
}
