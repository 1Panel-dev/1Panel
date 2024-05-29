//go:build !xpack

package xpack

import "net/http"

func RemoveTamper(website string) {}

func LoadRequestTransport() (bool, *http.Transport) {
	return false, nil
}

func LoadGpuInfo() []interface{} {
	return nil
}
