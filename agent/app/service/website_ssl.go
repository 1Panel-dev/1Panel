package service

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/certificate"
	legoLogger "github.com/go-acme/lego/v5/log"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/ssl"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	gormv2 "gorm.io/gorm"
)

type WebsiteSSLService struct {
}

const sslObtainTimeout = time.Hour

var legoLogMu sync.Mutex

func withLegoLogger(logger *log.Logger, fn func() error) error {
	if logger == nil {
		logger = log.New(io.Discard, "", log.LstdFlags)
	}

	legoLogMu.Lock()
	defer legoLogMu.Unlock()

	oldLogger := legoLogger.Default()
	legoLogger.SetDefault(slog.New(slog.NewTextHandler(logger.Writer(), nil)))
	defer legoLogger.SetDefault(oldLogger)

	return fn()
}

func withLegoLoggerTimeout(logger *log.Logger, timeout time.Duration, fn func(context.Context) error) error {
	return withLegoLogger(logger, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return fn(ctx)
	})
}

func newWebsiteSSLLegoClient(ctx context.Context, acmeAccount *model.WebsiteAcmeAccount) (*ssl.AcmeClient, error) {
	client, err := ssl.NewAcmeClientWithContext(ctx, acmeAccount, getSystemProxy(acmeAccount.UseProxy))
	if err == nil {
		return client, nil
	}
	if !errors.Is(err, ssl.ErrAcmeAccountURLMissing) {
		return nil, err
	}

	client, err = ssl.NewRegisterClientWithContext(ctx, acmeAccount, getSystemProxy(acmeAccount.UseProxy))
	if err != nil {
		return nil, err
	}
	if client.User.GetRegistration() == nil || client.User.GetRegistration().Location == "" {
		return nil, ssl.ErrAcmeAccountURLMissing
	}

	acmeAccount.URL = client.User.GetRegistration().Location
	if acmeAccount.PrivateKey == "" {
		return nil, errors.New("private key can not blank")
	}
	if err := websiteAcmeRepo.Save(*acmeAccount); err != nil {
		return nil, err
	}
	return client, nil
}

type IWebsiteSSLService interface {
	Page(search request.WebsiteSSLSearch) (int64, []response.WebsiteSSLDTO, error)
	GetSSL(id uint) (*response.WebsiteSSLDTO, error)
	Search(req request.WebsiteSSLListReq) ([]response.WebsiteSSLDTO, error)
	Create(create request.WebsiteSSLCreate) (request.WebsiteSSLCreate, error)
	GetDNSResolve(req request.WebsiteDNSReq) ([]response.WebsiteDNSRes, error)
	GetWebsiteSSL(websiteId uint) (response.WebsiteSSLDTO, error)
	Delete(ids []uint) error
	Update(update request.WebsiteSSLUpdate) error
	Upload(req request.WebsiteSSLUpload) error
	PushToNode(req request.WebsiteSSLPush) error
	ObtainSSL(apply request.WebsiteSSLApply) error
	AutoRenewSSL(id uint) error
	SyncForRestart() error
	DownloadFile(id uint) (*os.File, error)
	ImportMasterSSL(create model.WebsiteSSL) error
}

func NewIWebsiteSSLService() IWebsiteSSLService {
	return &WebsiteSSLService{}
}

func (w WebsiteSSLService) Page(search request.WebsiteSSLSearch) (int64, []response.WebsiteSSLDTO, error) {
	var (
		result []response.WebsiteSSLDTO
		opts   []repo.DBOption
	)
	if search.OrderBy != "" && search.Order != "null" {
		opts = append(opts, repo.WithOrderRuleBy(search.OrderBy, search.Order))
	} else {
		opts = append(opts, repo.WithOrderDesc("updated_at"))
	}
	if search.Domain != "" {
		opts = append(opts, websiteSSLRepo.WithByDomain(search.Domain))
	}
	total, sslList, err := websiteSSLRepo.Page(search.Page, search.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	for _, model := range sslList {
		result = append(result, response.WebsiteSSLDTO{
			WebsiteSSL: model,
			LogPath:    path.Join(global.Dir.SSLLogDir, fmt.Sprintf("%s-ssl-%d.log", model.PrimaryDomain, model.ID)),
		})
	}
	return total, result, err
}

func (w WebsiteSSLService) GetSSL(id uint) (*response.WebsiteSSLDTO, error) {
	var res response.WebsiteSSLDTO
	websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return nil, err
	}
	res.WebsiteSSL = *websiteSSL
	return &res, nil
}

