// Package huaweicloud implements a DNS provider for solving the DNS-01 challenge using Tencent Cloud DNS.
package huaweicloud

import (
	"errors"
	"fmt"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	dnsConfig "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	dnsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	dnsRegion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "HUAWEICLOUD_"

	EnvAccessKeyId     = envNamespace + "ACCESS_KEY_ID"
	EnvSecretAccessKey = envNamespace + "SECRET_ACCESS_KEY"
	EnvRegion          = envNamespace + "REGION"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	AccessKeyId     string
	SecretAccessKey string
	Region          string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *dns.DnsClient
}

// NewDNSProvider returns a DNSProvider instance configured for OTC DNS.
// Credentials must be passed in the environment variables: OTC_USER_NAME,
// OTC_DOMAIN_NAME, OTC_PASSWORD OTC_PROJECT_NAME and OTC_IDENTITY_ENDPOINT.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAccessKeyId, EnvSecretAccessKey, EnvRegion)
	if err != nil {
		return nil, fmt.Errorf("huaweicloud: %w", err)
	}

	config := NewDefaultConfig()
	config.AccessKeyId = values[EnvAccessKeyId]
	config.SecretAccessKey = values[EnvSecretAccessKey]
	config.Region = values[EnvRegion]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OTC DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("huaweicloud: the configuration of the DNS provider is nil")
	}

	if config.AccessKeyId == "" || config.SecretAccessKey == "" || config.Region == "" {
		return nil, errors.New("huaweicloud: credentials missing")
	}

	auth, err := basic.NewCredentialsBuilder().
		WithAk(config.AccessKeyId).
		WithSk(config.SecretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("huaweicloud auth: %w", err)
	}

	httpConfig := dnsConfig.DefaultHttpConfig()
	httpConfig.WithTimeout(config.HTTPTimeout)

	region, err := dnsRegion.SafeValueOf(config.Region)
	if err != nil {
		return nil, fmt.Errorf("huaweicloud region: %w", err)
	}

	c, err := dns.DnsClientBuilder().
		WithHttpConfig(httpConfig).
		WithRegion(region).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("huaweicloud region: %w", err)
	}

	return &DNSProvider{config: config, client: dns.NewDnsClient(c)}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("huaweicloud: could not find zone for domain %q: %w", domain, err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return err
	}

	description := "Added TXT record for ACME dns-01 challenge using lego client"
	ttl := int32(d.config.TTL)
	_, err = d.client.CreateRecordSet(&dnsModel.CreateRecordSetRequest{
		ZoneId: zoneID,
		Body: &dnsModel.CreateRecordSetRequestBody{
			Name:        info.EffectiveFQDN,
			Description: &description,
			Type:        "TXT",
			Ttl:         &ttl,
			Records:     []string{fmt.Sprintf("%q", info.Value)},
		},
	})
	if err != nil {
		return fmt.Errorf("huaweicloud create: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("huaweicloud: could not find zone for domain %q: %w", domain, err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return err
	}

	records, err := d.client.ListRecordSetsByZone(&dnsModel.ListRecordSetsByZoneRequest{
		ZoneId: zoneID,
		Name:   &info.EffectiveFQDN,
	})
	if err != nil {
		return fmt.Errorf("huaweicloud record list: unable to get record %s for zone %s: %w", info.EffectiveFQDN, domain, err)
	}

	if len(*records.Recordsets) != 1 || *(*records.Recordsets)[0].Id == "" {
		return fmt.Errorf("huaweicloud record id: record set error")
	}
	recordID := (*records.Recordsets)[0].Id

	_, err = d.client.DeleteRecordSet(&dnsModel.DeleteRecordSetRequest{
		ZoneId:      zoneID,
		RecordsetId: *recordID,
	})
	if err != nil {
		return fmt.Errorf("huaweicloud delete: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.PropagationTimeout
}

func (d *DNSProvider) getZoneID(authZone string) (string, error) {
	zones, err := d.client.ListPublicZones(&dnsModel.ListPublicZonesRequest{})
	if err != nil {
		return "", fmt.Errorf("huaweicloud: unable to get zone: %w", err)
	}

	for _, zone := range *zones.Zones {
		if *zone.Name == authZone {
			return *zone.Id, nil
		}
	}

	return "", fmt.Errorf("huaweicloud: zone %q not found", authZone)
}
