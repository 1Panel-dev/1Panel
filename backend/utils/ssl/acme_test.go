package ssl

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"os"
	"path"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"gopkg.in/yaml.v3"

	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"log"
)

type AppList struct {
	Version string      `json:"version"`
	Tags    []Tag       `json:"tags"`
	Items   []AppDefine `json:"items"`
}

type NewAppDefine struct {
	Name                 string    `yaml:"name"`
	Tags                 []string  `yaml:"tags"`
	Title                string    `yaml:"title"`
	Type                 string    `yaml:"type"`
	Description          string    `yaml:"description"`
	AdditionalProperties AppDefine `yaml:"additionalProperties"`
}

type NewAppConfig struct {
	AdditionalProperties map[string]interface{} `yaml:"additionalProperties"`
}

type AppDefine struct {
	Key                string   `json:"key" yaml:"key"`
	Name               string   `json:"name" yaml:"name"`
	Tags               []string `json:"tags" yaml:"tags"`
	Versions           []string `json:"versions" yaml:"-"`
	ShortDescZh        string   `json:"shortDescZh" yaml:"shortDescZh"`
	ShortDescEn        string   `json:"shortDescEn" yaml:"shortDescEn"`
	Type               string   `json:"type" yaml:"type"`
	CrossVersionUpdate bool     `json:"crossVersionUpdate" yaml:"crossVersionUpdate"`
	Limit              int      `json:"limit" yaml:"limit"`
	Recommend          int      `json:"recommend" yaml:"recommend"`
	Website            string   `json:"website" yaml:"website"`
	Github             string   `json:"github" yaml:"github"`
	Document           string   `json:"document" yaml:"document"`
}

type Tag struct {
	Key  string `json:"key" yaml:"key"`
	Name string `json:"name" yaml:"name"`
}

func getTagName(key string, tags []Tag) string {
	result := "应用"
	for _, tag := range tags {
		if tag.Key == key {
			return tag.Name
		}
	}
	return result
}

func TestAppToV2(t *testing.T) {
	oldDir := "/Users/wangzhengkun/projects/github.com/1Panel-dev/appstore/apps"
	newDir := "/Users/wangzhengkun/projects/github.com/1Panel-dev/appstore/apps_new"
	listJsonDir := path.Join(oldDir, "list.json")
	fileOp := files.NewFileOp()
	content, err := fileOp.GetContent(listJsonDir)
	if err != nil {
		panic(err)
	}
	appList := &AppList{}
	if err = json.Unmarshal(content, appList); err != nil {
		panic(err)
	}

	for _, appDefine := range appList.Items {
		newAppDefine := &NewAppDefine{
			Name:                 appDefine.Name,
			Tags:                 []string{getTagName(appDefine.Tags[0], appList.Tags)},
			Type:                 getTagName(appDefine.Tags[0], appList.Tags),
			Title:                appDefine.ShortDescZh,
			Description:          appDefine.ShortDescZh,
			AdditionalProperties: appDefine,
		}

		yamlContent, err := yaml.Marshal(newAppDefine)
		if err != nil {
			panic(err)
		}
		oldAppDir := oldDir + "/" + appDefine.Key
		newAppDir := newDir + "/" + appDefine.Key
		if !fileOp.Stat(newAppDir) {
			if err := fileOp.CreateDir(newAppDir, 0755); err != nil {
				panic(err)
			}
		}
		// logo
		oldLogoPath := oldAppDir + "/metadata/logo.png"
		if err := fileOp.CopyFile(oldLogoPath, newAppDir); err != nil {
			panic(err)
		}
		for _, version := range appDefine.Versions {
			oldVersionDir := oldAppDir + "/versions/" + version
			if err := fileOp.CopyDir(oldVersionDir, newAppDir); err != nil {
				panic(err)
			}
			oldConfigPath := oldVersionDir + "/config.json"
			configContent, err := fileOp.GetContent(oldConfigPath)
			if err != nil {
				panic(err)
			}
			var result map[string]interface{}
			if err := json.Unmarshal(configContent, &result); err != nil {
				panic(err)
			}
			newConfigD := &NewAppConfig{}
			newConfigD.AdditionalProperties = result
			configYamlByte, err := yaml.Marshal(newConfigD)
			if err != nil {
				panic(err)
			}
			newVersionDir := newAppDir + "/" + version
			if err := fileOp.WriteFile(newVersionDir+"/data.yml", bytes.NewReader(configYamlByte), 0755); err != nil {
				panic(err)
			}
			if err := fileOp.WriteFile(newAppDir+"/data.yml", bytes.NewReader(yamlContent), 0755); err != nil {
				panic(err)
			}
			_ = fileOp.DeleteFile(newVersionDir + "/config.json")
			oldReadMefile := newVersionDir + "/README.md"
			_ = fileOp.Cut([]string{oldReadMefile}, newAppDir, "", false)
			_ = fileOp.DeleteFile(oldReadMefile)
		}
	}
}