func (w WebsiteSSLService) Search(search request.WebsiteSSLListReq) ([]response.WebsiteSSLDTO, error) {
	var (
		opts   []repo.DBOption
		result []response.WebsiteSSLDTO
	)
	opts = append(opts, repo.WithOrderDesc("updated_at"))
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
	acmeAccount, err := websiteAcmeRepo.GetFirst(repo.WithByID(create.AcmeAccountID))
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
		IsIp:          create.IsIp,
	}
	if create.ExecShell {
		websiteSSL.Shell = create.Shell
	}
	if create.PushDir {
		fileOP := files.NewFileOp()
		if !fileOP.Stat(create.Dir) {
			return res, buserr.New("ErrLinkPathNotFound")
		}
		websiteSSL.Dir = create.Dir
	}
	setSSLPushConfig(&websiteSSL, create.PushNode, create.Nodes)

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
	if create.Provider == constant.Http {
		if strings.Contains(create.PrimaryDomain, "*") {
			return res, buserr.New("ErrWildcardDomain")
		}
		for _, domain := range domains {
			if strings.Contains(domain, "*") {
				return res, buserr.New("ErrWildcardDomain")
			}
		}
	}
	websiteSSL.Domains = strings.Join(domains, ",")

	if create.Provider == constant.DNSAccount || create.Provider == constant.Http {
		websiteSSL.AutoRenew = create.AutoRenew
	}
	if create.Provider == constant.DNSAccount {
		dnsAccount, err := websiteDnsRepo.GetFirst(repo.WithByID(create.DnsAccountID))
		if err != nil {
			return res, err
		}
		websiteSSL.DnsAccountID = dnsAccount.ID
	}

	if err := websiteSSLRepo.Create(context.TODO(), &websiteSSL); err != nil {
		return res, err
	}
	create.ID = websiteSSL.ID
	logFile, err := os.OpenFile(path.Join(global.Dir.SSLLogDir, fmt.Sprintf("%s-ssl-%d.log", websiteSSL.PrimaryDomain, websiteSSL.ID)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		global.LOG.Errorf("open ssl log file failed, domain: %s, err: %v", websiteSSL.PrimaryDomain, err)
	} else {
		logFile.Close()
	}
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

func printSSLLog(logger *log.Logger, msgKey string, params map[string]interface{}) {
	if logger == nil {
		return
	}
	logger.Println(i18n.GetMsgWithMap(msgKey, params))
}

func normalizeSSLPushConfig(pushNode bool, nodes string) (bool, string) {
	nodes = strings.TrimSpace(nodes)
	if !pushNode || nodes == "" {
		return false, ""
	}
	return true, nodes
}

func setSSLPushConfig(websiteSSL *model.WebsiteSSL, pushNode bool, nodes string) {
	pushNode, nodes = normalizeSSLPushConfig(pushNode, nodes)
	if !global.IsMaster || !xpack.MultiNodeProvider.IsXpack() {
		pushNode = false
		nodes = ""
	}
	websiteSSL.PushNode = pushNode
	websiteSSL.Nodes = nodes
}

func pushSSLToNode(websiteSSL *model.WebsiteSSL, logger *log.Logger) error {
	printSSLLog(logger, "StartPushSSLToNode", nil)
	if err := xpack.MultiNodeProvider.PushSSLToNode(websiteSSL); err != nil {
		printSSLLog(logger, "PushSSLToNodeFailed", map[string]interface{}{"err": err.Error()})
		return err
	}
	printSSLLog(logger, "PushSSLToNodeSuccess", nil)
	return nil
}

func pushSSLToNodeWithNewLogger(websiteSSL *model.WebsiteSSL) error {
	if !websiteSSL.PushNode {
		return nil
	}
	logFile, logger := newWebsiteSSLLogger(websiteSSL, false)
	if logFile != nil {
		defer func() {
			_ = logFile.Close()
		}()
	}
	return pushSSLToNode(websiteSSL, logger)
}

func newWebsiteSSLLogger(websiteSSL *model.WebsiteSSL, autoRenew bool) (*os.File, *log.Logger) {
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if autoRenew {
		flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}

	logFile, err := os.OpenFile(websiteSSL.GetLogPath(), flags, constant.FilePerm)
	if err != nil {
		global.LOG.Errorf("open ssl log file failed, domain: %s, err: %v", websiteSSL.PrimaryDomain, err)
		return nil, log.New(io.Discard, "", log.LstdFlags)
	}

	if autoRenew {
		if info, statErr := logFile.Stat(); statErr == nil && info.Size() > 0 {
			_, _ = logFile.WriteString("\n")
		}
		_, _ = logFile.WriteString(fmt.Sprintf("========== [%s] auto renew attempt ==========\n", time.Now().Format(constant.DateTimeLayout)))
	}

	return logFile, log.New(logFile, "", log.LstdFlags)
}

func reloadSystemSSL(websiteSSL *model.WebsiteSSL, logger *log.Logger) {
	if !global.IsMaster {
		return
	}
	systemSSLEnable, sslID := GetSystemSSL()
	if systemSSLEnable && sslID == websiteSSL.ID {
		fileOp := files.NewFileOp()
		certPath := path.Join(global.Dir.DataDir, "secret/server.crt")
		keyPath := path.Join(global.Dir.DataDir, "secret/server.key")
		printSSLLog(logger, "StartUpdateSystemSSL", nil)
		if err := fileOp.WriteFile(certPath, strings.NewReader(websiteSSL.Pem), 0600); err != nil {
			logger.Printf("Failed to update the SSL certificate File for 1Panel System domain [%s] , err:%s", websiteSSL.PrimaryDomain, err.Error())
			return
		}
		if err := fileOp.WriteFile(keyPath, strings.NewReader(websiteSSL.PrivateKey), 0600); err != nil {
			logger.Printf("Failed to update the SSL certificate for 1Panel System domain [%s] , err:%s", websiteSSL.PrimaryDomain, err.Error())
			return
		}
		if err := req_helper.PostLocalCore("/core/settings/ssl/reload"); err != nil {
			logger.Printf("Failed to update the SSL certificate for 1Panel System domain [%s] , err:%s", websiteSSL.PrimaryDomain, err.Error())
			return
		}
		printSSLLog(logger, "UpdateSystemSSLSuccess", nil)
	}
}

// SyncSystemSSL reconciles the panel certificate on disk with the WebsiteSSL
// row referenced by the SSLID setting. When they differ (typically because a
// previous renewal's reloadSystemSSL was skipped by a transient nginx reload
// failure or because the panel was restarted between the DB save and the file
// rewrite), the on-disk cert is refreshed and core is asked to reload its TLS
// store. Safe to call on every renewal cron tick: it is a no-op when SSL is
// not enabled, when SSLID does not resolve, or when the cert already matches.
func SyncSystemSSL() {
	if !global.IsMaster {
		return
	}
	systemSSLEnable, sslID := GetSystemSSL()
	if !systemSSLEnable {
		return
	}
	websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(sslID))
	if err != nil {
		return
	}
	if strings.TrimSpace(websiteSSL.Pem) == "" || strings.TrimSpace(websiteSSL.PrivateKey) == "" {
		return
	}
	certPath := path.Join(global.Dir.DataDir, "secret/server.crt")
	keyPath := path.Join(global.Dir.DataDir, "secret/server.key")
	diskCert, _ := os.ReadFile(certPath)
	diskKey, _ := os.ReadFile(keyPath)
	if strings.TrimSpace(string(diskCert)) == strings.TrimSpace(websiteSSL.Pem) &&
		strings.TrimSpace(string(diskKey)) == strings.TrimSpace(websiteSSL.PrivateKey) {
		return
	}
	global.LOG.Infof("panel SSL on disk diverged from DB (SSLID=%d, domain=%s), syncing", websiteSSL.ID, websiteSSL.PrimaryDomain)
	fileOp := files.NewFileOp()
	if err := fileOp.WriteFile(certPath, strings.NewReader(websiteSSL.Pem), 0600); err != nil {
		global.LOG.Errorf("sync panel SSL: write cert failed: %s", err.Error())
		return
	}
	if err := fileOp.WriteFile(keyPath, strings.NewReader(websiteSSL.PrivateKey), 0600); err != nil {
		global.LOG.Errorf("sync panel SSL: write key failed: %s", err.Error())
		return
	}
	if err := req_helper.PostLocalCore("/core/settings/ssl/reload"); err != nil {
		global.LOG.Errorf("sync panel SSL: notify core failed: %s", err.Error())
		return
	}
	global.LOG.Info("panel SSL synced from DB to disk")
}

