package service

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
	"github.com/go-acme/lego/v4/certcrypto"
	legoLogger "github.com/go-acme/lego/v4/log"
	"github.com/jinzhu/gorm"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type WebsiteSSLService struct {
}

type IWebsiteSSLService interface {
	Page(search request.WebsiteSSLSearch) (int64, []response.WebsiteSSLDTO, error)
	GetSSL(id uint) (*response.WebsiteSSLDTO, error)
	Search(req request.WebsiteSSLSearch) ([]response.WebsiteSSLDTO, error)
	Create(create request.WebsiteSSLCreate) (request.WebsiteSSLCreate, error)
	GetDNSResolve(req request.WebsiteDNSReq) ([]response.WebsiteDNSRes, error)
	GetWebsiteSSL(websiteId uint) (response.WebsiteSSLDTO, error)
	Delete(ids []uint) error
	Update(update request.WebsiteSSLUpdate) error
	Upload(req request.WebsiteSSLUpload) error
	ObtainSSL(apply request.WebsiteSSLApply) error
	SyncForRestart() error
	DownloadFile(id uint) (*os.File, error)
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
	for _, model := range sslList {
		result = append(result, response.WebsiteSSLDTO{
			WebsiteSSL: model,
			LogPath:    path.Join(constant.SSLLogDir, fmt.Sprintf("%s-ssl-%d.log", model.PrimaryDomain, model.ID)),
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
	res.WebsiteSSL = *websiteSSL
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
	if create.Nameserver1 != "" && !common.IsValidIP(create.Nameserver1) {
		return create, buserr.New("ErrParseIP")
	}
	if create.Nameserver2 != "" && !common.IsValidIP(create.Nameserver2) {
		return create, buserr.New("ErrParseIP")
	}
	var res request.WebsiteSSLCreate
	acmeAccount, err := websiteAcmeRepo.GetFirst(commonRepo.WithByID(create.AcmeAccountID))
	if err != nil {
		return res, err
	}
	websiteSSL := model.WebsiteSSL{
		Status:        constant.SSLInit,
		Provider:      create.Provider,
		AcmeAccountID: acmeAccount.ID,
		PrimaryDomain: create.PrimaryDomain,
		ExpireDate:    time.Now(),
		KeyType:       create.KeyType,
		PushDir:       create.PushDir,
		Description:   create.Description,
		Nameserver1:   create.Nameserver1,
		Nameserver2:   create.Nameserver2,
		SkipDNS:       create.SkipDNS,
		DisableCNAME:  create.DisableCNAME,
		ExecShell:     create.ExecShell,
	}
	if create.ExecShell {
		websiteSSL.Shell = create.Shell
	}
	if create.PushDir {
		fileOP := files.NewFileOp()
		if !fileOP.Stat(create.Dir) {
			_ = fileOP.CreateDir(create.Dir, 0755)
		}
		websiteSSL.Dir = create.Dir
	}

	var domains []string
	if create.OtherDomains != "" {
		otherDomainArray := strings.Split(create.OtherDomains, "\n")
		for _, domain := range otherDomainArray {
			if !common.IsValidDomain(domain) {
				err = buserr.WithName("ErrDomainFormat", domain)
				return res, err
			}
			domains = append(domains, domain)
		}
	}
	websiteSSL.Domains = strings.Join(domains, ",")

	if create.Provider == constant.DNSAccount || create.Provider == constant.Http {
		websiteSSL.AutoRenew = create.AutoRenew
	}
	if create.Provider == constant.DNSAccount {
		dnsAccount, err := websiteDnsRepo.GetFirst(commonRepo.WithByID(create.DnsAccountID))
		if err != nil {
			return res, err
		}
		websiteSSL.DnsAccountID = dnsAccount.ID
	}

	if err := websiteSSLRepo.Create(context.TODO(), &websiteSSL); err != nil {
		return res, err
	}
	create.ID = websiteSSL.ID
	go func() {
		if create.Provider != constant.DnsManual {
			if err = w.ObtainSSL(request.WebsiteSSLApply{
				ID: websiteSSL.ID,
			}); err != nil {
				global.LOG.Errorf("obtain ssl failed, err: %v", err)
			}
		}
	}()
	return create, nil
}

func printSSLLog(logger *log.Logger, msgKey string, params map[string]interface{}, disableLog bool) {
	if disableLog {
		return
	}
	logger.Println(i18n.GetMsgWithMap(msgKey, params))
}

func reloadSystemSSL(websiteSSL *model.WebsiteSSL, logger *log.Logger) {
	systemSSLEnable, sslID := GetSystemSSL()
	if systemSSLEnable && sslID == websiteSSL.ID {
		fileOp := files.NewFileOp()
		certPath := path.Join(global.CONF.System.BaseDir, "1panel/secret/server.crt")
		keyPath := path.Join(global.CONF.System.BaseDir, "1panel/secret/server.key")
		printSSLLog(logger, "StartUpdateSystemSSL", nil, logger == nil)
		if err := fileOp.WriteFile(certPath, strings.NewReader(websiteSSL.Pem), 0600); err != nil {
			logger.Printf("Failed to update the SSL certificate File for 1Panel System domain [%s] , err:%s", websiteSSL.PrimaryDomain, err.Error())
			return
		}
		if err := fileOp.WriteFile(keyPath, strings.NewReader(websiteSSL.PrivateKey), 0600); err != nil {
			logger.Printf("Failed to update the SSL certificate for 1Panel System domain [%s] , err:%s", websiteSSL.PrimaryDomain, err.Error())
			return
		}
		newCert, err := tls.X509KeyPair([]byte(websiteSSL.Pem), []byte(websiteSSL.PrivateKey))
		if err != nil {
			logger.Printf("Failed to update the SSL certificate for 1Panel System domain [%s] , err:%s", websiteSSL.PrimaryDomain, err.Error())
			return
		}
		printSSLLog(logger, "UpdateSystemSSLSuccess", nil, logger == nil)
		constant.CertStore.Store(&newCert)
	}
}

func (w WebsiteSSLService) ObtainSSL(apply request.WebsiteSSLApply) error {
	var (
		err         error
		websiteSSL  *model.WebsiteSSL
		acmeAccount *model.WebsiteAcmeAccount
		dnsAccount  *model.WebsiteDnsAccount
	)

	websiteSSL, err = websiteSSLRepo.GetFirst(commonRepo.WithByID(apply.ID))
	if err != nil {
		return err
	}
	acmeAccount, err = websiteAcmeRepo.GetFirst(commonRepo.WithByID(websiteSSL.AcmeAccountID))
	if err != nil {
		return err
	}
	client, err := ssl.NewAcmeClient(acmeAccount)
	if err != nil {
		return err
	}

	switch websiteSSL.Provider {
	case constant.DNSAccount:
		dnsAccount, err = websiteDnsRepo.GetFirst(commonRepo.WithByID(websiteSSL.DnsAccountID))
		if err != nil {
			return err
		}
		if err = client.UseDns(ssl.DnsType(dnsAccount.Type), dnsAccount.Authorization, *websiteSSL); err != nil {
			return err
		}
	case constant.Http:
		appInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return buserr.New("ErrOpenrestyNotFound")
			}
			return err
		}
		if err := client.UseHTTP(path.Join(appInstall.GetPath(), "root")); err != nil {
			return err
		}
	case constant.DnsManual:
		if err := client.UseManualDns(); err != nil {
			return err
		}
	}

	domains := []string{websiteSSL.PrimaryDomain}
	if websiteSSL.Domains != "" {
		domains = append(domains, strings.Split(websiteSSL.Domains, ",")...)
	}

	var privateKey crypto.PrivateKey
	if websiteSSL.PrivateKey == "" {
		privateKey, err = certcrypto.GeneratePrivateKey(ssl.KeyType(websiteSSL.KeyType))
		if err != nil {
			return err
		}
	} else {
		block, _ := pem.Decode([]byte(websiteSSL.PrivateKey))
		if block == nil {
			return buserr.New("invalid PEM block")
		}
		var privKey crypto.PrivateKey
		keyType := ssl.KeyType(websiteSSL.KeyType)
		switch keyType {
		case certcrypto.EC256, certcrypto.EC384:
			privKey, err = x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return nil
			}
		case certcrypto.RSA2048, certcrypto.RSA3072, certcrypto.RSA4096:
			privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil
			}
		}
		privateKey = privKey
	}

	websiteSSL.Status = constant.SSLApply
	err = websiteSSLRepo.Save(websiteSSL)
	if err != nil {
		return err
	}

	go func() {
		logFile, _ := os.OpenFile(path.Join(constant.SSLLogDir, fmt.Sprintf("%s-ssl-%d.log", websiteSSL.PrimaryDomain, websiteSSL.ID)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer logFile.Close()
		logger := log.New(logFile, "", log.LstdFlags)
		legoLogger.Logger = logger
		if !apply.DisableLog {
			startMsg := i18n.GetMsgWithMap("ApplySSLStart", map[string]interface{}{"domain": strings.Join(domains, ","), "type": i18n.GetMsgByKey(websiteSSL.Provider)})
			if websiteSSL.Provider == constant.DNSAccount {
				startMsg = startMsg + i18n.GetMsgWithMap("DNSAccountName", map[string]interface{}{"name": dnsAccount.Name, "type": dnsAccount.Type})
			}
			legoLogger.Logger.Println(startMsg)
		}
		resource, err := client.ObtainSSL(domains, privateKey)
		if err != nil {
			handleError(websiteSSL, err)
			return
		}
		websiteSSL.PrivateKey = string(resource.PrivateKey)
		websiteSSL.Pem = string(resource.Certificate)
		websiteSSL.CertURL = resource.CertURL
		certBlock, _ := pem.Decode(resource.Certificate)
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			handleError(websiteSSL, err)
			return
		}
		websiteSSL.ExpireDate = cert.NotAfter
		websiteSSL.StartDate = cert.NotBefore
		websiteSSL.Type = cert.Issuer.CommonName
		websiteSSL.Organization = cert.Issuer.Organization[0]
		websiteSSL.Status = constant.SSLReady
		printSSLLog(logger, "ApplySSLSuccess", map[string]interface{}{"domain": strings.Join(domains, ",")}, apply.DisableLog)
		saveCertificateFile(websiteSSL, logger)

		if websiteSSL.ExecShell {
			workDir := constant.DataDir
			if websiteSSL.PushDir {
				workDir = websiteSSL.Dir
			}
			printSSLLog(logger, "ExecShellStart", nil, apply.DisableLog)
			if err = cmd.ExecShellWithTimeOut(websiteSSL.Shell, workDir, logger, 30*time.Minute); err != nil {
				printSSLLog(logger, "ErrExecShell", map[string]interface{}{"err": err.Error()}, apply.DisableLog)
			} else {
				printSSLLog(logger, "ExecShellSuccess", nil, apply.DisableLog)
			}
		}

		err = websiteSSLRepo.Save(websiteSSL)
		if err != nil {
			return
		}

		websites, _ := websiteRepo.GetBy(websiteRepo.WithWebsiteSSLID(websiteSSL.ID))
		if len(websites) > 0 {
			for _, website := range websites {
				printSSLLog(logger, "ApplyWebSiteSSLLog", map[string]interface{}{"name": website.PrimaryDomain}, apply.DisableLog)
				if err := createPemFile(website, *websiteSSL); err != nil {
					printSSLLog(logger, "ErrUpdateWebsiteSSL", map[string]interface{}{"name": website.PrimaryDomain, "err": err.Error()}, apply.DisableLog)
				}
			}
			nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
			if err != nil {
				return
			}
			if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
				printSSLLog(logger, constant.ErrSSLApply, nil, apply.DisableLog)
				return
			}
			printSSLLog(logger, "ApplyWebSiteSSLSuccess", nil, apply.DisableLog)
		}

		reloadSystemSSL(websiteSSL, logger)
	}()

	return nil
}

