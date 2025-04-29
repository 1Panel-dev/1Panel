package tencent

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	errorsdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

type Provider struct {
	domain string
	name   string
	client *dnspod.Client

	recordID uint64
}

func NewProvider(accessKeyID, secretAccessKey, region, domain, name string) (*Provider, error) {
	if accessKeyID == "" || secretAccessKey == "" || region == "" || domain == "" {
		return nil, errors.New("credentials missing")
	}
	if name == "" {
		name = "@"
	}

	credential := common.NewCredential(
		accessKeyID,
		secretAccessKey,
	)
	client, err := dnspod.NewClient(credential, region, profile.NewClientProfile())
	if err != nil {
		return nil, fmt.Errorf("dns client build: %w", err)
	}

	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = &domain
	request.Subdomain = &name
	request.RecordType = pointer("A")
	records, err := client.DescribeRecordList(request)
	var sdkError *errorsdk.TencentCloudSDKError
	if errors.As(err, &sdkError) {
		if sdkError.Code != dnspod.RESOURCENOTFOUND_NODATAOFRECORD {
			return nil, fmt.Errorf("record list: unable to get record %s.%s: %w", name, domain, err)
		}
	}

	if records.Response != nil && *records.Response.RecordCountInfo.ListCount > 0 {
		return &Provider{
			domain:   domain,
			name:     name,
			client:   client,
			recordID: *(records.Response.RecordList)[0].RecordId,
		}, nil
	}

	return &Provider{
		domain: domain,
		name:   name,
		client: client,
	}, nil
}

func (p *Provider) CreateOrUpdateRecord(ip string) error {
	if p.recordID == 0 {
		request := dnspod.NewCreateRecordRequest()
		request.Domain = &p.domain
		request.RecordType = pointer("A")
		request.RecordLine = pointer("默认")
		request.Value = &ip
		request.SubDomain = &p.name
		request.Remark = pointer("Added record for 1Panel")
		_, err := p.client.CreateRecord(request)
		if err != nil {
			return fmt.Errorf("create record set: %w", err)
		}

		return nil
	}

	request := dnspod.NewModifyRecordRequest()
	request.Domain = &p.domain
	request.RecordType = pointer("A")
	request.RecordLine = pointer("默认")
	request.Value = &ip
	request.RecordId = &p.recordID
	request.SubDomain = &p.name
	request.Remark = pointer("Added record for 1Panel")
	_, err := p.client.ModifyRecord(request)
	if err != nil {
		return fmt.Errorf("update record set: %w", err)
	}

	return nil
}

func pointer[T any](v T) *T {
	return &v
}