func (w WebsiteSSLService) ObtainSSL(apply request.WebsiteSSLApply) error {
	return w.obtainSSL(apply.ID, false)
}

func (w WebsiteSSLService) AutoRenewSSL(id uint) error {
	return w.obtainSSL(id, true)
}

func (w WebsiteSSLService) obtainSSL(id uint, autoRenew bool) error {
	var (
		err         error
		websiteSSL  *model.WebsiteSSL
		acmeAccount *model.WebsiteAcmeAccount
		dnsAccount  *model.WebsiteDnsAccount
		logFile     *os.File
		logger      *log.Logger
		resource    certificate.Resource
		httpRoot    string
	)

	websiteSSL, err = websiteSSLRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return err
	}
	if websiteSSL.Status == constant.SSLApply {
		return buserr.New("InExecuting")
	}
	acmeAccount, err = websiteAcmeRepo.GetFirst(repo.WithByID(websiteSSL.AcmeAccountID))
	if err != nil {
		return err
	}
	domains := []string{websiteSSL.PrimaryDomain}
	if websiteSSL.Domains != "" {
		domains = append(domains, strings.Split(websiteSSL.Domains, ",")...)
	}
	if websiteSSL.Provider != constant.DnsManual {
		switch websiteSSL.Provider {
		case constant.DNSAccount:
			dnsAccount, err = websiteDnsRepo.GetFirst(repo.WithByID(websiteSSL.DnsAccountID))
			if err != nil {
				return err
			}
		case constant.Http:
			appInstall, err := getAppInstallByKey(constant.AppOpenresty)
			if err != nil {
				if errors.Is(err, gormv2.ErrRecordNotFound) {
					return buserr.New("ErrOpenrestyNotFound")
				}
				return err
			}
			for _, domain := range domains {
				if strings.Contains(domain, "*") {
					return buserr.New("ErrWildcardDomain")
				}
			}
			httpRoot = path.Join(appInstall.GetPath(), "root")
		}
	}
	marked, err := websiteSSLRepo.TryMarkApplying(websiteSSL.ID)
	if err != nil {
		return err
	}
	if !marked {
		return buserr.New("InExecuting")
	}
	websiteSSL.Status = constant.SSLApply

	logFile, logger = newWebsiteSSLLogger(websiteSSL, autoRenew)
	logFileOwnedByGoroutine := false
	if logFile != nil {
		defer func() {
			if !logFileOwnedByGoroutine {
				_ = logFile.Close()
			}
		}()
	}
	startMsg := i18n.GetMsgWithMap("ApplySSLStart", map[string]interface{}{"domain": strings.Join(domains, ","), "type": i18n.GetMsgByKey(websiteSSL.Provider)})
	if websiteSSL.Provider == constant.DNSAccount {
		startMsg = startMsg + i18n.GetMsgWithMap("DNSAccountName", map[string]interface{}{"name": dnsAccount.Name, "type": dnsAccount.Type})
	}
	logger.Println(startMsg)

	logFileOwnedByGoroutine = true
	go func(logFile *os.File, logger *log.Logger) {
		if logFile != nil {
			defer func() {
				_ = logFile.Close()
			}()
		}
		if websiteSSL.Provider != constant.DnsManual {
			privateKey, err := ssl.GetPrivateKeyByType(websiteSSL.KeyType, websiteSSL.PrivateKey)
			if err != nil {
				handleError(websiteSSL, logger, err)
				return
			}
			err = withLegoLoggerTimeout(logger, sslObtainTimeout, func(ctx context.Context) error {
				client, err := newWebsiteSSLLegoClient(ctx, acmeAccount)
				if err != nil {
					return err
				}
				switch websiteSSL.Provider {
				case constant.DNSAccount:
					if err = client.UseDns(ssl.DnsType(dnsAccount.Type), dnsAccount.Authorization, *websiteSSL); err != nil {
						return err
					}
				case constant.Http:
					if err = client.UseHTTP(httpRoot); err != nil {
						return err
					}
				}
				if websiteSSL.IsIp {
					resource, err = client.ObtainIPSSL(ctx, domains[0], privateKey)
				} else {
					resource, err = client.ObtainSSL(ctx, domains, privateKey)
				}
				return err
			})
			if err != nil {
				handleError(websiteSSL, logger, err)
				return
			}
		} else {
			manualClient, err := ssl.NewCustomAcmeClient(acmeAccount, getSystemProxy(acmeAccount.UseProxy), logger)
			if err != nil {
				handleError(websiteSSL, logger, err)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), sslObtainTimeout)
			resource, err = manualClient.RequestCertificate(ctx, websiteSSL)
			cancel()
			if err != nil {
				handleError(websiteSSL, logger, err)
				return
			}
		}

		websiteSSL.PrivateKey = string(resource.PrivateKey)
		websiteSSL.Pem = string(resource.Certificate)
		websiteSSL.CertURL = resource.CertURL
		cert, err := parseCertificatePEM(resource.Certificate)
		if err != nil {
			handleError(websiteSSL, logger, err)
			return
		}
		websiteSSL.ExpireDate = cert.NotAfter
		websiteSSL.StartDate = cert.NotBefore
		websiteSSL.Type = cert.Issuer.CommonName
		if len(cert.Issuer.Organization) > 0 {
			websiteSSL.Organization = cert.Issuer.Organization[0]
		}
		websiteSSL.Status = constant.SSLReady
		printSSLLog(logger, "ApplySSLSuccess", map[string]interface{}{"domain": strings.Join(domains, ",")})
		saveCertificateFile(websiteSSL, logger)

		if websiteSSL.ExecShell {
			workDir := global.Dir.DataDir
			if websiteSSL.PushDir {
				workDir = websiteSSL.Dir
			}
			printSSLLog(logger, "ExecShellStart", nil)
			if err = runShellScriptFile(workDir, websiteSSL.Shell, logger); err != nil {
				printSSLLog(logger, "ErrExecShell", map[string]interface{}{"err": err.Error()})
			} else {
				printSSLLog(logger, "ExecShellSuccess", nil)
			}
		}

		err = websiteSSLRepo.Save(websiteSSL)
		if err != nil {
			return
		}

		websites, _ := websiteRepo.GetBy(websiteRepo.WithWebsiteSSLID(websiteSSL.ID))
		if len(websites) > 0 {
			for _, website := range websites {
				printSSLLog(logger, "ApplyWebSiteSSLLog", map[string]interface{}{"name": website.PrimaryDomain})
				if err := createPemFile(website, *websiteSSL); err != nil {
					printSSLLog(logger, "ErrUpdateWebsiteSSL", map[string]interface{}{"name": website.PrimaryDomain, "err": err.Error()})
				}
			}
			if nginxInstall, err := getAppInstallByKey(constant.AppOpenresty); err == nil {
				if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
					printSSLLog(logger, "ErrSSLApply", nil)
				} else {
					printSSLLog(logger, "ApplyWebSiteSSLSuccess", nil)
				}
			}
		}
		reloadSystemSSL(websiteSSL, logger)
		if websiteSSL.PushNode {
			if err = pushSSLToNode(websiteSSL, logger); err != nil {
				return
			}
		}
	}(logFile, logger)

	return nil
}

