package ssl

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	legoacme "github.com/go-acme/lego/v5/acme"
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

var ErrAcmeAccountURLMissing = errors.New("acme account url is empty")

type AcmeClientOption func(*AcmeClientOptions)

type AcmeClientOptions struct {
	SystemProxy *dto.SystemProxy
}

type AcmeClient struct {
	Config   *lego.Config
	Client   *lego.Client
	User     *AcmeUser
	ProxyURL string

	dnsChallengeConfig *dnsChallengeConfig
}

type dnsChallengeConfig struct {
	recursiveNameservers []string
	disableCNAME         bool
}

func NewAcmeClient(acmeAccount *model.WebsiteAcmeAccount, systemProxy *dto.SystemProxy) (*AcmeClient, error) {
	return NewAcmeClientWithContext(context.Background(), acmeAccount, systemProxy)
}

func NewAcmeClientWithContext(ctx context.Context, acmeAccount *model.WebsiteAcmeAccount, systemProxy *dto.SystemProxy) (*AcmeClient, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if acmeAccount.Email == "" {
		return nil, errors.New("email can not blank")
	}
	accountURL := strings.TrimSpace(acmeAccount.URL)
	if accountURL == "" {
		return nil, ErrAcmeAccountURLMissing
	}
	if strings.TrimSpace(acmeAccount.PrivateKey) == "" {
		return nil, errors.New("private key can not blank")
	}

	client, err := newAcmeClient(acmeAccount, systemProxy, &legoacme.ExtendedAccount{Location: accountURL})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *AcmeClient) UseDns(dnsType DnsType, params string, websiteSSL model.WebsiteSSL) error {
	var httpClient *http.Client
	if c.Config != nil {
		httpClient = c.Config.HTTPClient
	}
	p, err := getDNSProviderConfig(dnsType, params, httpClient)
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

	var opts []dns01.ChallengeOption
	if websiteSSL.SkipDNS {
		opts = append(opts, dns01.DisableAuthoritativeNssPropagationRequirement())
	}

	if err := c.Client.Challenge.SetDNS01Provider(p, opts...); err != nil {
		return err
	}
	c.dnsChallengeConfig = &dnsChallengeConfig{
		recursiveNameservers: append([]string(nil), nameservers...),
		disableCNAME:         websiteSSL.DisableCNAME,
	}
	return nil
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
	c.dnsChallengeConfig = nil
	return nil
}

func (c *AcmeClient) ObtainSSL(ctx context.Context, domains []string, privateKey crypto.Signer) (certificate.Resource, error) {
	unlockDNSChallenge := c.lockDNSChallenge()
	defer unlockDNSChallenge()

	// lego v5 disables Common Name by default; explicitly enable it to keep
	// the v4 behaviour, so legacy Java/router clients that still rely on the
	// CommonName field do not fail TLS handshake.
	request := certificate.ObtainRequest{
		Domains:          domains,
		Bundle:           true,
		PrivateKey:       privateKey,
		EnableCommonName: true,
	}

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
			if err := waitForRetry(ctx, retryDelayOn503); err != nil {
				return certificate.Resource{}, err
			}
			continue
		}

		// Non-503 error or final attempt, return error
		return certificate.Resource{}, err
	}

	return certificate.Resource{}, err
}

func (c *AcmeClient) ObtainIPSSL(ctx context.Context, ipAddress string, privKey crypto.Signer) (certificate.Resource, error) {
	unlockDNSChallenge := c.lockDNSChallenge()
	defer unlockDNSChallenge()

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

	var certificates *certificate.Resource
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		certificates, err = c.Client.Certificate.ObtainForCSR(ctx, req)
		if err == nil {
			return *certificates, nil
		}

		if isHTTP503Error(err) && attempt < maxRetryAttempts {
			global.LOG.Warnf("ACME server returned 503 for IP SSL, retrying in %v (attempt %d/%d)",
				retryDelayOn503, attempt, maxRetryAttempts)
			if err := waitForRetry(ctx, retryDelayOn503); err != nil {
				return certificate.Resource{}, err
			}
			continue
		}

		return certificate.Resource{}, err
	}

	return certificate.Resource{}, err
}

func (c *AcmeClient) RevokeSSL(pemSSL []byte) error {
	return c.Client.Certificate.Revoke(context.Background(), pemSSL)
}

func (c *AcmeClient) lockDNSChallenge() func() {
	if c.dnsChallengeConfig == nil {
		return func() {}
	}

	dnsChallengeMu.Lock()

	oldCNAME, hadCNAME := os.LookupEnv("LEGO_DISABLE_CNAME_SUPPORT")
	if c.dnsChallengeConfig.disableCNAME {
		_ = os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "true")
	} else {
		_ = os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "false")
	}

	previousClient := dns01.DefaultClient()
	// lego v5 configures custom recursive nameservers through a client instance.
	dns01.SetDefaultClient(dns01.NewClient(&dns01.Options{
		RecursiveNameservers: c.dnsChallengeConfig.recursiveNameservers,
	}))

	return func() {
		dns01.SetDefaultClient(previousClient)
		if hadCNAME {
			_ = os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", oldCNAME)
		} else {
			_ = os.Unsetenv("LEGO_DISABLE_CNAME_SUPPORT")
		}
		dnsChallengeMu.Unlock()
	}
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
