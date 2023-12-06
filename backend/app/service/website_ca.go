package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
	"github.com/go-acme/lego/v4/certcrypto"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

type WebsiteCAService struct {
}

type IWebsiteCAService interface {
	Page(search request.WebsiteCASearch) (int64, []response.WebsiteCADTO, error)
	Create(create request.WebsiteCACreate) (*request.WebsiteCACreate, error)
	GetCA(id uint) (*response.WebsiteCADTO, error)
	Delete(id uint) error
	ObtainSSL(req request.WebsiteCAObtain) (*model.WebsiteSSL, error)
}

func NewIWebsiteCAService() IWebsiteCAService {
	return &WebsiteCAService{}
}

func (w WebsiteCAService) Page(search request.WebsiteCASearch) (int64, []response.WebsiteCADTO, error) {
	total, cas, err := websiteCARepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	if err != nil {
		return 0, nil, err
	}
	var caDTOs []response.WebsiteCADTO
	for _, ca := range cas {
		caDTOs = append(caDTOs, response.WebsiteCADTO{
			WebsiteCA: ca,
		})
	}
	return total, caDTOs, err
}

func (w WebsiteCAService) Create(create request.WebsiteCACreate) (*request.WebsiteCACreate, error) {
	if exist, _ := websiteCARepo.GetFirst(commonRepo.WithByName(create.Name)); exist.ID > 0 {
		return nil, buserr.New(constant.ErrNameIsExist)
	}

	ca := &model.WebsiteCA{
		Name:    create.Name,
		KeyType: create.KeyType,
	}

	pkixName := pkix.Name{
		CommonName:         create.CommonName,
		Country:            []string{create.Country},
		Organization:       []string{create.Organization},
		OrganizationalUnit: []string{create.OrganizationUint},
	}
	if create.Province != "" {
		pkixName.Province = []string{create.Province}
	}
	if create.City != "" {
		pkixName.Locality = []string{create.City}
	}

	rootCA := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().Unix()),
		Subject:               pkixName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
		MaxPathLenZero:        false,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	privateKey, err := certcrypto.GeneratePrivateKey(ssl.KeyType(create.KeyType))
	if err != nil {
		return nil, err
	}
	var (
		publicKey       any
		caPEM           = new(bytes.Buffer)
		caPrivateKeyPEM = new(bytes.Buffer)
		privateBlock    = &pem.Block{}
	)
	if ssl.KeyType(create.KeyType) == certcrypto.EC256 || ssl.KeyType(create.KeyType) == certcrypto.EC384 {
		publicKey = &privateKey.(*ecdsa.PrivateKey).PublicKey
		publicKey = publicKey.(*ecdsa.PublicKey)
		privateBlock.Type = "EC PRIVATE KEY"
		privateBytes, err := x509.MarshalECPrivateKey(privateKey.(*ecdsa.PrivateKey))
		if err != nil {
			return nil, err
		}
		privateBlock.Bytes = privateBytes
		_ = pem.Encode(caPrivateKeyPEM, privateBlock)
	} else {
		publicKey = privateKey.(*rsa.PrivateKey).PublicKey
		publicKey = publicKey.(*rsa.PublicKey)
		privateBlock.Type = "RSA PRIVATE KEY"
		privateBlock.Bytes = x509.MarshalPKCS1PrivateKey(privateKey.(*rsa.PrivateKey))
	}
	ca.PrivateKey = string(pem.EncodeToMemory(privateBlock))

	caBytes, err := x509.CreateCertificate(rand.Reader, rootCA, rootCA, publicKey, privateKey)
	if err != nil {
		return nil, err
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}
	_ = pem.Encode(caPEM, certBlock)
	pemData := pem.EncodeToMemory(certBlock)
	ca.CSR = string(pemData)

	if err := websiteCARepo.Create(context.Background(), ca); err != nil {
		return nil, err
	}
	return &create, nil
}

