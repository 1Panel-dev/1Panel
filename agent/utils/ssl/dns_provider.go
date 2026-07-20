package ssl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/providers/dns/acmedns"
	"github.com/go-acme/lego/v5/providers/dns/alidns"
	"github.com/go-acme/lego/v5/providers/dns/aliesa"
	"github.com/go-acme/lego/v5/providers/dns/baiducloud"
	"github.com/go-acme/lego/v5/providers/dns/clouddns"
	"github.com/go-acme/lego/v5/providers/dns/cloudflare"
	"github.com/go-acme/lego/v5/providers/dns/cloudns"
	"github.com/go-acme/lego/v5/providers/dns/dynadot"
	"github.com/go-acme/lego/v5/providers/dns/dynu"
	"github.com/go-acme/lego/v5/providers/dns/freemyip"
	"github.com/go-acme/lego/v5/providers/dns/godaddy"
	"github.com/go-acme/lego/v5/providers/dns/huaweicloud"
	"github.com/go-acme/lego/v5/providers/dns/ionos"
	"github.com/go-acme/lego/v5/providers/dns/ionoscloud"
	"github.com/go-acme/lego/v5/providers/dns/namecheap"
	"github.com/go-acme/lego/v5/providers/dns/namedotcom"
	"github.com/go-acme/lego/v5/providers/dns/namesilo"
	"github.com/go-acme/lego/v5/providers/dns/ovh"
	"github.com/go-acme/lego/v5/providers/dns/porkbun"
	"github.com/go-acme/lego/v5/providers/dns/rainyun"
	"github.com/go-acme/lego/v5/providers/dns/regru"
	"github.com/go-acme/lego/v5/providers/dns/route53"
	"github.com/go-acme/lego/v5/providers/dns/spaceship"
	"github.com/go-acme/lego/v5/providers/dns/technitium"
	"github.com/go-acme/lego/v5/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v5/providers/dns/vercel"
	"github.com/go-acme/lego/v5/providers/dns/volcengine"
	"github.com/go-acme/lego/v5/providers/dns/westcn"
)

type DnsType string

const (
	AcmeDNS      DnsType = "AcmeDNS"
	AliYun       DnsType = "AliYun"
	AliESA       DnsType = "AliESA"
	BaiduCloud   DnsType = "BaiduCloud"
	CloudDns     DnsType = "CloudDns"
	CloudFlare   DnsType = "CloudFlare"
	ClouDNS      DnsType = "ClouDNS"
	Dynadot      DnsType = "Dynadot"
	Dynu         DnsType = "Dynu"
	FreeMyIP     DnsType = "FreeMyIP"
	Godaddy      DnsType = "Godaddy"
	HuaweiCloud  DnsType = "HuaweiCloud"
	Ionos        DnsType = "Ionos"
	IonosCloud   DnsType = "IonosCloud"
	NameCheap    DnsType = "NameCheap"
	NameCom      DnsType = "NameCom"
	NameSilo     DnsType = "NameSilo"
	Ovh          DnsType = "Ovh"
	PorkBun      DnsType = "PorkBun"
	RainYun      DnsType = "RainYun"
	RegRu        DnsType = "RegRu"
	AWSRoute53   DnsType = "AWSRoute53"
	Spaceship    DnsType = "Spaceship"
	Technitium   DnsType = "Technitium"
	TencentCloud DnsType = "TencentCloud"
	Vercel       DnsType = "Vercel"
	Volcengine   DnsType = "Volcengine"
	WestCN       DnsType = "WestCN"
)

type DNSParam struct {
	ID           string `json:"id"`
	Token        string `json:"token"`
	AccessKey    string `json:"accessKey"`
	SecretKey    string `json:"secretKey"`
	Email        string `json:"email"`
	APIkey       string `json:"apiKey"`
	APIUser      string `json:"apiUser"`
	APISecret    string `json:"apiSecret"`
	APIPrefix    string `json:"apiPrefix"`
	SecretID     string `json:"secretID"`
	ClientID     string `json:"clientID"`
	Password     string `json:"password"`
	Region       string `json:"region"`
	Username     string `json:"username"`
	AuthID       string `json:"authID"`
	SubAuthID    string `json:"subAuthID"`
	AuthPassword string `json:"authPassword"`
	Endpoint     string `json:"endpoint"`
	AccessToken  string `json:"accessToken"`
	BaseURL      string `json:"baseURL"`
}

