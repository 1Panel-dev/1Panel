package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	dnsutil "github.com/1Panel-dev/1Panel/agent/utils/dns"
	"github.com/google/uuid"
)

type DnsService struct{}

type IDnsService interface {
	PageZone(req request.DnsZoneSearch) (int64, []response.DnsZoneRes, error)
	CreateZone(req request.DnsZoneCreate) error
	UpdateZone(req request.DnsZoneUpdate) error
	DeleteZone(id uint) error
	GetZone(id uint) (response.DnsZoneRes, error)

	PageRecord(req request.DnsRecordSearch) (int64, []response.DnsRecordRes, error)
	CreateRecord(req request.DnsRecordCreate) error
	UpdateRecord(req request.DnsRecordUpdate) error
	DeleteRecord(id uint) error

	Setup(req request.DnsSetup) error
	GetStatus() (response.DnsStatus, error)
	SyncZone(id uint) error

	AutoCreateARecord(domain, ip string) error
	AutoCreateMXRecords(domain, mailServer string) error
	FindZoneForDomain(domain string) (*model.DnsZone, error)
}

func NewIDnsService() IDnsService {
	return &DnsService{}
}

func (d *DnsService) getPdnsClient() (*dnsutil.PdnsClient, error) {
	enabled, _ := settingRepo.Get(settingRepo.WithByKey("DnsEnabled"))
	if enabled.Value != "true" {
		return nil, buserr.New("ErrDnsNotEnabled")
	}
	apiURL, _ := settingRepo.Get(settingRepo.WithByKey("DnsApiUrl"))
	apiKey, _ := settingRepo.Get(settingRepo.WithByKey("DnsApiKey"))
	if apiURL.Value == "" {
		apiURL.Value = "http://127.0.0.1:8081"
	}
	return dnsutil.NewPdnsClient(apiURL.Value, apiKey.Value), nil
}

// Zone operations

func (d *DnsService) PageZone(req request.DnsZoneSearch) (int64, []response.DnsZoneRes, error) {
	var opts []repo.DBOption
	if req.Info != "" {
		opts = append(opts, dnsZoneRepo.WithLikeName(req.Info))
	}
	opts = append(opts, repo.WithOrderRuleBy(req.OrderBy, req.Order))

	total, zones, err := dnsZoneRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}

	var results []response.DnsZoneRes
	for _, zone := range zones {
		count := int64(0)
		global.DB.Model(&model.DnsRecord{}).Where("zone_id = ?", zone.ID).Count(&count)
		results = append(results, response.DnsZoneRes{
			DnsZone:     zone,
			RecordCount: int(count),
		})
	}
	return total, results, nil
}

func (d *DnsService) GetZone(id uint) (response.DnsZoneRes, error) {
	zone, err := dnsZoneRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return response.DnsZoneRes{}, buserr.New("ErrDnsZoneNotFound")
	}
	count := int64(0)
	global.DB.Model(&model.DnsRecord{}).Where("zone_id = ?", zone.ID).Count(&count)
	return response.DnsZoneRes{
		DnsZone:     zone,
		RecordCount: int(count),
	}, nil
}

func (d *DnsService) CreateZone(req request.DnsZoneCreate) error {
	existing, _ := dnsZoneRepo.GetFirst(dnsZoneRepo.WithName(req.Name))
	if existing.ID != 0 {
		return buserr.WithMap("ErrDnsZoneExist", map[string]interface{}{"name": req.Name}, nil)
	}

	pdnsClient, err := d.getPdnsClient()
	if err != nil {
		return err
	}

	canonicalName := dnsutil.CanonicalName(req.Name)
	soaPrimary := req.SoaPrimary
	if soaPrimary == "" {
		soaPrimary = "ns1." + req.Name
	}
	defaultTTL := req.DefaultTTL
	if defaultTTL == 0 {
		defaultTTL = 3600
	}

	pdnsReq := dnsutil.PdnsZoneCreate{
		Name:        canonicalName,
		Kind:        req.Type,
		Nameservers: []string{dnsutil.CanonicalName(soaPrimary)},
		SOAEdit:     "DEFAULT",
		SOAEditAPI:  "DEFAULT",
	}
	if req.Type == "Slave" && req.MasterIP != "" {
		pdnsReq.Masters = []string{req.MasterIP}
	}

	if err := pdnsClient.CreateZone(pdnsReq); err != nil {
		return buserr.WithDetail("ErrDnsPdnsApi", err.Error(), err)
	}

	zone := &model.DnsZone{
		Name:       req.Name,
		Type:       req.Type,
		MasterIP:   req.MasterIP,
		SoaPrimary: soaPrimary,
		SoaEmail:   req.SoaEmail,
		SoaRefresh: 10800,
		SoaRetry:   3600,
		SoaExpire:  604800,
		SoaTTL:     3600,
		DefaultTTL: defaultTTL,
		Status:     "active",
		PdnsID:     canonicalName,
	}
	if err := dnsZoneRepo.Create(context.Background(), zone); err != nil {
		return err
	}

	// Sync records from PowerDNS (SOA and NS were auto-created)
	_ = d.syncRecordsFromPdns(zone)

	return nil
}

