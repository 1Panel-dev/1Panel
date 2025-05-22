package ssl

import "time"

type ManualConfig struct {
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

type CustomManualDnsProvider struct {
	config *ManualConfig
}

func NewCustomDNSProviderManual(config *ManualConfig) (*CustomManualDnsProvider, error) {
	return &CustomManualDnsProvider{config}, nil
}

func (p *CustomManualDnsProvider) Present(domain, token, keyAuth string) error {
	return nil
}

func (p *CustomManualDnsProvider) CleanUp(domain, token, keyAuth string) error {
	return nil
}

func (p *CustomManualDnsProvider) Sequential() time.Duration {
	return manualDnsTimeout
}

func (p *CustomManualDnsProvider) Timeout() (timeout, interval time.Duration) {
	return p.config.PropagationTimeout, p.config.PollingInterval
}