func handleError(websiteSSL *model.WebsiteSSL, err error) {
	if websiteSSL.Status == constant.SSLInit || websiteSSL.Status == constant.SSLError {
		websiteSSL.Status = constant.Error
	} else {
		websiteSSL.Status = constant.SSLApplyError
	}
	websiteSSL.Message = err.Error()
	legoLogger.Logger.Println(i18n.GetErrMsg("ApplySSLFailed", map[string]interface{}{"domain": websiteSSL.PrimaryDomain, "detail": err.Error()}))
	_ = websiteSSLRepo.Save(websiteSSL)
}

func (w WebsiteSSLService) GetDNSResolve(req request.WebsiteDNSReq) ([]response.WebsiteDNSRes, error) {
	acmeAccount, err := websiteAcmeRepo.GetFirst(commonRepo.WithByID(req.AcmeAccountID))
	if err != nil {
		return nil, err
	}

	client, err := ssl.NewAcmeClient(acmeAccount)
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
	res.WebsiteSSL = *websiteSSL
	return res, nil
}

func (w WebsiteSSLService) Delete(ids []uint) error {
	var names []string
	for _, id := range ids {
		if websites, _ := websiteRepo.GetBy(websiteRepo.WithWebsiteSSLID(id)); len(websites) > 0 {
			oldSSL, _ := websiteSSLRepo.GetFirst(commonRepo.WithByID(id))
			if oldSSL.ID > 0 {
				names = append(names, oldSSL.PrimaryDomain)
			}
			continue
		}
		sslSetting, _ := settingRepo.Get(settingRepo.WithByKey("SSL"))
		if sslSetting.Value == "enable" {
			sslID, _ := settingRepo.Get(settingRepo.WithByKey("SSLID"))
			idValue, _ := strconv.Atoi(sslID.Value)
			if idValue > 0 && uint(idValue) == id {
				return buserr.New("ErrDeleteWithPanelSSL")
			}
		}
		_ = websiteSSLRepo.DeleteBy(commonRepo.WithByID(id))
	}
	if len(names) > 0 {
		return buserr.WithName("ErrSSLCannotDelete", strings.Join(names, ","))
	}
	return nil
}

