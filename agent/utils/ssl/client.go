package ssl

import (
	"crypto"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/http/webroot"
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

	return c.Client.Challenge.SetDNS01Provider(p,
		dns01.CondOption(len(nameservers) > 0,
			dns01.AddRecursiveNameservers(nameservers)),
		dns01.CondOption(websiteSSL.SkipDNS,
			dns01.DisableAuthoritativeNssPropagationRequirement()),
		dns01.AddDNSTimeout(dnsTimeOut),
	)
}

func (c *AcmeClient) UseManualDns(websiteSSL model.WebsiteSSL) error {
	p, err := NewCustomDNSProviderManual(&ManualConfig{
		PropagationTimeout: 20 * time.Minute,
		PollingInterval:    pollingInterval,
		TTL:                ttl,
	})
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
	if err = c.Client.Challenge.SetDNS01Provider(p,
		dns01.CondOption(len(nameservers) > 0,
			dns01.AddRecursiveNameservers(nameservers)),
		dns01.CondOption(websiteSSL.SkipDNS,
			dns01.DisableAuthoritativeNssPropagationRequirement()),
		dns01.AddDNSTimeout(dnsTimeOut)); err != nil {
		return err
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
	return nil
}

func (c *AcmeClient) ObtainSSL(domains []string, privateKey crypto.PrivateKey) (certificate.Resource, error) {
	request := certificate.ObtainRequest{
		Domains:    domains,
		Bundle:     true,
		PrivateKey: privateKey,
	}

	certificates, err := c.Client.Certificate.Obtain(request)
	if err != nil {
		return certificate.Resource{}, err
	}

	return *certificates, nil
}

func (c *AcmeClient) RevokeSSL(pemSSL []byte) error {
	return c.Client.Certificate.Revoke(pemSSL)
}

type Resolve struct {
	Key   string
	Value string
	Err   string
}

func (c *AcmeClient) GetDNSResolve(domains []string) (map[string]Resolve, error) {
	core, err := api.New(c.Config.HTTPClient, c.Config.UserAgent, c.Config.CADirURL, c.User.Registration.URI, c.User.Key)
	if err != nil {
		return nil, err
	}
	order, err := core.Orders.New(domains)
	if err != nil {
		return nil, err
	}
	resolves := make(map[string]Resolve)
	resc, errc := make(chan acme.Authorization), make(chan domainError)
	for _, authzURL := range order.Authorizations {
		go func(authzURL string) {
			authz, err := core.Authorizations.Get(authzURL)
			if err != nil {
				errc <- domainError{Domain: authz.Identifier.Value, Error: err}
				return
			}
			resc <- authz
		}(authzURL)
	}

	var responses []acme.Authorization
	for i := 0; i < len(order.Authorizations); i++ {
		select {
		case res := <-resc:
			responses = append(responses, res)
		case err := <-errc:
			resolves[err.Domain] = Resolve{Err: err.Error.Error()}
		}
	}
	close(resc)
	close(errc)

	for _, auth := range responses {
		domain := challenge.GetTargetedDomain(auth)
		chlng, err := challenge.FindChallenge(challenge.DNS01, auth)
		if err != nil {
			resolves[domain] = Resolve{Err: err.Error()}
			continue
		}
		keyAuth, err := core.GetKeyAuthorization(chlng.Token)
		if err != nil {
			resolves[domain] = Resolve{Err: err.Error()}
			continue
		}
		challengeInfo := dns01.GetChallengeInfo(domain, keyAuth)
		fqdn := challengeInfo.FQDN
		if strings.HasPrefix(domain, "*.") && strings.Contains(fqdn, "*.") {
			fqdn = strings.Replace(fqdn, "*.", "", 1)
		}
		_, _ = dns01.FindZoneByFqdn(challengeInfo.EffectiveFQDN)
		resolves[domain] = Resolve{
			Key:   fqdn,
			Value: challengeInfo.Value,
		}
	}

	return resolves, nil
}
