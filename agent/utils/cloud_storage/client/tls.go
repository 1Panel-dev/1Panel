package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func newStorageHTTPTransport(vars map[string]interface{}) (*http.Transport, error) {
	config := &tls.Config{}
	value, exists := vars["caCertPEM"]
	if !exists || value == nil {
		return &http.Transport{TLSClientConfig: config}, nil
	}
	certificatePEM, ok := value.(string)
	if !ok {
		return nil, errors.New("caCertPEM must be a PEM certificate string")
	}
	rest := bytes.TrimSpace([]byte(certificatePEM))
	if len(rest) == 0 {
		return &http.Transport{TLSClientConfig: config}, nil
	}
	roots, err := x509.SystemCertPool()
	if err != nil {
		roots = x509.NewCertPool()
	}
	for len(rest) > 0 {
		if !bytes.HasPrefix(rest, []byte("-----BEGIN CERTIFICATE-----")) {
			return nil, errors.New("caCertPEM must contain only PEM certificates")
		}
		block, remaining := pem.Decode(rest)
		if block == nil || block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			return nil, errors.New("caCertPEM contains an invalid PEM certificate")
		}
		if bytes.Count(rest[:len(rest)-len(remaining)], []byte("-----BEGIN CERTIFICATE-----")) != 1 {
			return nil, errors.New("caCertPEM contains an invalid PEM certificate")
		}
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("caCertPEM contains an invalid certificate: %w", err)
		}
		roots.AddCert(certificate)
		rest = bytes.TrimSpace(remaining)
	}
	config.RootCAs = roots
	return &http.Transport{TLSClientConfig: config}, nil
}

func newALIHTTPClient() *resty.Client {
	return resty.New().SetTLSClientConfig(&tls.Config{})
}
