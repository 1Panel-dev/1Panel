package service

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
	"path"
	"strconv"
	"strings"
)

type WebsiteSSLService struct {
}

type IWebsiteSSLService interface {
	Page(search request.WebsiteSSLSearch) (int64, []response.WebsiteSSLDTO, error)
	GetSSL(id uint) (*response.WebsiteSSLDTO, error)
	Search(req request.WebsiteSSLSearch) ([]response.WebsiteSSLDTO, error)
	Create(create request.WebsiteSSLCreate) (request.WebsiteSSLCreate, error)
	Renew(sslId uint) error
	GetDNSResolve(req request.WebsiteDNSReq) ([]response.WebsiteDNSRes, error)
	GetWebsiteSSL(websiteId uint) (response.WebsiteSSLDTO, error)
	Delete(id uint) error
	Update(update request.WebsiteSSLUpdate) error
}

func NewIWebsiteSSLService() IWebsiteSSLService {
	return &WebsiteSSLService{}
}

func (w WebsiteSSLService) Page(search request.WebsiteSSLSearch) (int64, []response.WebsiteSSLDTO, error) {
	var (
		result []response.WebsiteSSLDTO
	)
	total, sslList, err := websiteSSLRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	if err != nil {
		return 0, nil, err
	}
	for _, sslModel := range sslList {
		result = append(result, response.WebsiteSSLDTO{
			WebsiteSSL: sslModel,
		})
	}
	return total, result, err
}

func (w WebsiteSSLService) GetSSL(id uint) (*response.WebsiteSSLDTO, error) {
	var res response.WebsiteSSLDTO
	websiteSSL, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	res.WebsiteSSL = websiteSSL
	return &res, nil
}

func (w WebsiteSSLService) Search(search request.WebsiteSSLSearch) ([]response.WebsiteSSLDTO, error) {
	var (
		opts   []repo.DBOption
		result []response.WebsiteSSLDTO
	)
	opts = append(opts, commonRepo.WithOrderBy("created_at desc"))
	if search.AcmeAccountID != "" {
		acmeAccountID, err := strconv.ParseUint(search.AcmeAccountID, 10, 64)
		if err != nil {
			return nil, err
		}
		opts = append(opts, websiteSSLRepo.WithByAcmeAccountId(uint(acmeAccountID)))
	}
	sslList, err := websiteSSLRepo.List(opts...)
	if err != nil {
		return nil, err
	}
	for _, sslModel := range sslList {
		result = append(result, response.WebsiteSSLDTO{
			WebsiteSSL: sslModel,
		})
	}
	return result, err
}

func (w WebsiteSSLService) Create(create request.WebsiteSSLCreate) (request.WebsiteSSLCreate, error) {
	var res request.WebsiteSSLCreate
	acmeAccount, err := websiteAcmeRepo.GetFirst(commonRepo.WithByID(create.AcmeAccountID))
	if err != nil {
		return res, err
	}
	client, err := ssl.NewPrivateKeyClient(acmeAccount.Email, acmeAccount.PrivateKey)
	if err != nil {
		return res, err
	}

	var websiteSSL model.WebsiteSSL

	switch create.Provider {
	case constant.DNSAccount:
		dnsAccount, err := websiteDnsRepo.GetFirst(commonRepo.WithByID(create.DnsAccountID))
		if err != nil {
			return res, err
		}
		if err := client.UseDns(ssl.DnsType(dnsAccount.Type), dnsAccount.Authorization); err != nil {
			return res, err
		}
		websiteSSL.AutoRenew = create.AutoRenew
	case constant.Http:
		appInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return request.WebsiteSSLCreate{}, err
		}
		if err := client.UseHTTP(path.Join(constant.AppInstallDir, constant.AppOpenresty, appInstall.Name, "root")); err != nil {
			return res, err
		}
		websiteSSL.AutoRenew = create.AutoRenew
	case constant.DnsManual:
		if err := client.UseManualDns(); err != nil {
			return res, err
		}
	}

	domains := []string{create.PrimaryDomain}
	otherDomainArray := strings.Split(create.OtherDomains, "\n")
	if create.OtherDomains != "" {
		domains = append(otherDomainArray, domains...)
	}
	resource, err := client.ObtainSSL(domains)
	if err != nil {
		return res, err
	}

	websiteSSL.DnsAccountID = create.DnsAccountID
	websiteSSL.AcmeAccountID = acmeAccount.ID
	websiteSSL.Provider = create.Provider
	websiteSSL.Domains = strings.Join(otherDomainArray, ",")
	websiteSSL.PrimaryDomain = create.PrimaryDomain
	websiteSSL.PrivateKey = string(resource.PrivateKey)
	websiteSSL.Pem = string(resource.Certificate)
	websiteSSL.CertURL = resource.CertURL
	certBlock, _ := pem.Decode(resource.Certificate)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return request.WebsiteSSLCreate{}, err
	}
	websiteSSL.ExpireDate = cert.NotAfter
	websiteSSL.StartDate = cert.NotBefore
	websiteSSL.Type = cert.Issuer.CommonName
	websiteSSL.Organization = cert.Issuer.Organization[0]

	if err := websiteSSLRepo.Create(context.TODO(), &websiteSSL); err != nil {
		return res, err
	}

	return create, nil
}

