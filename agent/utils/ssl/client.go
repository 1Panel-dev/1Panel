package ssl

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"os"
	"sync"
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

// dnsChallengeMu serializes DNS-01 issuance flows because lego v5 keeps the
// recursive-nameserver Client and the LEGO_DISABLE_CNAME_SUPPORT switch in
// process-wide globals. Without this lock, two concurrent SSL applications
// with different nameserver/CNAME settings would clobber each other and
// occasionally fail propagation checks against the wrong resolver.
var dnsChallengeMu sync.Mutex

type AcmeClientOption func(*AcmeClientOptions)

type AcmeClientOptions struct {
	SystemProxy *dto.SystemProxy
}

type AcmeClient struct {
	Config   *lego.Config
	Client   *lego.Client
	User     *AcmeUser
	ProxyURL string

	// dnsChallengeLocked records whether this client currently holds
	// dnsChallengeMu. It is set by UseDns and cleared by ObtainSSL/
	// ObtainIPSSL once the DNS-01 flow finishes. UseHTTP does not touch it.
	dnsChallengeLocked bool
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

	// Hold the global DNS-01 lock for the entire flow that follows, including
	// the Obtain call. lego v5 reads dns01.DefaultClient() inside its own
	// propagation-precheck loop, so the lock cannot be released right after
	// SetDefaultClient/SetDNS01Provider; it has to span the whole DNS-01
	// challenge. ObtainSSL / ObtainIPSSL release it once the request returns.
	dnsChallengeMu.Lock()
	c.dnsChallengeLocked = true

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

	if err := c.Client.Challenge.SetDNS01Provider(p, opts...); err != nil {
		c.releaseDNSLock()
		return err
	}
	return nil
}

// releaseDNSLock releases dnsChallengeMu if it was acquired by UseDns.
// Safe to call multiple times; calls after the first one are no-ops.
func (c *AcmeClient) releaseDNSLock() {
	if c.dnsChallengeLocked {
		c.dnsChallengeLocked = false
		dnsChallengeMu.Unlock()
	}
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
	defer c.releaseDNSLock()
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
	defer c.releaseDNSLock()
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