func (w WebsiteCAService) GetCA(id uint) (*response.WebsiteCADTO, error) {
	res := &response.WebsiteCADTO{}
	ca, err := websiteCARepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	res.WebsiteCA = ca
	certBlock, _ := pem.Decode([]byte(ca.CSR))
	if certBlock == nil {
		return nil, buserr.New("ErrSSLCertificateFormat")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}
	res.CommonName = cert.Issuer.CommonName
	res.Organization = strings.Join(cert.Issuer.Organization, ",")
	res.Country = strings.Join(cert.Issuer.Country, ",")
	res.Province = strings.Join(cert.Issuer.Province, ",")
	res.City = strings.Join(cert.Issuer.Locality, ",")
	res.OrganizationUint = strings.Join(cert.Issuer.OrganizationalUnit, ",")

	return res, nil
}

func (w WebsiteCAService) Delete(id uint) error {
	ssls, _ := websiteSSLRepo.List(websiteSSLRepo.WithByCAID(id))
	if len(ssls) > 0 {
		return buserr.New("ErrDeleteCAWithSSL")
	}
	exist, err := websiteCARepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return err
	}
	if exist.Name == "1Panel" {
		return buserr.New("ErrDefaultCA")
	}
	return websiteCARepo.DeleteBy(commonRepo.WithByID(id))
}