func TestCreatePrivate(t *testing.T) {
	priKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(priKey)
	block := &pem.Block{
		Type:  "privateKey",
		Bytes: derStream,
	}
	file, err := os.Create("private.key")
	if err != nil {
		return
	}
	if err = pem.Encode(file, block); err != nil {
		return
	}
}

func TestSSL(t *testing.T) {

	priKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	myUser := AcmeUser{
		Email: "you2@yours.com",
		Key:   priKey,
	}

	config := lego.NewConfig(&myUser)
	//config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	config.CADirURL = "https://acme.zerossl.com/v2/DV90"
	config.UserAgent = "acm_go/0.0.1"

	config.Certificate.KeyType = certcrypto.RSA2048
	//config.HTTPClient = httpClient

	client, err := lego.NewClient(config)
	if err != nil {
		panic(err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		panic(err)
	}

	myUser.Registration = reg

	//获取证书
	//certificates, err := client.Certificate.Get("https://acme-v02.api.letsencrypt.org/acme/cert/049cb98a8b3ea5a73f08dcdcf89263af8323", true)
	//if err != nil {
	//	panic(err)
	//}
	//certificates, err = client.Certificate.Renew(*certificates, true, true, "")
	//if err != nil {
	//	panic(err)
	//}

	//申请证书
	ewDomain := "tuxpanel.com"

	request := certificate.ObtainRequest{
		Domains: []string{ewDomain},
		// 证书链
		Bundle: true,
	}

	err = client.Challenge.SetDNS01Provider(&manualDnsProvider{}, dns01.AddDNSTimeout(6*time.Minute))
	if err != nil {
		panic(err)
	}

	core, err := api.New(config.HTTPClient, config.UserAgent, config.CADirURL, reg.URI, priKey)
	if err != nil {
		panic(err)
	}
	order, err := core.Orders.New([]string{ewDomain})
	if err != nil {
		panic(err)
	}

	auth, err := core.Authorizations.Get(order.Authorizations[0])
	if err != nil {
		panic(err)
	}

	domain := challenge.GetTargetedDomain(auth)
	chlng, err := challenge.FindChallenge(challenge.DNS01, auth)
	if err != nil {
		panic(err)
	}
	keyAuth, err := core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		panic(err)
	}
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	fmt.Println("fqdn", fqdn, value)

	//
	//keyAuth, err := client.Challenge.core.GetKeyAuthorization(chlng.Token)
	//if err != nil {
	//	panic(err)
	//}
	//

	//httpProvider, err := webroot.NewHTTPProvider("/opt/1Panel/data/apps/nginx/nginx-1/www/wwwroot")
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = client.Challenge.SetHTTP01Provider(httpProvider)
	//if err != nil {
	//	panic(err)
	//}
	//err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("43.142.178.16", "443"))
	//if err != nil {
	//	panic(err)
	//}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		panic(err)
	}

	fmt.Println("---private---")
	fmt.Println(string(certificates.PrivateKey))
	fmt.Println("---.pem---")
	fmt.Println(string(certificates.Certificate))
	fmt.Println("---.domain---")
	fmt.Println(certificates.Domain)
	fmt.Println("---.certUrl---")
	fmt.Println(certificates.CertURL)
	fmt.Println("---.csr---")
	fmt.Println(certificates.CSR)
	fmt.Println("---.cert string---")
	fmt.Println(string(certificates.IssuerCertificate))

	cer1, _ := pem.Decode(certificates.Certificate)

	cert, err := x509.ParseCertificate(cer1.Bytes)
	if err != nil {
		panic(err)
	}
	fmt.Println(cert)

	cer2, _ := pem.Decode(certificates.IssuerCertificate)

	cert2, err := x509.ParseCertificate(cer2.Bytes)
	if err != nil {
		panic(err)
	}
	fmt.Println(cert2)

}