func (w WebsiteSSLService) Renew(sslId uint) error {
	websiteSSL, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(sslId))
	if err != nil {
		return err
	}
	acmeAccount, err := websiteAcmeRepo.GetFirst(commonRepo.WithByID(websiteSSL.AcmeAccountID))
	if err != nil {
		return err
	}

	client, err := ssl.NewPrivateKeyClient(acmeAccount.Email, acmeAccount.PrivateKey)
	if err != nil {
		return err
	}
	switch websiteSSL.Provider {
	case constant.DNSAccount:
		dnsAccount, err := websiteDnsRepo.GetFirst(commonRepo.WithByID(websiteSSL.DnsAccountID))
		if err != nil {
			return err
		}
		if err := client.UseDns(ssl.DnsType(dnsAccount.Type), dnsAccount.Authorization); err != nil {
			return err
		}
	case constant.Http:
		appInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return err
		}
		if err := client.UseHTTP(path.Join(constant.AppInstallDir, constant.AppOpenresty, appInstall.Name, "root")); err != nil {
			return err
		}
	}

	resource, err := client.RenewSSL(websiteSSL.CertURL)
	if err != nil {
		return err
	}
	websiteSSL.PrivateKey = string(resource.PrivateKey)
	websiteSSL.Pem = string(resource.Certificate)
	websiteSSL.CertURL = resource.CertURL
	certBlock, _ := pem.Decode(resource.Certificate)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return err
	}
	websiteSSL.ExpireDate = cert.NotAfter
	websiteSSL.StartDate = cert.NotBefore
	websiteSSL.Type = cert.Issuer.CommonName
	websiteSSL.Organization = cert.Issuer.Organization[0]

	if err := websiteSSLRepo.Save(websiteSSL); err != nil {
		return err
	}

	websites, _ := websiteRepo.GetBy(websiteRepo.WithWebsiteSSLID(sslId))
	for _, website := range websites {
		if err := createPemFile(website, websiteSSL); err != nil {
			global.LOG.Errorf("create website [%s] ssl file failed! err:%s", website.PrimaryDomain, err.Error())
		}
	}
	if len(websites) > 0 {
		nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return err
		}
		if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
			return buserr.New(constant.ErrSSLApply)
		}
	}
	return nil
}

func (w WebsiteSSLService) GetDNSResolve(req request.WebsiteDNSReq) ([]response.WebsiteDNSRes, error) {
	acmeAccount, err := websiteAcmeRepo.GetFirst(commonRepo.WithByID(req.AcmeAccountID))
	if err != nil {
		return nil, err
	}

	client, err := ssl.NewPrivateKeyClient(acmeAccount.Email, acmeAccount.PrivateKey)
	if err != nil {
		return nil, err
	}
	resolves, err := client.GetDNSResolve(req.Domains)
	if err != nil {
		return nil, err
	}
	var res []response.WebsiteDNSRes
	for k, v := range resolves {
		res = append(res, response.WebsiteDNSRes{
			Domain: k,
			Key:    v.Key,
			Value:  v.Value,
			Err:    v.Err,
		})
	}
	return res, nil
}

func (w WebsiteSSLService) GetWebsiteSSL(websiteId uint) (response.WebsiteSSLDTO, error) {
	var res response.WebsiteSSLDTO
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(websiteId))
	if err != nil {
		return res, err
	}
	websiteSSL, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(website.WebsiteSSLID))
	if err != nil {
		return res, err
	}
	res.WebsiteSSL = websiteSSL
	return res, nil
}

func (w WebsiteSSLService) Delete(id uint) error {
	if websites, _ := websiteRepo.GetBy(websiteRepo.WithWebsiteSSLID(id)); len(websites) > 0 {
		return buserr.New(constant.ErrSSLCannotDelete)
	}
	return websiteSSLRepo.DeleteBy(commonRepo.WithByID(id))
}

func (w WebsiteSSLService) Update(update request.WebsiteSSLUpdate) error {
	websiteSSL, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(update.ID))
	if err != nil {
		return err
	}
	websiteSSL.AutoRenew = update.AutoRenew
	return websiteSSLRepo.Save(websiteSSL)
}
