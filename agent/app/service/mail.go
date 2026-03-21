package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	mailutil "github.com/1Panel-dev/1Panel/agent/utils/mail"
)

type MailService struct{}

type IMailService interface {
	ListDomains() ([]response.MailDomainRes, error)
	CreateDomain(req request.MailDomainCreate) error
	DeleteDomain(req request.MailDomainDelete) error

	SearchAccounts(req request.MailAccountSearch) (int64, []response.MailAccountRes, error)
	CreateAccount(req request.MailAccountCreate) error
	UpdateAccount(req request.MailAccountUpdate) error
	DeleteAccount(id uint) error

	SearchAliases(req request.MailAliasSearch) (int64, []response.MailAliasRes, error)
	CreateAlias(req request.MailAliasCreate) error
	DeleteAlias(id uint) error

	Setup(req request.MailSetup) error
	GetStatus() (response.MailStatus, error)
}

func NewIMailService() IMailService {
	return &MailService{}
}

// Domain operations

func (m *MailService) ListDomains() ([]response.MailDomainRes, error) {
	domains, err := mailDomainRepo.List(repo.WithOrderDesc("created_at"))
	if err != nil {
		return nil, err
	}
	var results []response.MailDomainRes
	for _, domain := range domains {
		accountCount := int64(0)
		aliasCount := int64(0)
		global.DB.Model(&model.MailAccount{}).Where("domain_id = ?", domain.ID).Count(&accountCount)
		global.DB.Model(&model.MailAlias{}).Where("domain_id = ?", domain.ID).Count(&aliasCount)
		results = append(results, response.MailDomainRes{
			MailDomain:   domain,
			AccountCount: int(accountCount),
			AliasCount:   int(aliasCount),
		})
	}
	return results, nil
}

func (m *MailService) CreateDomain(req request.MailDomainCreate) error {
	existing, _ := mailDomainRepo.GetFirst(mailDomainRepo.WithName(req.Name))
	if existing.ID != 0 {
		return buserr.WithMap("ErrMailDomainExist", map[string]interface{}{"name": req.Name}, nil)
	}

	// Generate DKIM key via docker exec
	if err := mailutil.GenerateDKIM(req.Name); err != nil {
		global.LOG.Errorf("failed to generate DKIM for %s: %v", req.Name, err)
	}

	dkimKey, _ := mailutil.GetDKIMPublicKey(req.Name)

	domain := &model.MailDomain{
		Name:       req.Name,
		Status:     "active",
		DkimKey:    dkimKey,
		DnsAutoGen: req.DnsAutoGen,
	}
	if err := mailDomainRepo.Create(context.Background(), domain); err != nil {
		return err
	}

	// Auto-create DNS records if DNS is enabled
	if req.DnsAutoGen {
		m.autoCreateDnsRecords(domain)
	}

	return nil
}

func (m *MailService) DeleteDomain(req request.MailDomainDelete) error {
	domain, err := mailDomainRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return buserr.New("ErrMailDomainNotFound")
	}

	// Delete all accounts from mailserver
	accounts, _ := mailAccountRepo.List(mailAccountRepo.WithDomainID(domain.ID))
	for _, account := range accounts {
		_ = mailutil.DeleteAccount(account.Email)
	}

	// Delete all aliases from mailserver
	aliases, _ := mailAliasRepo.List(mailAliasRepo.WithDomainID(domain.ID))
	for _, alias := range aliases {
		_ = mailutil.DeleteAlias(alias.Source, alias.Target)
	}

	// Cleanup DB
	_ = mailAccountRepo.DeleteBy(context.Background(), mailAccountRepo.WithDomainID(domain.ID))
	_ = mailAliasRepo.DeleteBy(context.Background(), mailAliasRepo.WithDomainID(domain.ID))
	return mailDomainRepo.DeleteBy(context.Background(), repo.WithByID(req.ID))
}

// Account operations