func generateCSR(privateKey crypto.PrivateKey, domain string) ([]byte, error) {
	// 创建证书请求的模板
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: domain,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
	}

	// 生成 CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, err
	}

	// 将 CSR 编码为 PEM 格式
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	// 这里可以将 CSR 写入文件或者返回
	err = os.WriteFile("csr.pem", csrPEM, 0644)
	if err != nil {
		return nil, err
	}

	return csrPEM, nil
}

func TestZeroSSL(t *testing.T) {

	domain := "1panel.store"
	acmeServer := "https://acme.zerossl.com/v2/DV90"

	//acmeServer = "https://api.test4.buypass.no/acme/directory"
	//
	//acmeServer = "https://api.buypass.com/acme/directory"

	priKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	user := AcmeUser{
		Email: "zhengkunwang123@sina.com",
		Key:   priKey,
	}

	//logFile, err := os.OpenFile("/opt/1panel/ssl.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err != nil {
	//	log.Fatalf("Failed to open log file: %v", err)
	//}
	//defer logFile.Close()
	//
	//logger := log.New(logFile, "", log.LstdFlags)
	//legoLogger.Logger = logger

	config := lego.NewConfig(&user)

	// 设置ACME服务器URL
	config.CADirURL = acmeServer
	config.Certificate.KeyType = certcrypto.RSA2048
	config.UserAgent = "acm_go/0.0.1"

	// 创建ACME客户端
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// ZeroSSl

	kid := ""
	hmacEncoded := ""

	eabOptions := registration.RegisterEABOptions{
		TermsOfServiceAgreed: true,
		Kid:                  kid,
		HmacEncoded:          hmacEncoded,
	}

	reg, err := client.Registration.RegisterWithExternalAccountBinding(eabOptions)
	if err != nil {
		log.Fatal(err)
	}

	// ZeroSSl

	//reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	//if err != nil {
	//	log.Fatal(err)
	//}

	user.Registration = reg

	cloudflareConfig := cloudflare.NewDefaultConfig()
	cloudflareConfig.AuthEmail = ""
	cloudflareConfig.AuthKey = ""
	p, err := cloudflare.NewDNSProviderConfig(cloudflareConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Challenge.SetDNS01Provider(p, dns01.AddDNSTimeout(3*time.Minute)); err != nil {
		log.Fatal(err)
	}

	// 申请证书
	pk, err := certcrypto.GeneratePrivateKey(certcrypto.EC256)
	if err != nil {
		return
	}

	request := certificate.ObtainRequest{
		Domains:    []string{domain},
		Bundle:     true,
		PrivateKey: pk,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	// 保存证书
	err = os.WriteFile("certificate.crt", certificates.Certificate, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("private.key", certificates.PrivateKey, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func TestGetEABCre(t *testing.T) {
	res, err := getZeroSSLEabCredentials("zen@11.com")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", res)
}
