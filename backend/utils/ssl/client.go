package ssl

import (
	"crypto"
	"encoding/json"
	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"
	"github.com/pkg/errors"
	"time"
)

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

type AcmeClient struct {
	Config *lego.Config
	Client *lego.Client
	User   *AcmeUser
}

func NewAcmeClient(email, privateKey string) (*AcmeClient, error) {
	if email == "" {
		return nil, errors.New("email can not blank")
	}
	if privateKey == "" {
		client, err := NewRegisterClient(email)
		if err != nil {
			return nil, err
		}
		return client, nil
	} else {
		client, err := NewPrivateKeyClient(email, privateKey)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}

type DnsType string

const (
	DnsPod     DnsType = "DnsPod"
	AliYun     DnsType = "AliYun"
	CloudFlare DnsType = "CloudFlare"
)

type DNSParam struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Email     string `json:"email"`
	APIkey    string `json:"APIkey"`
}

func (c *AcmeClient) UseDns(dnsType DnsType, params string) error {
	var param DNSParam
	if err := json.Unmarshal([]byte(params), &param); err != nil {
		return err
	}

	var p challenge.Provider
	var err error
	if dnsType == DnsPod {
		dnsPodConfig := dnspod.NewDefaultConfig()
		dnsPodConfig.LoginToken = param.ID + "," + param.Token
		p, err = dnspod.NewDNSProviderConfig(dnsPodConfig)
		if err != nil {
			return err
		}
	}
	if dnsType == AliYun {
		alidnsConfig := alidns.NewDefaultConfig()
		alidnsConfig.SecretKey = param.SecretKey
		alidnsConfig.APIKey = param.AccessKey
		p, err = alidns.NewDNSProviderConfig(alidnsConfig)
		if err != nil {
			return err
		}
	}
	if dnsType == CloudFlare {
		cloudflareConfig := cloudflare.NewDefaultConfig()
		cloudflareConfig.AuthEmail = param.Email
		cloudflareConfig.AuthKey = param.APIkey
		p, err = cloudflare.NewDNSProviderConfig(cloudflareConfig)
		if err != nil {
			return err
		}
	}

	return c.Client.Challenge.SetDNS01Provider(p, dns01.AddDNSTimeout(3*time.Minute))
}

func (c *AcmeClient) UseManualDns() error {
	p := &manualDnsProvider{}
	if err := c.Client.Challenge.SetDNS01Provider(p, dns01.AddDNSTimeout(3*time.Minute)); err != nil {
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

func (c *AcmeClient) ObtainSSL(domains []string) (certificate.Resource, error) {
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	certificates, err := c.Client.Certificate.Obtain(request)
	if err != nil {
		return certificate.Resource{}, err
	}

	return *certificates, nil
}

func (c *AcmeClient) RenewSSL(certUrl string) (certificate.Resource, error) {
	certificates, err := c.Client.Certificate.Get(certUrl, true)
	if err != nil {
		return certificate.Resource{}, err
	}
	certificates, err = c.Client.Certificate.Renew(*certificates, true, true, "")
	if err != nil {
		return certificate.Resource{}, err
	}

	return *certificates, nil
}

type Resolve struct {
	Key   string
	Value string
	Err   string
}

type manualDnsProvider struct {
	Resolve *Resolve
}

func (p *manualDnsProvider) Present(domain, token, keyAuth string) error {
	return nil
}

func (p *manualDnsProvider) CleanUp(domain, token, keyAuth string) error {
	return nil
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
		fqdn, value := dns01.GetRecord(domain, keyAuth)
		resolves[domain] = Resolve{
			Key:   fqdn,
			Value: value,
		}
	}

	return resolves, nil
}
