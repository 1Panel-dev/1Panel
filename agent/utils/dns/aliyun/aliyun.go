package aliyun

import (
	"fmt"
	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	aliapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/pkg/errors"
)

type Provider struct {
	domain string
	name   string
	client *alidns.Client

	recordID *string
}

func NewProvider(accessKeyID, secretAccessKey, endpoint, domain, name string) (*Provider, error) {
	if accessKeyID == "" || secretAccessKey == "" || endpoint == "" || domain == "" {
		return nil, errors.New("credentials missing")
	}
	if name == "" {
		name = "@"
	}

	client, err := alidns.NewClient(&aliapi.Config{
		AccessKeyId:     &accessKeyID,
		AccessKeySecret: &secretAccessKey,
		Endpoint:        &endpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("dns client build: %w", err)
	}

	records, err := client.DescribeDomainRecords(&alidns.DescribeDomainRecordsRequest{
		DomainName: &domain,
		KeyWord:    &name,
		SearchMode: pointer("EXACT"),
	})
	if err != nil {
		return nil, fmt.Errorf("record list: unable to get record %s.%s: %w", name, domain, err)
	}

	for _, record := range records.Body.DomainRecords.Record {
		if *record.Type == "A" {
			return &Provider{domain: domain, name: name, client: client, recordID: record.RecordId}, nil
		}
	}

	return &Provider{
		domain: domain,
		name:   name,
		client: client,
	}, nil
}

func (p *Provider) CreateOrUpdateRecord(ip string) error {
	if p.recordID == nil {
		_, err := p.client.AddDomainRecord(&alidns.AddDomainRecordRequest{
			DomainName: &p.domain,
			RR:         &p.name,
			Type:       pointer("A"),
			Value:      &ip,
		})
		if err != nil {
			return fmt.Errorf("create record set: %w", err)
		}
		return nil
	}

	_, err := p.client.UpdateDomainRecord(&alidns.UpdateDomainRecordRequest{
		RecordId: p.recordID,
		RR:       &p.name,
		Type:     pointer("A"),
		Value:    &ip,
	})
	if err != nil {
		return fmt.Errorf("update record set: %w", err)
	}

	return nil
}

func pointer[T any](v T) *T {
	return &v
}
