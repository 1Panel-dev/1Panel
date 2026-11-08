package req_helper

import (
	"crypto/tls"
	"errors"
	"net/http"
	"time"
)

func newExternalHTTPClient(transport *http.Transport, timeout time.Duration) *http.Client {
	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport)
	}
	verified := transport.Clone()
	if verified.TLSClientConfig == nil {
		verified.TLSClientConfig = &tls.Config{}
	} else {
		verified.TLSClientConfig = verified.TLSClientConfig.Clone()
	}
	verified.TLSClientConfig.InsecureSkipVerify = false
	verified.TLSClientConfig.ServerName = ""
	verified.DialTLS = nil
	verified.DialTLSContext = nil
	return &http.Client{
		Timeout:       timeout,
		Transport:     verified,
		CheckRedirect: checkExternalRedirect,
	}
}

func checkExternalRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if len(via) > 0 && via[len(via)-1].URL.Scheme == "https" && req.URL.Scheme != "https" {
		return errors.New("refusing HTTPS redirect to an insecure scheme")
	}
	return nil
}
