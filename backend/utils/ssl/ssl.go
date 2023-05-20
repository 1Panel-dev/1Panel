package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
)

func GenerateSSL(domain string) error {
	rootPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	ipItem := net.ParseIP(domain)
	isIP := false
	if len(ipItem) != 0 {
		isIP = true
	}

	rootTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "1Panel Root CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	if isIP {
		rootTemplate.IPAddresses = []net.IP{ipItem}
	} else {
		rootTemplate.DNSNames = []string{domain}
	}

	rootCertBytes, _ := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootPrivateKey.PublicKey, rootPrivateKey)
	rootCertBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rootCertBytes,
	}

	interPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	interTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "1Panel Intermediate CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	if isIP {
		interTemplate.IPAddresses = []net.IP{ipItem}
	} else {
		interTemplate.DNSNames = []string{domain}
	}

	interCertBytes, _ := x509.CreateCertificate(rand.Reader, &interTemplate, &rootTemplate, &interPrivateKey.PublicKey, rootPrivateKey)
	interCertBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: interCertBytes,
	}

	clientPrivateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: domain},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	if isIP {
		clientTemplate.IPAddresses = []net.IP{ipItem}
	} else {
		clientTemplate.DNSNames = []string{domain}
	}

	clientCertBytes, _ := x509.CreateCertificate(rand.Reader, &clientTemplate, &interTemplate, &clientPrivateKey.PublicKey, interPrivateKey)
	clientCertBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertBytes,
	}

	pemBytes := []byte{}
	pemBytes = append(pemBytes, pem.EncodeToMemory(clientCertBlock)...)
	pemBytes = append(pemBytes, pem.EncodeToMemory(interCertBlock)...)
	pemBytes = append(pemBytes, pem.EncodeToMemory(rootCertBlock)...)
	certOut, err := os.OpenFile(global.CONF.System.BaseDir+"/1panel/secret/server.crt.tmp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer certOut.Close()
	if _, err := certOut.Write(pemBytes); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(global.CONF.System.BaseDir+"/1panel/secret/server.key.tmp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientPrivateKey)}); err != nil {
		return err
	}

	return nil
}
