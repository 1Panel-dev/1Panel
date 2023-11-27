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
	GetCA(id uint) (response.WebsiteCADTO, error)
	Delete(id uint) error
	ObtainSSL(req request.WebsiteCAObtain) error
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
		CommonName:   create.CommonName,
		Country:      []string{create.Country},
		Organization: []string{create.Organization},
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

func (w WebsiteCAService) GetCA(id uint) (response.WebsiteCADTO, error) {
	ca, err := websiteCARepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return response.WebsiteCADTO{}, err
	}
	return response.WebsiteCADTO{
		WebsiteCA: ca,
	}, nil
}

func (w WebsiteCAService) Delete(id uint) error {
	return websiteCARepo.DeleteBy(commonRepo.WithByID(id))
}

func (w WebsiteCAService) ObtainSSL(req request.WebsiteCAObtain) error {
	ca, err := websiteCARepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	newSSL := &model.WebsiteSSL{
		Provider: constant.SelfSigned,
		KeyType:  req.KeyType,
		PushDir:  req.PushDir,
	}
	if req.PushDir {
		if !files.NewFileOp().Stat(req.Dir) {
			return buserr.New(constant.ErrLinkPathNotFound)
		}
		newSSL.Dir = req.Dir
	}

	var (
		domains []string
		ips     []net.IP
	)
	if req.Domains != "" {
		domainArray := strings.Split(req.Domains, "\n")
		for _, domain := range domainArray {
			if !common.IsValidDomain(domain) {
				err = buserr.WithName("ErrDomainFormat", domain)
				return err
			} else {
				if ipAddress := net.ParseIP(domain); ipAddress == nil {
					domains = append(domains, domain)
				} else {
					ips = append(ips, ipAddress)
				}
			}
		}
		if len(domains) > 0 {
			newSSL.PrimaryDomain = domains[0]
			newSSL.Domains = strings.Join(domains[1:], ",")
		}
	}

	rootCertBlock, _ := pem.Decode([]byte(ca.CSR))
	if rootCertBlock == nil {
		return buserr.New("ErrSSLCertificateFormat")
	}
	rootCsr, err := x509.ParseCertificate(rootCertBlock.Bytes)
	if err != nil {
		return err
	}
	rootPrivateKeyBlock, _ := pem.Decode([]byte(ca.PrivateKey))
	if rootPrivateKeyBlock == nil {
		return buserr.New("ErrSSLCertificateFormat")
	}

	var rootPrivateKey any
	if ssl.KeyType(ca.KeyType) == certcrypto.EC256 || ssl.KeyType(ca.KeyType) == certcrypto.EC384 {
		rootPrivateKey, err = x509.ParseECPrivateKey(rootPrivateKeyBlock.Bytes)
		if err != nil {
			return err
		}
	} else {
		rootPrivateKey, err = x509.ParsePKCS1PrivateKey(rootPrivateKeyBlock.Bytes)
		if err != nil {
			return err
		}
	}
	interPrivateKey, interPublicKey, _, err := createPrivateKey(req.KeyType)
	if err != nil {
		return err
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
		return err
	}
	interCert, err := x509.ParseCertificate(interDer)
	if err != nil {
		return err
	}

	_, publicKey, privateKeyBytes, err := createPrivateKey(req.KeyType)
	if err != nil {
		return err
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
		return err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return err
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	pemData := pem.EncodeToMemory(certBlock)
	newSSL.Pem = string(pemData)
	newSSL.PrivateKey = string(privateKeyBytes)
	newSSL.ExpireDate = cert.NotAfter
	newSSL.StartDate = cert.NotBefore
	newSSL.Type = cert.Issuer.CommonName
	newSSL.Organization = rootCsr.Subject.Organization[0]

	if err := websiteSSLRepo.Create(context.Background(), newSSL); err != nil {
		return err
	}
	logFile, _ := os.OpenFile(path.Join(constant.SSLLogDir, fmt.Sprintf("%s-ssl-%d.log", newSSL.PrimaryDomain, newSSL.ID)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	logger.Println(i18n.GetMsgWithMap("ApplySSLSuccess", map[string]interface{}{"domain": strings.Join(domains, ",")}))
	saveCertificateFile(*newSSL, logger)
	return nil
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
		publicKey = privateKey.(*rsa.PrivateKey).PublicKey
		publicKey = publicKey.(*rsa.PublicKey)
		_ = pem.Encode(caPrivateKeyPEM, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey.(*rsa.PrivateKey)),
		})
	}
	privateKeyBytes = caPrivateKeyPEM.Bytes()
	return
}
