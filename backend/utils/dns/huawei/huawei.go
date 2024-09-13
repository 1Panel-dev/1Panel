package huawei

import (
	"fmt"
	hwauthbasic "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	hwdns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	hwmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	hwregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
	"github.com/pkg/errors"
)

type Provider struct {
	name string

	client   *hwdns.DnsClient
	zone     *hwmodel.PublicZoneResp
	recordID string
}

func NewProvider(accessKeyID, secretAccessKey, region, domain, name string) (*Provider, error) {
	if accessKeyID == "" || secretAccessKey == "" || region == "" || domain == "" {
		return nil, errors.New("credentials missing")
	}

	provider := &Provider{
		name: name,
	}

	auth, err := hwauthbasic.NewCredentialsBuilder().
		WithAk(accessKeyID).
		WithSk(secretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("crendential build: %w", err)
	}

	r, err := hwregion.SafeValueOf(region)
	if err != nil {
		return nil, fmt.Errorf("safe region: %w", err)
	}

	c, err := hwdns.DnsClientBuilder().
		WithRegion(r).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("dns client build: %w", err)
	}
	provider.client = hwdns.NewDnsClient(c)

	// get zoneID
	zones, err := provider.client.ListPublicZones(&hwmodel.ListPublicZonesRequest{
		Name:       &domain,
		SearchMode: pointer("equal"),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get domain: %w", err)
	}
	if *zones.Metadata.TotalCount != 1 {
		return nil, fmt.Errorf("domain %q not found", domain)
	}
	provider.zone = &(*zones.Zones)[0]

	// get recordID
	records, err := provider.client.ListRecordSetsByZone(&hwmodel.ListRecordSetsByZoneRequest{
		ZoneId:     *provider.zone.Id,
		SearchMode: pointer("equal"),
		Name:       pointer(name + "." + domain),
		Type:       pointer("A"),
	})
	if err != nil {
		return nil, fmt.Errorf("record list: unable to get record %s.%s: %w", name, domain, err)
	}
	if *records.Metadata.TotalCount > 0 {
		provider.recordID = *(*records.Recordsets)[0].Id
	}

	return provider, nil
}

func (p *Provider) CreateOrUpdateRecord(ip string) error {
	if p.recordID == "" {
		request := &hwmodel.CreateRecordSetRequest{
			ZoneId: *p.zone.Id,
			Body: &hwmodel.CreateRecordSetRequestBody{
				Name:        fmt.Sprintf("%s.%s", p.name, *p.zone.Name),
				Description: pointer("Added record for 1Panel"),
				Type:        "A",
				Records:     []string{ip},
			},
		}

		_, err := p.client.CreateRecordSet(request)
		if err != nil {
			return fmt.Errorf("create record set: %w", err)
		}

		return nil
	}

	updateRequest := &hwmodel.UpdateRecordSetRequest{
		ZoneId:      *p.zone.Id,
		RecordsetId: p.recordID,
		Body: &hwmodel.UpdateRecordSetReq{
			Description: pointer("Added record for 1Panel"),
			Records:     pointer([]string{ip}),
		},
	}

	_, err := p.client.UpdateRecordSet(updateRequest)
	if err != nil {
		return fmt.Errorf("update record set: %w", err)
	}

	return nil
}

func (p *Provider) getZoneID(name string) (string, error) {
	zones, err := p.client.ListPublicZones(&hwmodel.ListPublicZonesRequest{
		Name:       &name,
		SearchMode: pointer("equal"),
	})
	if err != nil {
		return "", fmt.Errorf("unable to get zone: %w", err)
	}
	if *zones.Metadata.TotalCount != 1 {
		return "", fmt.Errorf("zone %q not found", name)
	}

	return *(*zones.Zones)[0].Id, nil
}

func pointer[T any](v T) *T {
	return &v
}
