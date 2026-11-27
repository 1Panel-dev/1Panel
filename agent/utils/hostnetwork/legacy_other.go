//go:build !linux

package hostnetwork

import "errors"

func readLegacyLinks(string) ([]Link, error) {
	return nil, &IPRoute2Error{Err: errors.New("iproute2 JSON output or Linux netlink is required")}
}

func readLegacyRoutes(string) ([]Route, error) {
	return nil, &IPRoute2Error{Err: errors.New("iproute2 JSON output or Linux netlink is required")}
}