func (w WebsiteCAService) ObtainSSL(req request.WebsiteCAObtain) (*model.WebsiteSSL, error) {
	var (
		domains    []string
		ips        []net.IP
		websiteSSL = &model.WebsiteSSL{}
		err        error
		ca         model.WebsiteCA
	)
	if req.Renew {
		websiteSSL, err = websiteSSLRepo.GetFirst(commonRepo.WithByID(req.SSLID))
		if err != nil {
			return nil, err
		}
		ca, err = websiteCARepo.GetFirst(commonRepo.WithByID(websiteSSL.CaID))
		if err != nil {
			return nil, err
		}
		existDomains := []string{websiteSSL.PrimaryDomain}
		if websiteSSL.Domains != "" {
			existDomains = append(existDomains, strings.Split(websiteSSL.Domains, ",")...)
		}
		for _, domain := range existDomains {
			if ipAddress := net.ParseIP(domain); ipAddress == nil {
				domains = append(domains, domain)
			} else {
				ips = append(ips, ipAddress)
			}
		}
	} else {
		ca, err = websiteCARepo.GetFirst(commonRepo.WithByID(req.ID))
		if err != nil {
			return nil, err
		}
		websiteSSL = &model.WebsiteSSL{
			Provider:    constant.SelfSigned,
			KeyType:     req.KeyType,
			PushDir:     req.PushDir,
			CaID:        ca.ID,
			AutoRenew:   req.AutoRenew,
			Description: req.Description,
		}
		if req.PushDir {
			if !files.NewFileOp().Stat(req.Dir) {
				return nil, buserr.New(constant.ErrLinkPathNotFound)
			}
			websiteSSL.Dir = req.Dir
		}
		if req.Domains != "" {
			domainArray := strings.Split(req.Domains, "\n")
			for _, domain := range domainArray {
				if !common.IsValidDomain(domain) {
					err = buserr.WithName("ErrDomainFormat", domain)
					return nil, err
				} else {
					if ipAddress := net.ParseIP(domain); ipAddress == nil {
						domains = append(domains, domain)
					} else {
						ips = append(ips, ipAddress)
					}
				}
			}
			if len(domains) > 0 {
				websiteSSL.PrimaryDomain = domains[0]
				websiteSSL.Domains = strings.Join(domains[1:], ",")
			}
			ipStrings := make([]string, len(ips))
			for i, ip := range ips {
				ipStrings[i] = ip.String()
			}
			if websiteSSL.PrimaryDomain == "" && len(ips) > 0 {
				websiteSSL.PrimaryDomain = ipStrings[0]
				ipStrings = ipStrings[1:]
			}
			if len(ipStrings) > 0 {
				if websiteSSL.Domains != "" {
					websiteSSL.Domains += ","
				}
				websiteSSL.Domains += strings.Join(ipStrings, ",")
			}

		}
	}

	rootCertBlock, _ := pem.Decode([]byte(ca.CSR))
	if rootCertBlock == nil {
		return nil, buserr.New("ErrSSLCertificateFormat")
	}
	rootCsr, err := x509.ParseCertificate(rootCertBlock.Bytes)
	if err != nil {
		return nil, err
	}
	rootPrivateKeyBlock, _ := pem.Decode([]byte(ca.PrivateKey))
	if rootPrivateKeyBlock == nil {
		return nil, buserr.New("ErrSSLCertificateFormat")
	}

	var rootPrivateKey any
	if ssl.KeyType(ca.KeyType) == certcrypto.EC256 || ssl.KeyType(ca.KeyType) == certcrypto.EC384 {
		rootPrivateKey, err = x509.ParseECPrivateKey(rootPrivateKeyBlock.Bytes)
		if err != nil {
			return nil, err
		}
	} else {
		rootPrivateKey, err = x509.ParsePKCS1PrivateKey(rootPrivateKeyBlock.Bytes)
		if err != nil {
			return nil, err
		}
	}
	interPrivateKey, interPublicKey, _, err := createPrivateKey(websiteSSL.KeyType)
	if err != nil {
		return nil, err
	}
	notAfter := time.Now()
	if req.Unit == "year" {
		notAfter = notAfter.AddDate(req.Time, 0, 0)
	} else {
		notAfter = notAfter.AddDate(0, 0, req.Time)
	}
	interCsr := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().Unix()),
		Subject:               rootCsr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	interDer, err := x509.CreateCertificate(rand.Reader, interCsr, rootCsr, interPublicKey, rootPrivateKey)
	if err != nil {
		return nil, err
	}
	interCert, err := x509.ParseCertificate(interDer)
	if err != nil {
		return nil, err
	}

	_, publicKey, privateKeyBytes, err := createPrivateKey(websiteSSL.KeyType)
	if err != nil {
		return nil, err
	}

	csr := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().Unix()),
		Subject:               rootCsr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		IsCA:                  false,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              domains,
		IPAddresses:           ips,
	}

	der, err := x509.CreateCertificate(rand.Reader, csr, interCert, publicKey, interPrivateKey)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	pemData := pem.EncodeToMemory(certBlock)
	websiteSSL.Pem = string(pemData)
	websiteSSL.PrivateKey = string(privateKeyBytes)
	websiteSSL.ExpireDate = cert.NotAfter
	websiteSSL.StartDate = cert.NotBefore
	websiteSSL.Type = cert.Issuer.CommonName
	websiteSSL.Organization = rootCsr.Subject.Organization[0]

	if req.Renew {
		if err := websiteSSLRepo.Save(websiteSSL); err != nil {
			return nil, err
		}
	} else {
		if err := websiteSSLRepo.Create(context.Background(), websiteSSL); err != nil {
			return nil, err
		}
	}

	logFile, _ := os.OpenFile(path.Join(constant.SSLLogDir, fmt.Sprintf("%s-ssl-%d.log", websiteSSL.PrimaryDomain, websiteSSL.ID)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	logger.Println(i18n.GetMsgWithMap("ApplySSLSuccess", map[string]interface{}{"domain": strings.Join(domains, ",")}))
	saveCertificateFile(websiteSSL, logger)
	return websiteSSL, nil
}

func createPrivateKey(keyType string) (privateKey any, publicKey any, privateKeyBytes []byte, err error) {
	privateKey, err = certcrypto.GeneratePrivateKey(ssl.KeyType(keyType))
	if err != nil {
		return
	}
	var (
		caPrivateKeyPEM = new(bytes.Buffer)
	)
	if ssl.KeyType(keyType) == certcrypto.EC256 || ssl.KeyType(keyType) == certcrypto.EC384 {
		publicKey = &privateKey.(*ecdsa.PrivateKey).PublicKey
		publicKey = publicKey.(*ecdsa.PublicKey)
		block := &pem.Block{
			Type: "EC PRIVATE KEY",
		}
		privateBytes, sErr := x509.MarshalECPrivateKey(privateKey.(*ecdsa.PrivateKey))
		if sErr != nil {
			err = sErr
			return
		}
		block.Bytes = privateBytes
		_ = pem.Encode(caPrivateKeyPEM, block)
	} else {
		publicKey = &privateKey.(*rsa.PrivateKey).PublicKey
		_ = pem.Encode(caPrivateKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey.(*rsa.PrivateKey)),
		})
	}
	privateKeyBytes = caPrivateKeyPEM.Bytes()
	return
}
