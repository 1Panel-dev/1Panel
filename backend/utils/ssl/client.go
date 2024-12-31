package ssl

import (
	"crypto"
	"encoding/json"
	"github.com/go-acme/lego/v4/providers/dns/rainyun"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/clouddns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/providers/dns/godaddy"
	"github.com/go-acme/lego/v4/providers/dns/huaweicloud"
	"github.com/go-acme/lego/v4/providers/dns/namecheap"
	"github.com/go-acme/lego/v4/providers/dns/namedotcom"
	"github.com/go-acme/lego/v4/providers/dns/namesilo"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/providers/dns/volcengine"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"
	"github.com/pkg/errors"
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

func NewAcmeClient(acmeAccount *model.WebsiteAcmeAccount) (*AcmeClient, error) {
	if acmeAccount.Email == "" {
		return nil, errors.New("email can not blank")
	}
	client, err := NewRegisterClient(acmeAccount)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type DnsType string

const (
	DnsPod       DnsType = "DnsPod"
	AliYun       DnsType = "AliYun"
	Volcengine   DnsType = "Volcengine"
	CloudFlare   DnsType = "CloudFlare"
	CloudDns     DnsType = "CloudDns"
	NameSilo     DnsType = "NameSilo"
	NameCheap    DnsType = "NameCheap"
	NameCom      DnsType = "NameCom"
	Godaddy      DnsType = "Godaddy"
	TencentCloud DnsType = "TencentCloud"
	HuaweiCloud  DnsType = "HuaweiCloud"
	RainYun      DnsType = "RainYun"
)

type DNSParam struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Email     string `json:"email"`
	APIkey    string `json:"apiKey"`
	APIUser   string `json:"apiUser"`
	APISecret string `json:"apiSecret"`
	SecretID  string `json:"secretID"`
	Region    string `json:"region"`
	ClientID  string `json:"clientID"`
	Password  string `json:"password"`
}

var (
	propagationTimeout = 30 * time.Minute
	pollingInterval    = 10 * time.Second
	ttl                = 3600
)

func (c *AcmeClient) UseDns(dnsType DnsType, params string, websiteSSL model.WebsiteSSL) error {
	var (
		param DNSParam
		p     challenge.Provider
		err   error
	)

	if err = json.Unmarshal([]byte(params), &param); err != nil {
		return err
	}

	switch dnsType {
	case DnsPod:
		dnsPodConfig := dnspod.NewDefaultConfig()
		dnsPodConfig.LoginToken = param.ID + "," + param.Token
		dnsPodConfig.PropagationTimeout = propagationTimeout
		dnsPodConfig.PollingInterval = pollingInterval
		dnsPodConfig.TTL = ttl
		p, err = dnspod.NewDNSProviderConfig(dnsPodConfig)
	case AliYun:
		alidnsConfig := alidns.NewDefaultConfig()
		alidnsConfig.SecretKey = param.SecretKey
		alidnsConfig.APIKey = param.AccessKey
		alidnsConfig.PropagationTimeout = propagationTimeout
		alidnsConfig.PollingInterval = pollingInterval
		alidnsConfig.TTL = ttl
		p, err = alidns.NewDNSProviderConfig(alidnsConfig)
	case Volcengine:
		volcConfig := volcengine.NewDefaultConfig()
		volcConfig.SecretKey = param.SecretKey
		volcConfig.AccessKey = param.AccessKey
		volcConfig.PropagationTimeout = propagationTimeout
		volcConfig.PollingInterval = pollingInterval
		volcConfig.TTL = ttl
		p, err = volcengine.NewDNSProviderConfig(volcConfig)
	case CloudFlare:
		cloudflareConfig := cloudflare.NewDefaultConfig()
		cloudflareConfig.AuthEmail = param.Email
		cloudflareConfig.AuthToken = param.APIkey
		cloudflareConfig.PropagationTimeout = propagationTimeout
		cloudflareConfig.PollingInterval = pollingInterval
		cloudflareConfig.TTL = 3600
		p, err = cloudflare.NewDNSProviderConfig(cloudflareConfig)
	case CloudDns:
		clouddnsConfig := clouddns.NewDefaultConfig()
		clouddnsConfig.ClientID = param.ClientID
		clouddnsConfig.Email = param.Email
		clouddnsConfig.Password = param.Password
		clouddnsConfig.PropagationTimeout = propagationTimeout
		clouddnsConfig.PollingInterval = pollingInterval
		clouddnsConfig.TTL = ttl
		p, err = clouddns.NewDNSProviderConfig(clouddnsConfig)
	case NameCheap:
		namecheapConfig := namecheap.NewDefaultConfig()
		namecheapConfig.APIKey = param.APIkey
		namecheapConfig.APIUser = param.APIUser
		namecheapConfig.PropagationTimeout = propagationTimeout
		namecheapConfig.PollingInterval = pollingInterval
		namecheapConfig.TTL = ttl
		p, err = namecheap.NewDNSProviderConfig(namecheapConfig)
	case NameSilo:
		nameSiloConfig := namesilo.NewDefaultConfig()
		nameSiloConfig.APIKey = param.APIkey
		nameSiloConfig.PropagationTimeout = propagationTimeout
		nameSiloConfig.PollingInterval = pollingInterval
		nameSiloConfig.TTL = ttl
		p, err = namesilo.NewDNSProviderConfig(nameSiloConfig)
	case Godaddy:
		godaddyConfig := godaddy.NewDefaultConfig()
		godaddyConfig.APIKey = param.APIkey
		godaddyConfig.APISecret = param.APISecret
		godaddyConfig.PropagationTimeout = propagationTimeout
		godaddyConfig.PollingInterval = pollingInterval
		godaddyConfig.TTL = ttl
		p, err = godaddy.NewDNSProviderConfig(godaddyConfig)
	case NameCom:
		nameComConfig := namedotcom.NewDefaultConfig()
		nameComConfig.APIToken = param.Token
		nameComConfig.Username = param.APIUser
		nameComConfig.PropagationTimeout = propagationTimeout
		nameComConfig.PollingInterval = pollingInterval
		nameComConfig.TTL = ttl
		p, err = namedotcom.NewDNSProviderConfig(nameComConfig)
	case TencentCloud:
		tencentCloudConfig := tencentcloud.NewDefaultConfig()
		tencentCloudConfig.SecretID = param.SecretID
		tencentCloudConfig.SecretKey = param.SecretKey
		tencentCloudConfig.PropagationTimeout = propagationTimeout
		tencentCloudConfig.PollingInterval = pollingInterval
		tencentCloudConfig.TTL = ttl
		p, err = tencentcloud.NewDNSProviderConfig(tencentCloudConfig)
	case HuaweiCloud:
		huaweiCloudConfig := huaweicloud.NewDefaultConfig()
		huaweiCloudConfig.AccessKeyID = param.AccessKey
		huaweiCloudConfig.SecretAccessKey = param.SecretKey
		huaweiCloudConfig.Region = param.Region
		huaweiCloudConfig.PropagationTimeout = propagationTimeout
		huaweiCloudConfig.PollingInterval = pollingInterval
		huaweiCloudConfig.TTL = int32(ttl)
		p, err = huaweicloud.NewDNSProviderConfig(huaweiCloudConfig)
	case RainYun:
		rainyunConfig := rainyun.NewDefaultConfig()
		rainyunConfig.APIKey = param.APIkey
		rainyunConfig.PropagationTimeout = propagationTimeout
		rainyunConfig.PollingInterval = pollingInterval
		rainyunConfig.TTL = ttl
		p, err = rainyun.NewDNSProviderConfig(rainyunConfig)
	}
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
		dns01.AddDNSTimeout(10*time.Minute),
	)
}

func (c *AcmeClient) UseManualDns() error {
	p := &manualDnsProvider{}
	if err := c.Client.Challenge.SetDNS01Provider(p, dns01.AddDNSTimeout(10*time.Minute)); err != nil {
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
		challengeInfo := dns01.GetChallengeInfo(domain, keyAuth)
		fqdn := challengeInfo.FQDN
		if strings.HasPrefix(domain, "*.") && strings.Contains(fqdn, "*.") {
			fqdn = strings.Replace(fqdn, "*.", "", 1)
		}
		resolves[domain] = Resolve{
			Key:   fqdn,
			Value: challengeInfo.Value,
		}
	}

	return resolves, nil
}
