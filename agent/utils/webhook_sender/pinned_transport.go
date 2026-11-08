package webhook_sender

import (
	"errors"
	"io"
	"net"
	"net/http"
)

func RoundTripPinned(request *http.Request, addresses []net.IP, input *http.Transport) (*http.Response, error) {
	if request == nil || request.URL == nil || len(addresses) == 0 || addresses[0] == nil {
		return nil, errors.New("invalid pinned HTTP target")
	}
	originalURL := cloneURL(request.URL)
	pinned := request.Clone(request.Context())
	pinned.URL = cloneURL(originalURL)
	pinned.URL.Host = pinnedHost(addresses[0], originalURL.Port())
	pinned.Host = originalURL.Host
	transport := systemTLSTransport(input, originalURL.Hostname(), originalURL, addresses)
	transport.DisableKeepAlives = true
	var roundTripper http.RoundTripper = transport
	if originalURL.Scheme == "http" {
		roundTripper = &pinnedHTTPProxyRoundTripper{transport: transport, originalHost: originalURL.Host}
	}
	response, err := roundTripper.RoundTrip(pinned)
	if err != nil {
		transport.CloseIdleConnections()
		return nil, err
	}
	response.Request = request
	response.Body = &pinnedResponseBody{ReadCloser: response.Body, transport: transport}
	return response, nil
}

type pinnedResponseBody struct {
	io.ReadCloser
	transport *http.Transport
}

func (body *pinnedResponseBody) Close() error {
	err := body.ReadCloser.Close()
	body.transport.CloseIdleConnections()
	return err
}