func (d *DnsService) UpdateZone(req request.DnsZoneUpdate) error {
	zone, err := dnsZoneRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return buserr.New("ErrDnsZoneNotFound")
	}

	upMap := make(map[string]interface{})
	if req.SoaPrimary != "" {
		upMap["soa_primary"] = req.SoaPrimary
	}
	if req.SoaEmail != "" {
		upMap["soa_email"] = req.SoaEmail
	}
	if req.SoaRefresh > 0 {
		upMap["soa_refresh"] = req.SoaRefresh
	}
	if req.SoaRetry > 0 {
		upMap["soa_retry"] = req.SoaRetry
	}
	if req.SoaExpire > 0 {
		upMap["soa_expire"] = req.SoaExpire
	}
	if req.SoaTTL > 0 {
		upMap["soa_ttl"] = req.SoaTTL
	}
	if req.DefaultTTL > 0 {
		upMap["default_ttl"] = req.DefaultTTL
	}

	if len(upMap) == 0 {
		return nil
	}

	_ = zone // used for future SOA update via PowerDNS if needed
	return dnsZoneRepo.Update(req.ID, upMap)
}

func (d *DnsService) DeleteZone(id uint) error {
	zone, err := dnsZoneRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return buserr.New("ErrDnsZoneNotFound")
	}

	pdnsClient, err := d.getPdnsClient()
	if err != nil {
		return err
	}

	if err := pdnsClient.DeleteZone(zone.PdnsID); err != nil {
		global.LOG.Errorf("failed to delete zone %s from PowerDNS: %v", zone.Name, err)
	}

	if err := dnsRecordRepo.DeleteBy(context.Background(), dnsRecordRepo.WithZoneID(id)); err != nil {
		return err
	}
	return dnsZoneRepo.DeleteBy(context.Background(), repo.WithByID(id))
}

// Record operations

func (d *DnsService) PageRecord(req request.DnsRecordSearch) (int64, []response.DnsRecordRes, error) {
	var opts []repo.DBOption
	opts = append(opts, dnsRecordRepo.WithZoneID(req.ZoneID))
	if req.Type != "" {
		opts = append(opts, dnsRecordRepo.WithType(req.Type))
	}
	if req.Info != "" {
		opts = append(opts, dnsRecordRepo.WithLikeName(req.Info))
	}
	opts = append(opts, repo.WithOrderDesc("created_at"))

	total, records, err := dnsRecordRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}

	zone, _ := dnsZoneRepo.GetFirst(repo.WithByID(req.ZoneID))
	var results []response.DnsRecordRes
	for _, record := range records {
		results = append(results, response.DnsRecordRes{
			DnsRecord: record,
			ZoneName:  zone.Name,
		})
	}
	return total, results, nil
}

