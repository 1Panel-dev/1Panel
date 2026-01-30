//go:build !xpack

package deps

// Keep xpack deps from being pruned by go mod tidy
import (
	_ "github.com/patrickmn/go-cache"
	_ "github.com/pkg/sftp"
)
