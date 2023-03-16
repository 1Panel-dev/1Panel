package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
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

	key, err := ioutil.ReadFile("private.key")
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
