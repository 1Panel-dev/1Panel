package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"testing"
)

func TestCCX(t *testing.T) {
	fmt.Println(GenerateSSL([]string{"172.16.10.111"}))
}

func TestSSLx(t *testing.T) {
	certFile := "/opt/1panel/secret/son1"
	keyFile := "/opt/1panel/secret/son2"

	// 从 PEM 文件中读取证书和密钥
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Println("无法加载证书和密钥：", err)
		return
	}

	// 解析证书信息
	certData, err := os.ReadFile(certFile)
	if err != nil {
		fmt.Println("无法读取证书文件：", err)
		return
	}

	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		fmt.Println("无法解析 PEM 编码的证书数据")
		return
	}

	certObj, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		fmt.Println("无法解析证书数据：", err)
		return
	}

	// 打印证书和密钥信息
	fmt.Println(certObj.IsCA)
	fmt.Printf("证书信息：\nSubject: %s\nIssuer: %s\nNotBefore: %s\nNotAfter: %s\n", certObj.Subject, certObj.Issuer, certObj.NotBefore, certObj.NotAfter)
	fmt.Println("密钥信息：", cert.PrivateKey)
}