func (m *MailService) SearchAccounts(req request.MailAccountSearch) (int64, []response.MailAccountRes, error) {
	var opts []repo.DBOption
	if req.DomainID > 0 {
		opts = append(opts, mailAccountRepo.WithDomainID(req.DomainID))
	}
	if req.Info != "" {
		opts = append(opts, mailAccountRepo.WithLikeInfo(req.Info))
	}
	opts = append(opts, repo.WithOrderDesc("created_at"))

	total, accounts, err := mailAccountRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}

	var results []response.MailAccountRes
	for _, account := range accounts {
		domainName := ""
		domain, _ := mailDomainRepo.GetFirst(repo.WithByID(account.DomainID))
		if domain.ID != 0 {
			domainName = domain.Name
		}
		results = append(results, response.MailAccountRes{
			MailAccount: account,
			DomainName:  domainName,
		})
	}
	return total, results, nil
}

func (m *MailService) CreateAccount(req request.MailAccountCreate) error {
	domain, err := mailDomainRepo.GetFirst(repo.WithByID(req.DomainID))
	if err != nil {
		return buserr.New("ErrMailDomainNotFound")
	}

	email := fmt.Sprintf("%s@%s", req.Username, domain.Name)
	existing, _ := mailAccountRepo.GetFirst(mailAccountRepo.WithEmail(email))
	if existing.ID != 0 {
		return buserr.WithMap("ErrMailAccountExist", map[string]interface{}{"email": email}, nil)
	}

	if err := mailutil.AddAccount(email, req.Password); err != nil {
		return buserr.WithDetail("ErrMailExecFailed", err.Error(), err)
	}

	if req.Quota > 0 {
		if err := mailutil.SetQuota(email, req.Quota); err != nil {
			global.LOG.Errorf("failed to set quota for %s: %v", email, err)
		}
	}

	account := &model.MailAccount{
		DomainID: req.DomainID,
		Username: req.Username,
		Email:    email,
		Quota:    req.Quota,
		Status:   "active",
	}
	return mailAccountRepo.Create(context.Background(), account)
}

func (m *MailService) UpdateAccount(req request.MailAccountUpdate) error {
	account, err := mailAccountRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return buserr.New("ErrMailAccountNotFound")
	}

	if req.Password != "" {
		if err := mailutil.UpdatePassword(account.Email, req.Password); err != nil {
			return buserr.WithDetail("ErrMailExecFailed", err.Error(), err)
		}
	}

	if req.Quota != account.Quota {
		if err := mailutil.SetQuota(account.Email, req.Quota); err != nil {
			global.LOG.Errorf("failed to set quota for %s: %v", account.Email, err)
		}
		account.Quota = req.Quota
	}

	return mailAccountRepo.Save(context.Background(), &account)
}

func (m *MailService) DeleteAccount(id uint) error {
	account, err := mailAccountRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return buserr.New("ErrMailAccountNotFound")
	}

	if err := mailutil.DeleteAccount(account.Email); err != nil {
		global.LOG.Errorf("failed to delete account %s from mailserver: %v", account.Email, err)
	}

	return mailAccountRepo.DeleteBy(context.Background(), repo.WithByID(id))
}

// Alias operations

func (m *MailService) SearchAliases(req request.MailAliasSearch) (int64, []response.MailAliasRes, error) {
	var opts []repo.DBOption
	if req.DomainID > 0 {
		opts = append(opts, mailAliasRepo.WithDomainID(req.DomainID))
	}
	if req.Info != "" {
		opts = append(opts, mailAliasRepo.WithLikeInfo(req.Info))
	}
	opts = append(opts, repo.WithOrderDesc("created_at"))

	total, aliases, err := mailAliasRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}

	var results []response.MailAliasRes
	for _, alias := range aliases {
		domainName := ""
		domain, _ := mailDomainRepo.GetFirst(repo.WithByID(alias.DomainID))
		if domain.ID != 0 {
			domainName = domain.Name
		}
		results = append(results, response.MailAliasRes{
			MailAlias:  alias,
			DomainName: domainName,
		})
	}
	return total, results, nil
}

