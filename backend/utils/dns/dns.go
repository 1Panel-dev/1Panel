package dns

import (
	"encoding/json"
	"github.com/1Panel-dev/1Panel/backend/utils/dns/aliyun"
	"github.com/1Panel-dev/1Panel/backend/utils/dns/huawei"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
)

type Provider interface {
	CreateOrUpdateRecord(ip string) error
}

type Param struct {
	ssl.DNSParam
	Domain string
	Name   string
}

func NewDNSProvider(dnsType ssl.DnsType, params string) (*Provider, error) {
	var (
		param Param
		p     Provider
		err   error
	)

	if err = json.Unmarshal([]byte(params), &param); err != nil {
		return nil, err
	}

	switch dnsType {
	case ssl.AliYun:
		p, err = aliyun.NewProvider(
			param.AccessKey,
			param.SecretKey,
			param.Region,
			param.Domain,
			param.Name,
		)
	case ssl.TencentCloud:
		p, err = aliyun.NewProvider(
			param.AccessKey,
			param.SecretKey,
			param.Region,
			param.Domain,
			param.Name,
		)
	case ssl.HuaweiCloud:
		p, err = huawei.NewProvider(
			param.AccessKey,
			param.SecretKey,
			param.Region,
			param.Domain,
			param.Name,
		)
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}