func runShellScriptFile(workDir, shell string, logger *log.Logger) error {
	file, err := os.CreateTemp("", "1panel-shell-*.sh")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()
	if _, err := file.WriteString(shell); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(30*time.Minute), cmd.WithLogger(logger), cmd.WithWorkDir(workDir))
	return cmdMgr.Run("bash", file.Name())
}

func parseCertificatePEM(certPEM []byte) (*x509.Certificate, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, errors.New("invalid certificate PEM")
	}
	return x509.ParseCertificate(certBlock.Bytes)
}

func handleError(websiteSSL *model.WebsiteSSL, logger *log.Logger, err error) {
	if websiteSSL.Status == constant.SSLInit || websiteSSL.Status == constant.SSLError {
		websiteSSL.Status = constant.StatusError
	} else {
		websiteSSL.Status = constant.SSLApplyError
	}
	websiteSSL.Message = err.Error()
	if logger != nil {
		logger.Println(i18n.GetErrMsg("ApplySSLFailed", map[string]interface{}{"domain": websiteSSL.PrimaryDomain, "detail": err.Error()}))
	}
	_ = websiteSSLRepo.Save(websiteSSL)
}

func (w WebsiteSSLService) GetDNSResolve(req request.WebsiteDNSReq) ([]response.WebsiteDNSRes, error) {
	acmeAccount, err := websiteAcmeRepo.GetFirst(repo.WithByID(req.AcmeAccountID))
	if err != nil {
		return nil, err
	}
	client, err := ssl.NewCustomAcmeClient(acmeAccount, getSystemProxy(acmeAccount.UseProxy), nil)
	if err != nil {
		return nil, err
	}
	websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(req.WebsiteSSLID))
	if err != nil {
		return nil, err
	}
	resolves, err := client.GetDNSResolve(context.TODO(), websiteSSL)
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
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteId))
	if err != nil {
		return res, err
	}
	websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(website.WebsiteSSLID))
	if err != nil {
		return res, err
	}
	res.WebsiteSSL = *websiteSSL
	return res, nil
}