func (w WebsiteSSLService) Update(update request.WebsiteSSLUpdate) error {
	websiteSSL, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(update.ID))
	if err != nil {
		return err
	}
	updateParams := make(map[string]interface{})
	updateParams["primary_domain"] = update.PrimaryDomain
	updateParams["description"] = update.Description
	updateParams["provider"] = update.Provider
	updateParams["push_dir"] = update.PushDir
	updateParams["disable_cname"] = update.DisableCNAME
	updateParams["skip_dns"] = update.SkipDNS
	updateParams["nameserver1"] = update.Nameserver1
	updateParams["nameserver2"] = update.Nameserver2
	updateParams["exec_shell"] = update.ExecShell
	if update.ExecShell {
		updateParams["shell"] = update.Shell
	} else {
		updateParams["shell"] = ""
	}

	if websiteSSL.Provider != constant.SelfSigned {
		acmeAccount, err := websiteAcmeRepo.GetFirst(commonRepo.WithByID(update.AcmeAccountID))
		if err != nil {
			return err
		}
		updateParams["acme_account_id"] = acmeAccount.ID
	}

	if update.PushDir {
		fileOP := files.NewFileOp()
		if !fileOP.Stat(update.Dir) {
			_ = fileOP.CreateDir(update.Dir, 0755)
		}
		updateParams["dir"] = update.Dir
	}
	var domains []string
	if update.OtherDomains != "" {
		otherDomainArray := strings.Split(update.OtherDomains, "\n")
		for _, domain := range otherDomainArray {
			if !common.IsValidDomain(domain) {
				return buserr.WithName("ErrDomainFormat", domain)
			}
			domains = append(domains, domain)
		}
	}
	updateParams["domains"] = strings.Join(domains, ",")
	if update.Provider == constant.DNSAccount || update.Provider == constant.Http || update.Provider == constant.SelfSigned {
		updateParams["auto_renew"] = update.AutoRenew
	} else {
		updateParams["auto_renew"] = false
	}
	if update.Provider == constant.DNSAccount {
		dnsAccount, err := websiteDnsRepo.GetFirst(commonRepo.WithByID(update.DnsAccountID))
		if err != nil {
			return err
		}
		updateParams["dns_account_id"] = dnsAccount.ID
	} else {
		updateParams["dns_account_id"] = 0
	}
	return websiteSSLRepo.SaveByMap(websiteSSL, updateParams)
}