var (
	propagationTimeout = 30 * time.Minute
	pollingInterval    = 10 * time.Second
	manualDnsTimeout   = 10 * time.Minute
)

func getDNSProviderConfig(dnsType DnsType, params string, httpClient *http.Client) (challenge.Provider, error) {
	var (
		param DNSParam
		p     challenge.Provider
		err   error
	)
	if err := json.Unmarshal([]byte(params), &param); err != nil {
		return nil, err
	}
	switch dnsType {
	case AliYun:
		config := newDNSProviderConfig(alidns.NewDefaultConfig(), httpClient)
		config.SecretKey = param.SecretKey
		config.APIKey = param.AccessKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = alidns.NewDNSProviderConfig(config)
	case AliESA:
		config := newDNSProviderConfig(aliesa.NewDefaultConfig(), httpClient)
		config.SecretKey = param.SecretKey
		config.APIKey = param.AccessKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = aliesa.NewDNSProviderConfig(config)
	case AWSRoute53:
		config := newDNSProviderConfig(route53.NewDefaultConfig(), httpClient)
		config.AccessKeyID = param.AccessKey
		config.SecretAccessKey = param.SecretKey
		config.Region = param.Region
		if config.Region == "" {
			config.Region = "us-east-1"
		}
		config.HostedZoneID = param.Endpoint
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = route53.NewDNSProviderConfig(config)
	case CloudFlare:
		config := newDNSProviderConfig(cloudflare.NewDefaultConfig(), httpClient)
		config.AuthEmail = param.Email
		config.AuthToken = param.APIkey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = cloudflare.NewDNSProviderConfig(config)
	case CloudDns:
		config := newDNSProviderConfig(clouddns.NewDefaultConfig(), httpClient)
		config.ClientID = param.ClientID
		config.Email = param.Email
		config.Password = param.Password
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = clouddns.NewDNSProviderConfig(config)
	case NameCheap:
		config := newDNSProviderConfig(namecheap.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIkey
		config.APIUser = param.APIUser
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = namecheap.NewDNSProviderConfig(config)
	case NameSilo:
		config := newDNSProviderConfig(namesilo.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIkey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = namesilo.NewDNSProviderConfig(config)
	case Godaddy:
		config := newDNSProviderConfig(godaddy.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIkey
		config.APISecret = param.APISecret
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = godaddy.NewDNSProviderConfig(config)
	case NameCom:
		config := newDNSProviderConfig(namedotcom.NewDefaultConfig(), httpClient)
		config.APIToken = param.Token
		config.Username = param.APIUser
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = namedotcom.NewDNSProviderConfig(config)
	case TencentCloud:
		config := newDNSProviderConfig(tencentcloud.NewDefaultConfig(), httpClient)
		config.SecretID = param.SecretID
		config.SecretKey = param.SecretKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = tencentcloud.NewDNSProviderConfig(config)
	case RainYun:
		config := newDNSProviderConfig(rainyun.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIkey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = rainyun.NewDNSProviderConfig(config)
	case Volcengine:
		config := newDNSProviderConfig(volcengine.NewDefaultConfig(), httpClient)
		config.SecretKey = param.SecretKey
		config.AccessKey = param.AccessKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = volcengine.NewDNSProviderConfig(config)
	case HuaweiCloud:
		config := newDNSProviderConfig(huaweicloud.NewDefaultConfig(), httpClient)
		config.AccessKeyID = param.AccessKey
		config.SecretAccessKey = param.SecretKey
		config.Region = param.Region
		if config.Region == "" {
			config.Region = "cn-north-1"
		}
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = huaweicloud.NewDNSProviderConfig(config)
	case Ionos:
		config := newDNSProviderConfig(ionos.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIPrefix + "." + param.APISecret
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = ionos.NewDNSProviderConfig(config)
	case IonosCloud:
		config := newDNSProviderConfig(ionoscloud.NewDefaultConfig(), httpClient)
		config.APIToken = param.Token
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = ionoscloud.NewDNSProviderConfig(config)
	case FreeMyIP:
		config := newDNSProviderConfig(freemyip.NewDefaultConfig(), httpClient)
		config.Token = param.Token
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = freemyip.NewDNSProviderConfig(config)
	case Vercel:
		config := newDNSProviderConfig(vercel.NewDefaultConfig(), httpClient)
		config.AuthToken = param.Token
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = vercel.NewDNSProviderConfig(config)
	case Spaceship:
		config := newDNSProviderConfig(spaceship.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIkey
		config.APISecret = param.APISecret
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = spaceship.NewDNSProviderConfig(config)
	case WestCN:
		config := newDNSProviderConfig(westcn.NewDefaultConfig(), httpClient)
		config.Username = param.Username
		config.Password = param.Password
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = westcn.NewDNSProviderConfig(config)
	case ClouDNS:
		config := newDNSProviderConfig(cloudns.NewDefaultConfig(), httpClient)
		config.AuthID = param.AuthID
		config.SubAuthID = param.SubAuthID
		config.AuthPassword = param.AuthPassword
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = cloudns.NewDNSProviderConfig(config)
	case RegRu:
		config := newDNSProviderConfig(regru.NewDefaultConfig(), httpClient)
		config.Username = param.Username
		config.Password = param.Password
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = regru.NewDNSProviderConfig(config)
	case Dynu:
		config := newDNSProviderConfig(dynu.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIkey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = dynu.NewDNSProviderConfig(config)
	case Dynadot:
		config := newDNSProviderConfig(dynadot.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIkey
		config.APISecret = param.APISecret
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = dynadot.NewDNSProviderConfig(config)
	case BaiduCloud:
		config := newDNSProviderConfig(baiducloud.NewDefaultConfig(), httpClient)
		config.AccessKeyID = param.AccessKey
		config.SecretAccessKey = param.SecretKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = baiducloud.NewDNSProviderConfig(config)
	case Ovh:
		config := newDNSProviderConfig(ovh.NewDefaultConfig(), httpClient)
		config.APIEndpoint = param.Endpoint
		config.AccessToken = param.AccessToken
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = ovh.NewDNSProviderConfig(config)
	case AcmeDNS:
		config := newDNSProviderConfig(acmedns.NewDefaultConfig(), httpClient)
		config.APIBase = param.Endpoint
		config.StorageBaseURL = param.BaseURL
		p, err = acmedns.NewDNSProviderConfig(config)
	case PorkBun:
		config := newDNSProviderConfig(porkbun.NewDefaultConfig(), httpClient)
		config.APIKey = param.APIkey
		config.SecretAPIKey = param.SecretKey
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = porkbun.NewDNSProviderConfig(config)
	case Technitium:
		config := newDNSProviderConfig(technitium.NewDefaultConfig(), httpClient)
		config.BaseURL = param.BaseURL
		config.APIToken = param.Token
		config.PropagationTimeout = propagationTimeout
		config.PollingInterval = pollingInterval
		p, err = technitium.NewDNSProviderConfig(config)
	default:
		// Surfaces clear errors for legacy values (e.g. "DnsPod", which lego v5
		// removed) and for any future provider that the frontend can pick but
		// the backend has not yet wired up. Without this default branch, p and
		// err would both stay nil and the DNS-01 step would fail far away from
		// the real cause.
		if dnsType == "DnsPod" {
			return nil, fmt.Errorf("DNS provider %q has been removed in lego v5; please switch this DNS account to TencentCloud, which manages DNSPod-hosted zones via the same underlying API", dnsType)
		}
		return nil, fmt.Errorf("unsupported DNS provider %q", dnsType)
	}

	if err != nil {
		return nil, err
	}
	return p, nil
}

func newDNSProviderConfig[T any](config *T, httpClient *http.Client) *T {
	setDNSProviderHTTPClient(config, httpClient)
	return config
}

func setDNSProviderHTTPClient(config any, httpClient *http.Client) {
	if config == nil || httpClient == nil {
		return
	}
	value := reflect.ValueOf(config)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return
	}
	field := value.Elem().FieldByName("HTTPClient")
	if !field.IsValid() || !field.CanSet() {
		return
	}
	httpClientType := reflect.TypeOf((*http.Client)(nil))
	if field.Type() != httpClientType {
		return
	}
	field.Set(reflect.ValueOf(httpClient))
}
