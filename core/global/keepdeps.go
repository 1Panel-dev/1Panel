package global

// Keep xpack deps from being pruned by go mod tidy in OSS builds.
import (
	_ "github.com/patrickmn/go-cache"
	_ "github.com/pkg/sftp"
)