func (w WebsiteSSLService) Upload(req request.WebsiteSSLUpload) error {
	websiteSSL := &model.WebsiteSSL{
		Provider:    constant.Manual,
		Description: req.Description,
		Status:      constant.SSLReady,
	}
	var err error
	if req.SSLID > 0 {
		websiteSSL, err = websiteSSLRepo.GetFirst(commonRepo.WithByID(req.SSLID))
		if err != nil {
			return err
		}
		websiteSSL.Description = req.Description
	}
	if req.Type == "local" {
		fileOp := files.NewFileOp()
		if !fileOp.Stat(req.PrivateKeyPath) {
			return buserr.New("ErrSSLKeyNotFound")
		}
		if !fileOp.Stat(req.CertificatePath) {
			return buserr.New("ErrSSLCertificateNotFound")
		}
		if content, err := fileOp.GetContent(req.PrivateKeyPath); err != nil {
			return err
		} else {
			websiteSSL.PrivateKey = string(content)
		}
		if content, err := fileOp.GetContent(req.CertificatePath); err != nil {
			return err
		} else {
			websiteSSL.Pem = string(content)
		}
	} else {
		websiteSSL.PrivateKey = req.PrivateKey
		websiteSSL.Pem = req.Certificate
	}

	privateKeyCertBlock, _ := pem.Decode([]byte(websiteSSL.PrivateKey))
	if privateKeyCertBlock == nil {
		return buserr.New("ErrSSLKeyFormat")
	}

	var (
		cert    *x509.Certificate
		pemData = []byte(websiteSSL.Pem)
	)
	for {
		certBlock, reset := pem.Decode(pemData)
		if certBlock == nil {
			break
		}
		cert, err = x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return err
		}
		if len(cert.DNSNames) > 0 || len(cert.IPAddresses) > 0 {
			break
		}
		pemData = reset
	}
	if pemData == nil {
		return buserr.New("ErrSSLCertificateFormat")
	}

	websiteSSL.ExpireDate = cert.NotAfter
	websiteSSL.StartDate = cert.NotBefore
	websiteSSL.Type = cert.Issuer.CommonName
	if len(cert.Issuer.Organization) > 0 {
		websiteSSL.Organization = cert.Issuer.Organization[0]
	} else {
		websiteSSL.Organization = cert.Issuer.CommonName
	}

	var domains []string
	if len(cert.DNSNames) > 0 {
		websiteSSL.PrimaryDomain = cert.DNSNames[0]
		domains = cert.DNSNames[1:]
	}
	if len(cert.IPAddresses) > 0 {
		if websiteSSL.PrimaryDomain == "" {
			websiteSSL.PrimaryDomain = cert.IPAddresses[0].String()
			for _, ip := range cert.IPAddresses[1:] {
				domains = append(domains, ip.String())
			}
		} else {
			for _, ip := range cert.IPAddresses {
				domains = append(domains, ip.String())
			}
		}
	}
	websiteSSL.Domains = strings.Join(domains, ",")

	if websiteSSL.ID > 0 {
		if err := UpdateSSLConfig(*websiteSSL); err != nil {
			return err
		}
		return websiteSSLRepo.Save(websiteSSL)
	}
	return websiteSSLRepo.Create(context.Background(), websiteSSL)
}