func (d *DnsService) CreateRecord(req request.DnsRecordCreate) error {
	zone, err := dnsZoneRepo.GetFirst(repo.WithByID(req.ZoneID))
	if err != nil {
		return buserr.New("ErrDnsZoneNotFound")
	}

	pdnsClient, err := d.getPdnsClient()
	if err != nil {
		return err
	}

	fqdn := d.buildFQDN(req.Name, zone.Name)
	ttl := req.TTL
	if ttl == 0 {
		ttl = zone.DefaultTTL
	}

	content := req.Content
	if req.Type == "MX" || req.Type == "SRV" {
		content = fmt.Sprintf("%d %s", req.Priority, req.Content)
	}

	rrset := dnsutil.RRSet{
		Name:       dnsutil.CanonicalName(fqdn),
		Type:       req.Type,
		TTL:        ttl,
		ChangeType: "REPLACE",
		Records: []dnsutil.Record{
			{Content: content, Disabled: false},
		},
	}

	// Check for existing records with same name+type and merge
	existingRecords, _ := dnsRecordRepo.List(
		dnsRecordRepo.WithZoneID(req.ZoneID),
		dnsRecordRepo.WithType(req.Type),
	)
	for _, existing := range existingRecords {
		if existing.Name == fqdn {
			rrset.Records = append(rrset.Records, dnsutil.Record{
				Content:  existing.Content,
				Disabled: existing.Disabled,
			})
		}
	}

	if err := pdnsClient.PatchRecords(zone.PdnsID, []dnsutil.RRSet{rrset}); err != nil {
		return buserr.WithDetail("ErrDnsPdnsApi", err.Error(), err)
	}

	record := &model.DnsRecord{
		ZoneID:   req.ZoneID,
		Name:     fqdn,
		Type:     req.Type,
		Content:  req.Content,
		TTL:      ttl,
		Priority: req.Priority,
	}
	return dnsRecordRepo.Create(context.Background(), record)
}

func (d *DnsService) UpdateRecord(req request.DnsRecordUpdate) error {
	record, err := dnsRecordRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return buserr.New("ErrDnsRecordNotFound")
	}

	zone, err := dnsZoneRepo.GetFirst(repo.WithByID(record.ZoneID))
	if err != nil {
		return buserr.New("ErrDnsZoneNotFound")
	}

	pdnsClient, err := d.getPdnsClient()
	if err != nil {
		return err
	}

	fqdn := d.buildFQDN(req.Name, zone.Name)
	ttl := req.TTL
	if ttl == 0 {
		ttl = zone.DefaultTTL
	}

	content := req.Content
	if req.Type == "MX" || req.Type == "SRV" {
		content = fmt.Sprintf("%d %s", req.Priority, req.Content)
	}

	// If name or type changed, delete old record first
	if record.Name != fqdn || record.Type != req.Type {
		oldRRSet := dnsutil.RRSet{
			Name:       dnsutil.CanonicalName(record.Name),
			Type:       record.Type,
			ChangeType: "DELETE",
		}
		_ = pdnsClient.PatchRecords(zone.PdnsID, []dnsutil.RRSet{oldRRSet})
	}

	rrset := dnsutil.RRSet{
		Name:       dnsutil.CanonicalName(fqdn),
		Type:       req.Type,
		TTL:        ttl,
		ChangeType: "REPLACE",
		Records: []dnsutil.Record{
			{Content: content, Disabled: req.Disabled},
		},
	}

	if err := pdnsClient.PatchRecords(zone.PdnsID, []dnsutil.RRSet{rrset}); err != nil {
		return buserr.WithDetail("ErrDnsPdnsApi", err.Error(), err)
	}

	record.Name = fqdn
	record.Type = req.Type
	record.Content = req.Content
	record.TTL = ttl
	record.Priority = req.Priority
	record.Disabled = req.Disabled
	return dnsRecordRepo.Save(context.Background(), &record)
}

func (d *DnsService) DeleteRecord(id uint) error {
	record, err := dnsRecordRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return buserr.New("ErrDnsRecordNotFound")
	}

	zone, err := dnsZoneRepo.GetFirst(repo.WithByID(record.ZoneID))
	if err != nil {
		return buserr.New("ErrDnsZoneNotFound")
	}

	pdnsClient, err := d.getPdnsClient()
	if err != nil {
		return err
	}

	rrset := dnsutil.RRSet{
		Name:       dnsutil.CanonicalName(record.Name),
		Type:       record.Type,
		ChangeType: "DELETE",
	}

	if err := pdnsClient.PatchRecords(zone.PdnsID, []dnsutil.RRSet{rrset}); err != nil {
		global.LOG.Errorf("failed to delete record from PowerDNS: %v", err)
	}

	return dnsRecordRepo.DeleteBy(context.Background(), repo.WithByID(id))
}

// Setup / status

