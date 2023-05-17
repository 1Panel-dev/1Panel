package ssl

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
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
			fileOp.CreateDir(newAppDir, 0755)
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
			_ = fileOp.Cut([]string{oldReadMefile}, newAppDir)
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

	//// 本地ACME测试用
	//httpClient := &http.Client{
	//	Transport: &http.Transport{
	//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//	}}

	//priKey, err := rsa.GenerateKey(rand.Reader, 2048)
	//if err != nil {
	//	panic(err)
	//}
	//derStream := x509.MarshalPKCS1PrivateKey(priKey)
	//block := &pem.Block{
	//	Type:  "privateKey",
	//	Bytes: derStream,
	//}
	//file, err := os.Create("private.key")
	//if err != nil {
	//	return
	//}
	//privateByte := pem.EncodeToMemory(block)
	//
	//fmt.Println(string(privateByte))
	//
	//if err = pem.Encode(file, block); err != nil {
	//	return
	//}

	key, err := os.ReadFile("private.key")
	if err != nil {
		panic(err)
	}

	block2, _ := pem.Decode(key)
	priKey, err := x509.ParsePKCS1PrivateKey(block2.Bytes)
	if err != nil {
		panic(err)
	}

	//block, _ := pem.Decode([]byte("vcCzoue3m0ufCzVx673quKBYQOho4uULpwj_P6tR60Q"))
	//priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	//if err != nil {
	//	panic(err)
	//}

	myUser := AcmeUser{
		Email: "you2@yours.com",
		Key:   priKey,
	}

	config := lego.NewConfig(&myUser)
	//config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
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

	//reg, err := client.Registration.ResolveAccountByKey()
	//if err != nil {
	//	panic(err)
	//}

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

	request := certificate.ObtainRequest{
		Domains: []string{"1panel.cloud"},
		// 证书链
		Bundle: true,
	}

	//dns01.NewDNSProviderManual()

	//dnsPodConfig := dnspod.NewDefaultConfig()
	//dnsPodConfig.LoginToken = "1,1"
	//provider, err := dnspod.1(dnsPodConfig)
	//if err != nil {
	//	panic(err)
	//}
	//
	//alidnsConfig := alidns.NewDefaultConfig()
	//alidnsConfig.SecretKey = "1"
	//alidnsConfig.APIKey = "1"
	//p, err := alidns.NewDNSProviderConfig(alidnsConfig)
	//if err != nil {
	//	panic(err)
	//}

	//p, err := dns01.NewDNSProviderManual()
	//if err != nil {
	//	panic(err)
	//}

	err = client.Challenge.SetDNS01Provider(&manualDnsProvider{}, dns01.AddDNSTimeout(6*time.Minute))
	if err != nil {
		panic(err)
	}

	core, err := api.New(config.HTTPClient, config.UserAgent, config.CADirURL, reg.URI, priKey)
	if err != nil {
		panic(err)
	}
	order, err := core.Orders.New([]string{"1panel.cloud"})
	if err != nil {
		panic(err)
	}

	auth, err := core.Authorizations.Get(order.Authorizations[0])
	if err != nil {
		panic(err)
	}
	//core.Challenges.New()

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