func (w WebsiteSSLService) DownloadFile(id uint) (*os.File, error) {
	websiteSSL, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	fileOp := files.NewFileOp()
	dir := path.Join(global.CONF.System.BaseDir, "1panel/tmp/ssl", websiteSSL.PrimaryDomain)
	if fileOp.Stat(dir) {
		if err = fileOp.DeleteDir(dir); err != nil {
			return nil, err
		}
	}
	if err = fileOp.CreateDir(dir, 0666); err != nil {
		return nil, err
	}
	if err = fileOp.WriteFile(path.Join(dir, "fullchain.pem"), strings.NewReader(websiteSSL.Pem), 0644); err != nil {
		return nil, err
	}
	if err = fileOp.WriteFile(path.Join(dir, "privkey.pem"), strings.NewReader(websiteSSL.PrivateKey), 0644); err != nil {
		return nil, err
	}
	fileName := websiteSSL.PrimaryDomain + ".zip"
	if err = fileOp.Compress([]string{path.Join(dir, "fullchain.pem"), path.Join(dir, "privkey.pem")}, dir, fileName, files.SdkZip, ""); err != nil {
		return nil, err
	}
	return os.Open(path.Join(dir, fileName))
}

func (w WebsiteSSLService) SyncForRestart() error {
	sslList, err := websiteSSLRepo.List()
	if err != nil {
		return err
	}
	for _, ssl := range sslList {
		if ssl.Status == constant.SSLApply {
			ssl.Status = constant.SystemRestart
			ssl.Message = "System restart causing interrupt"
			_ = websiteSSLRepo.Save(&ssl)
		}
	}
	return nil
}