func (d *DnsService) Setup(req request.DnsSetup) error {
	if req.Enable {
		apiKey := uuid.New().String()
		if err := dnsutil.EnsurePdnsContainer(apiKey); err != nil {
			return buserr.WithDetail("ErrDnsContainerFailed", err.Error(), err)
		}
		d.saveSetting("DnsEnabled", "true")
		d.saveSetting("DnsApiKey", apiKey)
		d.saveSetting("DnsApiUrl", "http://127.0.0.1:8081")
		d.saveSetting("DnsContainerName", dnsutil.PdnsContainerName)
	} else {
		_ = dnsutil.StopPdnsContainer()
		d.saveSetting("DnsEnabled", "false")
	}
	return nil
}

func (d *DnsService) GetStatus() (response.DnsStatus, error) {
	enabled, _ := settingRepo.Get(settingRepo.WithByKey("DnsEnabled"))
	containerName, _ := settingRepo.Get(settingRepo.WithByKey("DnsContainerName"))

	status := dnsutil.GetContainerStatus()

	result := response.DnsStatus{
		Enabled:       enabled.Value == "true",
		ContainerUp:   status.Running,
		ContainerName: containerName.Value,
		Version:       status.Version,
	}
	return result, nil
}

func (d *DnsService) SyncZone(id uint) error {
	zone, err := dnsZoneRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return buserr.New("ErrDnsZoneNotFound")
	}
	return d.syncRecordsFromPdns(&zone)
}

// Integration hooks

func (d *DnsService) AutoCreateARecord(domain, ip string) error {
	zone, err := d.FindZoneForDomain(domain)
	if err != nil || zone == nil {
		return nil
	}
	return d.CreateRecord(request.DnsRecordCreate{
		ZoneID:  zone.ID,
		Name:    domain,
		Type:    "A",
		Content: ip,
	})
}

func (d *DnsService) AutoCreateMXRecords(domain, mailServer string) error {
	zone, err := d.FindZoneForDomain(domain)
	if err != nil || zone == nil {
		return nil
	}
	return d.CreateRecord(request.DnsRecordCreate{
		ZoneID:   zone.ID,
		Name:     domain,
		Type:     "MX",
		Content:  mailServer,
		Priority: 10,
	})
}

func (d *DnsService) FindZoneForDomain(domain string) (*model.DnsZone, error) {
	parts := strings.Split(domain, ".")
	for i := 0; i < len(parts)-1; i++ {
		candidate := strings.Join(parts[i:], ".")
		zone, err := dnsZoneRepo.GetFirst(dnsZoneRepo.WithName(candidate))
		if err == nil && zone.ID != 0 {
			return &zone, nil
		}
	}
	return nil, nil
}

// Internal helpers

func (d *DnsService) buildFQDN(name, zoneName string) string {
	if name == "@" || name == "" {
		return zoneName
	}
	if strings.HasSuffix(name, "."+zoneName) {
		return name
	}
	if strings.HasSuffix(name, ".") {
		return name[:len(name)-1]
	}
	return name + "." + zoneName
}

func (d *DnsService) syncRecordsFromPdns(zone *model.DnsZone) error {
	pdnsClient, err := d.getPdnsClient()
	if err != nil {
		return err
	}

	pdnsZone, err := pdnsClient.GetZone(zone.PdnsID)
	if err != nil {
		return err
	}

	// Delete existing cached records
	_ = dnsRecordRepo.DeleteBy(context.Background(), dnsRecordRepo.WithZoneID(zone.ID))

	// Re-create from PowerDNS data
	for _, rrset := range pdnsZone.RRSets {
		for _, record := range rrset.Records {
			name := strings.TrimSuffix(rrset.Name, ".")
			priority := 0
			content := record.Content

			// Parse priority from MX/SRV content
			if rrset.Type == "MX" || rrset.Type == "SRV" {
				parts := strings.SplitN(content, " ", 2)
				if len(parts) == 2 {
					fmt.Sscanf(parts[0], "%d", &priority)
					content = parts[1]
				}
			}

			_ = dnsRecordRepo.Create(context.Background(), &model.DnsRecord{
				ZoneID:   zone.ID,
				Name:     name,
				Type:     rrset.Type,
				Content:  content,
				TTL:      rrset.TTL,
				Priority: priority,
				Disabled: record.Disabled,
			})
		}
	}
	return nil
}

func (d *DnsService) saveSetting(key, value string) {
	_ = settingRepo.UpdateOrCreate(key, value)
}
