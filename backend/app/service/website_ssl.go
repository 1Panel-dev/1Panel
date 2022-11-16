package service

import (
	"context"
	"crypto/x509"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
	"path"
	"strings"
)

type WebSiteSSLService struct {
}

func (w WebSiteSSLService) Page(search dto.PageInfo) (int64, []dto.WebsiteSSLDTO, error) {
	total, sslList, err := websiteSSLRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var sslDTOs []dto.WebsiteSSLDTO
	for _, ssl := range sslList {
		sslDTOs = append(sslDTOs, dto.WebsiteSSLDTO{
			WebSiteSSL: ssl,
		})
	}
	return total, sslDTOs, err
}

func (w WebSiteSSLService) Create(create dto.WebsiteSSLCreate) (dto.WebsiteSSLCreate, error) {

	var res dto.WebsiteSSLCreate
	acmeAccount, err := websiteAcmeRepo.GetFirst(commonRepo.WithByID(create.AcmeAccountID))
	if err != nil {
		return res, err
	}
	dnsAccount, err := websiteDnsRepo.GetFirst(commonRepo.WithByID(create.DnsAccountID))
	if err != nil {
		return res, err
	}
	client, err := ssl.NewPrivateKeyClient(acmeAccount.Email, acmeAccount.PrivateKey)
	if err != nil {
		return res, err
	}
	if create.Provider == dto.Http {

	} else {
		if err := client.UseDns(ssl.DnsType(dnsAccount.Type), dnsAccount.Authorization); err != nil {
			return res, err
		}
	}

	resource, err := client.GetSSL(create.Domains)
	if err != nil {
		return res, err
	}

	var websiteSSL model.WebSiteSSL

	//TODO 判断同一个账号下的证书
	websiteSSL.Alias = create.Domains[0]
	websiteSSL.Domain = strings.Join(create.Domains, ",")
	websiteSSL.PrivateKey = string(resource.PrivateKey)
	websiteSSL.Pem = string(resource.Certificate)
	websiteSSL.CertURL = resource.CertURL

	cert, err := x509.ParseCertificate([]byte(websiteSSL.Pem))
	if err != nil {
		return dto.WebsiteSSLCreate{}, err
	}
	websiteSSL.ExpireDate = cert.NotAfter
	websiteSSL.StartDate = cert.NotBefore
	websiteSSL.Type = cert.Issuer.CommonName
	websiteSSL.IssuerName = cert.Issuer.Organization[0]

	if err := createPemFile(websiteSSL); err != nil {
		return dto.WebsiteSSLCreate{}, err
	}

	if err := websiteSSLRepo.Create(context.TODO(), &websiteSSL); err != nil {
		return res, err
	}

	return create, nil
}

func (w WebSiteSSLService) Apply(apply dto.WebsiteSSLApply) (dto.WebsiteSSLApply, error) {
	websiteSSL, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(apply.SSLID))
	if err != nil {
		return dto.WebsiteSSLApply{}, err
	}
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(apply.WebsiteID))
	if err != nil {
		return dto.WebsiteSSLApply{}, err
	}
	if err := createPemFile(websiteSSL); err != nil {
		return dto.WebsiteSSLApply{}, err
	}
	nginxParams := getNginxParamsFromStaticFile(dto.SSL)
	for i, param := range nginxParams {
		if param.Name == "ssl_certificate" {
			nginxParams[i].Params = []string{path.Join("/etc/nginx/ssl", websiteSSL.Alias, "fullchain.pem")}
		}
		if param.Name == "ssl_certificate_key" {
			nginxParams[i].Params = []string{path.Join("/etc/nginx/ssl", websiteSSL.Alias, "privkey.pem")}
		}
	}
	if err := updateNginxConfig(website, nginxParams, dto.SSL); err != nil {
		return dto.WebsiteSSLApply{}, err
	}
	website.WebSiteSSLID = websiteSSL.ID
	if err := websiteRepo.Save(context.TODO(), &website); err != nil {
		return dto.WebsiteSSLApply{}, err
	}

	return apply, nil
}

func (w WebSiteSSLService) GetDNSResolve(req dto.WebsiteDNSReq) (dto.WebsiteDNSRes, error) {
	acmeAccount, err := websiteAcmeRepo.GetFirst(commonRepo.WithByID(req.AcmeAccountID))
	if err != nil {
		return dto.WebsiteDNSRes{}, err
	}

	client, err := ssl.NewPrivateKeyClient(acmeAccount.Email, acmeAccount.PrivateKey)
	if err != nil {
		return dto.WebsiteDNSRes{}, err
	}
	re, err := client.UseManualDns(req.Domains)
	if err != nil {
		return dto.WebsiteDNSRes{}, err
	}
	var res dto.WebsiteDNSRes
	res.Key = re.Key
	res.Value = re.Value
	res.Type = "TXT"
	return res, nil
}

func (w WebSiteSSLService) Delete(id uint) error {
	return websiteSSLRepo.DeleteBy(commonRepo.WithByID(id))
}