func (m *MailService) CreateAlias(req request.MailAliasCreate) error {
	domain, err := mailDomainRepo.GetFirst(repo.WithByID(req.DomainID))
	if err != nil {
		return buserr.New("ErrMailDomainNotFound")
	}

	source := req.Source
	if !strings.Contains(source, "@") {
		source = source + "@" + domain.Name
	}

	existing, _ := mailAliasRepo.GetFirst(mailAliasRepo.WithSource(source))
	if existing.ID != 0 {
		return buserr.New("ErrMailAliasExist")
	}

	if err := mailutil.AddAlias(source, req.Target); err != nil {
		return buserr.WithDetail("ErrMailExecFailed", err.Error(), err)
	}

	alias := &model.MailAlias{
		DomainID: req.DomainID,
		Source:   source,
		Target:   req.Target,
	}
	return mailAliasRepo.Create(context.Background(), alias)
}

func (m *MailService) DeleteAlias(id uint) error {
	alias, err := mailAliasRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return buserr.New("ErrMailAliasNotFound")
	}

	if err := mailutil.DeleteAlias(alias.Source, alias.Target); err != nil {
		global.LOG.Errorf("failed to delete alias %s from mailserver: %v", alias.Source, err)
	}

	return mailAliasRepo.DeleteBy(context.Background(), repo.WithByID(id))
}

// Setup

func (m *MailService) Setup(req request.MailSetup) error {
	if req.Enable {
		if err := mailutil.EnsureMailContainers(); err != nil {
			return buserr.WithDetail("ErrMailContainerFailed", err.Error(), err)
		}
		m.saveSetting("MailEnabled", "true")
		m.saveSetting("MailContainerName", mailutil.MailContainerName)
		m.saveSetting("RoundcubeURL", fmt.Sprintf("http://localhost:%s", mailutil.RoundcubePort))
	} else {
		_ = mailutil.StopMailContainers()
		m.saveSetting("MailEnabled", "false")
	}
	return nil
}

func (m *MailService) GetStatus() (response.MailStatus, error) {
	enabled, _ := settingRepo.GetValueByKey("MailEnabled")
	containerName, _ := settingRepo.GetValueByKey("MailContainerName")
	roundcubeURL, _ := settingRepo.GetValueByKey("RoundcubeURL")

	status := mailutil.GetContainerStatus()
	return response.MailStatus{
		Enabled:           enabled == "true",
		MailContainerUp:   status.MailRunning,
		MailContainerName: containerName,
		RoundcubeUp:       status.RoundcubeRunning,
		RoundcubeURL:      roundcubeURL,
		Version:           status.MailVersion,
	}, nil
}

// Helpers

func (m *MailService) saveSetting(key, value string) {
	_ = settingRepo.UpdateOrCreate(key, value)
}

func (m *MailService) autoCreateDnsRecords(domain *model.MailDomain) {
	dnsEnabled, _ := settingRepo.GetValueByKey("DnsEnabled")
	if dnsEnabled != "true" {
		return
	}
	serverIP, _ := settingRepo.GetValueByKey("SystemIP")
	dnsSvc := NewIDnsService()

	// MX record
	_ = dnsSvc.AutoCreateMXRecords(domain.Name, "mail."+domain.Name)

	// A record for mail subdomain
	if serverIP != "" {
		_ = dnsSvc.AutoCreateARecord("mail."+domain.Name, serverIP)
	}

	// SPF TXT record
	zone, _ := dnsSvc.FindZoneForDomain(domain.Name)
	if zone != nil {
		_ = dnsSvc.CreateRecord(request.DnsRecordCreate{
			ZoneID:  zone.ID,
			Name:    domain.Name,
			Type:    "TXT",
			Content: "\"v=spf1 mx a ~all\"",
		})

		// DKIM TXT record
		if domain.DkimKey != "" {
			_ = dnsSvc.CreateRecord(request.DnsRecordCreate{
				ZoneID:  zone.ID,
				Name:    "mail._domainkey." + domain.Name,
				Type:    "TXT",
				Content: domain.DkimKey,
			})
		}

		// DMARC TXT record
		_ = dnsSvc.CreateRecord(request.DnsRecordCreate{
			ZoneID:  zone.ID,
			Name:    "_dmarc." + domain.Name,
			Type:    "TXT",
			Content: "\"v=DMARC1; p=quarantine; rua=mailto:postmaster@" + domain.Name + "\"",
		})
	}
}

