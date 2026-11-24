package service

import (
	"encoding/json"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/dns"
)

type WebsiteDnsRecordService struct {
}

type IWebsiteDnsRecordService interface {
	ListZones(req request.DnsZoneSearch) ([]dns.Zone, error)
	ListRecords(req request.DnsRecordSearch) ([]dns.Record, error)
	CreateRecord(req request.DnsRecordOperate) error
	UpdateRecord(req request.DnsRecordOperate) error
	DeleteRecord(req request.DnsRecordDelete) error
}

func NewIWebsiteDnsRecordService() IWebsiteDnsRecordService {
	return &WebsiteDnsRecordService{}
}

func (w WebsiteDnsRecordService) client(accountID uint) (dns.Client, error) {
	account, err := websiteDnsRepo.GetFirst(repo.WithByID(accountID))
	if err != nil || account == nil {
		return nil, buserr.New("ErrDnsAccountNotFound")
	}
	auth := make(map[string]string)
	if err := json.Unmarshal([]byte(account.Authorization), &auth); err != nil {
		return nil, err
	}
	return dns.NewClient(account.Type, auth)
}

func (w WebsiteDnsRecordService) ListZones(req request.DnsZoneSearch) ([]dns.Zone, error) {
	client, err := w.client(req.AccountID)
	if err != nil {
		return nil, err
	}
	return client.ListZones()
}

func (w WebsiteDnsRecordService) ListRecords(req request.DnsRecordSearch) ([]dns.Record, error) {
	client, err := w.client(req.AccountID)
	if err != nil {
		return nil, err
	}
	return client.ListRecords(dns.Zone{ID: req.ZoneID, Name: req.ZoneName})
}

func toDnsRecord(req request.DnsRecordOperate) dns.Record {
	return dns.Record{
		ID:       req.ID,
		Name:     req.Name,
		Type:     req.Type,
		Value:    req.Value,
		TTL:      req.TTL,
		Priority: req.Priority,
		Proxied:  req.Proxied,
		Remark:   req.Remark,
	}
}

func (w WebsiteDnsRecordService) CreateRecord(req request.DnsRecordOperate) error {
	client, err := w.client(req.AccountID)
	if err != nil {
		return err
	}
	return client.CreateRecord(dns.Zone{ID: req.ZoneID, Name: req.ZoneName}, toDnsRecord(req))
}

func (w WebsiteDnsRecordService) UpdateRecord(req request.DnsRecordOperate) error {
	client, err := w.client(req.AccountID)
	if err != nil {
		return err
	}
	return client.UpdateRecord(dns.Zone{ID: req.ZoneID, Name: req.ZoneName}, toDnsRecord(req))
}

func (w WebsiteDnsRecordService) DeleteRecord(req request.DnsRecordDelete) error {
	client, err := w.client(req.AccountID)
	if err != nil {
		return err
	}
	return client.DeleteRecord(dns.Zone{ID: req.ZoneID, Name: req.ZoneName}, dns.Record{ID: req.ID, Name: req.Name, Type: req.Type})
}
