package ssl

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/providers/http/webroot"
	"github.com/pkg/errors"
)

type AcmeClientOption func(*AcmeClientOptions)

type AcmeClientOptions struct {
	SystemProxy *dto.SystemProxy
}

type AcmeClient struct {
	Config   *lego.Config
	Client   *lego.Client
	User     *AcmeUser
	ProxyURL string
}

func NewAcmeClient(acmeAccount *model.WebsiteAcmeAccount, systemProxy *dto.SystemProxy) (*AcmeClient, error) {
	if acmeAccount.Email == "" {
		return nil, errors.New("email can not blank")
	}

	client, err := NewRegisterClient(acmeAccount, systemProxy)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *AcmeClient) UseDns(dnsType DnsType, params string, websiteSSL model.WebsiteSSL) error {
	p, err := getDNSProviderConfig(dnsType, params)
	if err != nil {
		return err
	}
	var nameservers []string
	if websiteSSL.Nameserver1 != "" {
		nameservers = append(nameservers, websiteSSL.Nameserver1)
	}
	if websiteSSL.Nameserver2 != "" {
		nameservers = append(nameservers, websiteSSL.Nameserver2)
	}
	if websiteSSL.DisableCNAME {
		_ = os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "true")
	} else {
		_ = os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "false")
	}

	// lego v5 removed dns01.AddRecursiveNameservers and dns01.AddDNSTimeout;
	// configure them via dns01.NewClient(&dns01.Options{...}) + SetDefaultClient.
	dns01.SetDefaultClient(dns01.NewClient(&dns01.Options{
		RecursiveNameservers: nameservers,
		Timeout:              dnsTimeOut,
	}))

	var opts []dns01.ChallengeOption
	if websiteSSL.SkipDNS {
		opts = append(opts, dns01.DisableAuthoritativeNssPropagationRequirement())
	}

	return c.Client.Challenge.SetDNS01Provider(p, opts...)
}

func (c *AcmeClient) UseHTTP(path string) error {
	httpProvider, err := webroot.NewHTTPProvider(path)
	if err != nil {
		return err
	}

	err = c.Client.Challenge.SetHTTP01Provider(httpProvider)
	if err != nil {
		return err
	}
	return nil
}

func (c *AcmeClient) ObtainSSL(domains []string, privateKey crypto.Signer) (certificate.Resource, error) {
	// lego v5 disables Common Name by default; explicitly enable it to keep
	// the v4 behaviour, so legacy Java/router clients that still rely on the
	// CommonName field do not fail TLS handshake.
	request := certificate.ObtainRequest{
		Domains:          domains,
		Bundle:           true,
		PrivateKey:       privateKey,
		EnableCommonName: true,
	}

	ctx := context.Background()

	var certificates *certificate.Resource
	var err error

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		certificates, err = c.Client.Certificate.Obtain(ctx, request)
		if err == nil {
			return *certificates, nil
		}

		if isHTTP503Error(err) && attempt < maxRetryAttempts {
			global.LOG.Warnf("ACME server returned 503, retrying in %v (attempt %d/%d)",
				retryDelayOn503, attempt, maxRetryAttempts)
			time.Sleep(retryDelayOn503)
			continue
		}

		// Non-503 error or final attempt, return error
		return certificate.Resource{}, err
	}

	return certificate.Resource{}, err
}

func (c *AcmeClient) ObtainIPSSL(ipAddress string, privKey crypto.Signer) (certificate.Resource, error) {
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "",
		},
		IPAddresses: []net.IP{
			net.ParseIP(ipAddress),
		},
	}
	csrDER, err := x509.CreateCertificateRequest(
		rand.Reader,
		csrTemplate,
		privKey,
	)
	if err != nil {
		return certificate.Resource{}, err
	}
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return certificate.Resource{}, err
	}
	req := certificate.ObtainForCSRRequest{
		CSR:        csr,
		PrivateKey: privKey,
		Profile:    "shortlived",
		Bundle:     true,
	}

	ctx := context.Background()

	var certificates *certificate.Resource
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		certificates, err = c.Client.Certificate.ObtainForCSR(ctx, req)
		if err == nil {
			return *certificates, nil
		}

		if isHTTP503Error(err) && attempt < maxRetryAttempts {
			global.LOG.Warnf("ACME server returned 503 for IP SSL, retrying in %v (attempt %d/%d)",
				retryDelayOn503, attempt, maxRetryAttempts)
			time.Sleep(retryDelayOn503)
			continue
		}

		return certificate.Resource{}, err
	}

	return certificate.Resource{}, err
}

func (c *AcmeClient) RevokeSSL(pemSSL []byte) error {
	return c.Client.Certificate.Revoke(context.Background(), pemSSL)
}