func (w WebsiteSSLService) Delete(ids []uint) error {
	var (
		websiteSSLS []string
		applySSLS   []string
	)
	for _, id := range ids {
		if websites, _ := websiteRepo.GetBy(websiteRepo.WithWebsiteSSLID(id)); len(websites) > 0 {
			oldSSL, _ := websiteSSLRepo.GetFirst(repo.WithByID(id))
			if oldSSL != nil && oldSSL.ID > 0 {
				websiteSSLS = append(websiteSSLS, oldSSL.PrimaryDomain)
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
		websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(id))
		if err != nil {
			return err
		}
		if websiteSSL.Status == constant.SSLApply {
			applySSLS = append(applySSLS, websiteSSL.PrimaryDomain)
			continue
		}
		if websiteSSL.Provider != constant.Manual && websiteSSL.Provider != constant.SelfSigned {
			go func() {
				acmeAccount, err := websiteAcmeRepo.GetFirst(repo.WithByID(websiteSSL.AcmeAccountID))
				if err != nil {
					global.LOG.Errorf("Failed to get acme account for SSL revoke, err: %v", err)
					return
				}
				err = withLegoLogger(nil, func() error {
					client, err := newWebsiteSSLLegoClient(context.Background(), acmeAccount)
					if err != nil {
						return err
					}
					return client.RevokeSSL([]byte(websiteSSL.Pem))
				})
				if err != nil {
					global.LOG.Errorf("Failed to revoke SSL for domain %s, err: %v", websiteSSL.PrimaryDomain, err)
					return
				}
			}()
		}
		_ = websiteSSLRepo.DeleteBy(repo.WithByID(id))
	}
	if len(websiteSSLS) > 0 {
		return buserr.WithName("ErrSSLCannotDelete", strings.Join(websiteSSLS, ","))
	}
	if len(applySSLS) > 0 {
		return buserr.WithName("ErrApplySSLCanNotDelete", strings.Join(applySSLS, ","))
	}
	return nil
}

func (w WebsiteSSLService) Update(update request.WebsiteSSLUpdate) error {
	websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(update.ID))
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
	pushNode, nodes := normalizeSSLPushConfig(update.PushNode, update.Nodes)
	if !global.IsMaster || !xpack.MultiNodeProvider.IsXpack() {
		pushNode = false
		nodes = ""
	}
	updateParams["push_node"] = pushNode
	updateParams["nodes"] = nodes

	if websiteSSL.Provider != constant.SelfSigned && websiteSSL.Provider != constant.Manual {
		acmeAccount, err := websiteAcmeRepo.GetFirst(repo.WithByID(update.AcmeAccountID))
		if err != nil {
			return err
		}
		updateParams["acme_account_id"] = acmeAccount.ID
	}

	if update.PushDir {
		fileOP := files.NewFileOp()
		if !fileOP.Stat(update.Dir) {
			_ = fileOP.CreateDir(update.Dir, constant.DirPerm)
		}
		updateParams["dir"] = update.Dir
	}
	var domains []string
	if update.OtherDomains != "" {
		otherDomainArray := strings.Split(update.OtherDomains, "\n")
		for _, domain := range otherDomainArray {
			if websiteSSL.Provider != constant.SelfSigned && !common.IsValidDomain(domain) {
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
		dnsAccount, err := websiteDnsRepo.GetFirst(repo.WithByID(update.DnsAccountID))
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
	setSSLPushConfig(websiteSSL, req.PushNode, req.Nodes)
	var err error
	if req.SSLID > 0 {
		websiteSSL, err = websiteSSLRepo.GetFirst(repo.WithByID(req.SSLID))
		if err != nil {
			return err
		}
		websiteSSL.Description = req.Description
		setSSLPushConfig(websiteSSL, req.PushNode, req.Nodes)
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
		websiteSSL.CertPath = req.CertificatePath
		websiteSSL.PrivateKeyPath = req.PrivateKeyPath
	} else {
		websiteSSL.PrivateKey = req.PrivateKey
		websiteSSL.Pem = req.Certificate
		websiteSSL.CertPath = ""
		websiteSSL.PrivateKeyPath = ""
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
	if pemData == nil || cert == nil {
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
		if err := websiteSSLRepo.Save(websiteSSL); err != nil {
			return err
		}
		return pushSSLToNodeWithNewLogger(websiteSSL)
	}
	if err := websiteSSLRepo.Create(context.Background(), websiteSSL); err != nil {
		return err
	}
	return pushSSLToNodeWithNewLogger(websiteSSL)
}

func (w WebsiteSSLService) PushToNode(req request.WebsiteSSLPush) error {
	if !global.IsMaster {
		return errors.New("only master node can push SSL to nodes")
	}
	if !xpack.MultiNodeProvider.IsXpack() {
		return errors.New("SSL node push is an XPack feature")
	}
	pushNode, nodes := normalizeSSLPushConfig(req.PushNode, req.Nodes)
	if !pushNode {
		return errors.New("please select nodes to push SSL")
	}
	websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if websiteSSL.Provider == constant.FromMaster {
		return errors.New("SSL imported from master node can not be pushed")
	}
	if websiteSSL.Status != constant.SSLReady {
		return errors.New("only ready SSL can be pushed")
	}
	if task.CheckResourceTaskIsExecuting(task.TaskPush, task.TaskScopeWebsite, websiteSSL.ID) {
		return buserr.New("TaskIsExecuting")
	}
	if err := websiteSSLRepo.SaveByMap(websiteSSL, map[string]interface{}{
		"push_node": pushNode,
		"nodes":     nodes,
	}); err != nil {
		return err
	}
	websiteSSL.PushNode = pushNode
	websiteSSL.Nodes = nodes

	if req.Sync {
		return xpack.MultiNodeProvider.PushSSLToNode(websiteSSL)
	}

	pushTask, err := task.NewTaskWithOps(websiteSSL.PrimaryDomain, task.TaskPush, task.TaskScopeWebsite, req.TaskID, websiteSSL.ID)
	if err != nil {
		return err
	}
	pushTask.AddSubTask(i18n.GetMsgByKey("StartPushSSLToNode"), func(t *task.Task) error {
		t.Log(i18n.GetMsgByKey("StartPushSSLToNode"))
		if err := xpack.MultiNodeProvider.PushSSLToNode(websiteSSL); err != nil {
			t.Log(i18n.GetMsgWithMap("PushSSLToNodeFailed", map[string]interface{}{"err": err.Error()}))
			return err
		}
		t.Log(i18n.GetMsgByKey("PushSSLToNodeSuccess"))
		return nil
	}, nil)
	go func() {
		if err := pushTask.Execute(); err != nil {
			global.LOG.Errorf("push ssl to node failed, sslID: %d, err: %v", websiteSSL.ID, err)
		}
	}()
	return nil
}

func (w WebsiteSSLService) DownloadFile(id uint) (*os.File, error) {
	websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return nil, err
	}
	fileOp := files.NewFileOp()
	dir := path.Join(global.Dir.DataDir, "tmp/ssl", websiteSSL.PrimaryDomain)
	if fileOp.Stat(dir) {
		if err = fileOp.DeleteDir(dir); err != nil {
			return nil, err
		}
	}
	if err = fileOp.CreateDir(dir, constant.DirPerm); err != nil {
		return nil, err
	}
	if err = fileOp.WriteFile(path.Join(dir, "fullchain.pem"), strings.NewReader(websiteSSL.Pem), constant.DirPerm); err != nil {
		return nil, err
	}
	if err = fileOp.WriteFile(path.Join(dir, "privkey.pem"), strings.NewReader(websiteSSL.PrivateKey), constant.DirPerm); err != nil {
		return nil, err
	}
	fileName := websiteSSL.PrimaryDomain + ".zip"
	if err = fileOp.Compress(context.Background(), []string{path.Join(dir, "fullchain.pem"), path.Join(dir, "privkey.pem")}, dir, fileName, files.SdkZip, "", nil); err != nil {
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

func (w WebsiteSSLService) ImportMasterSSL(create model.WebsiteSSL) error {
	websiteSSL, err := websiteSSLRepo.GetFirst(websiteSSLRepo.WithByMasterSSLID(create.ID))
	if err != nil {
		if !errors.Is(err, gormv2.ErrRecordNotFound) {
			return err
		}
		websiteSSL = &model.WebsiteSSL{}
	}
	websiteSSL.Status = constant.SSLReady
	websiteSSL.Provider = constant.FromMaster
	websiteSSL.PrimaryDomain = create.PrimaryDomain
	websiteSSL.StartDate = create.StartDate
	websiteSSL.ExpireDate = create.ExpireDate
	websiteSSL.KeyType = create.KeyType
	websiteSSL.Description = create.Description
	websiteSSL.PrivateKey = create.PrivateKey
	websiteSSL.Pem = create.Pem
	websiteSSL.Type = create.Type
	websiteSSL.Organization = create.Organization
	websiteSSL.MasterSSLID = create.ID
	websiteSSL.Domains = create.Domains
	if websiteSSL.ID == 0 {
		if err := websiteSSLRepo.Create(context.Background(), websiteSSL); err != nil {
			return err
		}
	} else {
		if err := websiteSSLRepo.Save(websiteSSL); err != nil {
			return err
		}
	}
	websites, _ := websiteRepo.GetBy(websiteRepo.WithWebsiteSSLID(websiteSSL.ID))
	if len(websites) == 0 {
		return nil
	}
	for _, website := range websites {
		if err := createPemFile(website, *websiteSSL); err != nil {
			continue
		}
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err == nil {
		if err := opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
			return err
		}
	}
	return nil
}
