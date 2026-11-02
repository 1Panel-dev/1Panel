//go:build !amd64 && !arm64

package helper

func embeddedHelperBytes() ([]byte, bool) {
	return nil, false
}
