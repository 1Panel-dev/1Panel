package ssl

import (
	"encoding/json"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/clouddns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/cloudns"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/providers/dns/freemyip"
	"github.com/go-acme/lego/v4/providers/dns/godaddy"
	"github.com/go-acme/lego/v4/providers/dns/huaweicloud"
	"github.com/go-acme/lego/v4/providers/dns/namecheap"
	"github.com/go-acme/lego/v4/providers/dns/namedotcom"
	"github.com/go-acme/lego/v4/providers/dns/namesilo"
	"github.com/go-acme/lego/v4/providers/dns/rainyun"
	"github.com/go-acme/lego/v4/providers/dns/spaceship"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/providers/dns/vercel"
	"github.com/go-acme/lego/v4/providers/dns/volcengine"
	"github.com/go-acme/lego/v4/providers/dns/westcn"
	"time"
)

type DnsType string

const (
	DnsPod       DnsType = "DnsPod"
	AliYun       DnsType = "AliYun"
	CloudFlare   DnsType = "CloudFlare"
	CloudDns     DnsType = "CloudDns"
	NameSilo     DnsType = "NameSilo"
	NameCheap    DnsType = "NameCheap"
	NameCom      DnsType = "NameCom"
	Godaddy      DnsType = "Godaddy"
	TencentCloud DnsType = "TencentCloud"
	RainYun      DnsType = "RainYun"
	Volcengine   DnsType = "Volcengine"
	HuaweiCloud  DnsType = "HuaweiCloud"
	FreeMyIP     DnsType = "FreeMyIP"
	Vercel       DnsType = "Vercel"
	Spaceship    DnsType = "Spaceship"
	WestCN       DnsType = "WestCN"
	ClouDNS      DnsType = "ClouDNS"
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
	SecretID     string `json:"secretID"`
	ClientID     string `json:"clientID"`
	Password     string `json:"password"`
	Region       string `json:"region"`
	Username     string `json:"username"`
	AuthID       string `json:"authID"`
	SubAuthID    string `json:"subAuthID"`
	AuthPassword string `json:"authPassword"`
}

var (
	propagationTimeout = 30 * time.Minute
	pollingInterval    = 10 * time.Second
	ttl                = 3600
	dnsTimeOut         = 30 * time.Minute
	manualDnsTimeout   = 10 * time.Minute
)

func getDNSProviderConfig(dnsType DnsType, params string) (challenge.Provider, error) {
	var (
		param DNSParam
		p     challenge.Provider
		err   error
	)
	if err := json.Unmarshal([]byte(params), &param); err != nil {
		return nil, err
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
	case CloudFlare:
		cloudflareConfig := cloudflare.NewDefaultConfig()
		cloudflareConfig.AuthEmail = param.Email
		cloudflareConfig.AuthToken = param.APIkey
		cloudflareConfig.PropagationTimeout = propagationTimeout
		cloudflareConfig.PollingInterval = pollingInterval
		cloudflareConfig.TTL = ttl
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
	case RainYun:
		rainyunConfig := rainyun.NewDefaultConfig()
		rainyunConfig.APIKey = param.APIkey
		rainyunConfig.PropagationTimeout = propagationTimeout
		rainyunConfig.PollingInterval = pollingInterval
		rainyunConfig.TTL = ttl
		p, err = rainyun.NewDNSProviderConfig(rainyunConfig)
	case Volcengine:
		volcConfig := volcengine.NewDefaultConfig()
		volcConfig.SecretKey = param.SecretKey
		volcConfig.AccessKey = param.AccessKey
		volcConfig.PropagationTimeout = propagationTimeout
		volcConfig.PollingInterval = pollingInterval
		volcConfig.TTL = ttl
		p, err = volcengine.NewDNSProviderConfig(volcConfig)
	case HuaweiCloud:
		huaweiCloudConfig := huaweicloud.NewDefaultConfig()
		huaweiCloudConfig.AccessKeyID = param.AccessKey
		huaweiCloudConfig.SecretAccessKey = param.SecretKey
		huaweiCloudConfig.Region = param.Region
		huaweiCloudConfig.PropagationTimeout = propagationTimeout
		huaweiCloudConfig.PollingInterval = pollingInterval
		huaweiCloudConfig.TTL = int32(ttl)
		p, err = huaweicloud.NewDNSProviderConfig(huaweiCloudConfig)
	case FreeMyIP:
		freeMyIpConfig := freemyip.NewDefaultConfig()
		freeMyIpConfig.Token = param.Token
		freeMyIpConfig.PropagationTimeout = propagationTimeout
		freeMyIpConfig.PollingInterval = pollingInterval
		p, err = freemyip.NewDNSProviderConfig(freeMyIpConfig)
	case Vercel:
		vercelConfig := vercel.NewDefaultConfig()
		vercelConfig.AuthToken = param.Token
		vercelConfig.PropagationTimeout = propagationTimeout
		vercelConfig.PollingInterval = pollingInterval
		p, err = vercel.NewDNSProviderConfig(vercelConfig)
	case Spaceship:
		spaceshipConfig := spaceship.NewDefaultConfig()
		spaceshipConfig.APIKey = param.APIkey
		spaceshipConfig.APISecret = param.APISecret
		spaceshipConfig.PropagationTimeout = propagationTimeout
		spaceshipConfig.PollingInterval = pollingInterval
		spaceshipConfig.TTL = ttl
		p, err = spaceship.NewDNSProviderConfig(spaceshipConfig)
	case WestCN:
		westcnConfig := westcn.NewDefaultConfig()
		westcnConfig.Username = param.Username
		westcnConfig.Password = param.Password
		westcnConfig.PropagationTimeout = propagationTimeout
		westcnConfig.PollingInterval = pollingInterval
		westcnConfig.TTL = ttl
		p, err = westcn.NewDNSProviderConfig(westcnConfig)

	case ClouDNS:
		cloudnsConfig := cloudns.NewDefaultConfig()
		cloudnsConfig.AuthID = param.AuthID
		cloudnsConfig.SubAuthID = param.SubAuthID
		cloudnsConfig.AuthPassword = param.AuthPassword
		cloudnsConfig.PropagationTimeout = propagationTimeout
		cloudnsConfig.PollingInterval = pollingInterval
		cloudnsConfig.TTL = ttl
		p, err = cloudns.NewDNSProviderConfig(cloudnsConfig)
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}
