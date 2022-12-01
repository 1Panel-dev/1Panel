package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/cenkalti/backoff/v4"
	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type plainDnsProvider struct {
	Resolve
}

func (p *plainDnsProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	p.Key = fqdn
	p.Value = value
	return nil
}

func (p *plainDnsProvider) CleanUp(domain, token, keyAuth string) error {
	fmt.Sprintf("%s,%s,%s", domain, token, keyAuth)
	return nil
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

func checkChallengeStatus(chlng acme.ExtendedChallenge) (bool, error) {
	switch chlng.Status {
	case acme.StatusValid:
		return true, nil
	case acme.StatusPending, acme.StatusProcessing:
		return false, nil
	case acme.StatusInvalid:
		return false, chlng.Error
	default:
		return false, errors.New("the server returned an unexpected state")
	}
}

func checkAuthorizationStatus(authz acme.Authorization) (bool, error) {
	switch authz.Status {
	case acme.StatusValid:
		return true, nil
	case acme.StatusPending, acme.StatusProcessing:
		return false, nil
	case acme.StatusDeactivated, acme.StatusExpired, acme.StatusRevoked:
		return false, fmt.Errorf("the authorization state %s", authz.Status)
	case acme.StatusInvalid:
		for _, chlg := range authz.Challenges {
			if chlg.Status == acme.StatusInvalid && chlg.Error != nil {
				return false, chlg.Error
			}
		}
		return false, fmt.Errorf("the authorization state %s", authz.Status)
	default:
		return false, errors.New("the server returned an unexpected state")
	}
}

func validate(core *api.Core, domain string, chlg acme.Challenge) error {
	chlng, err := core.Challenges.New(chlg.URL)
	if err != nil {
		return fmt.Errorf("failed to initiate challenge: %w", err)
	}

	valid, err := checkChallengeStatus(chlng)
	if err != nil {
		return err
	}

	if valid {
		return nil
	}

	ra, err := strconv.Atoi(chlng.RetryAfter)
	if err != nil {
		// The ACME server MUST return a Retry-After.
		// If it doesn't, we'll just poll hard.
		// Boulder does not implement the ability to retry challenges or the Retry-After header.
		// https://github.com/letsencrypt/boulder/blob/master/docs/acme-divergences.md#section-82
		ra = 5
	}
	initialInterval := time.Duration(ra) * time.Second

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = initialInterval
	bo.MaxInterval = 10 * initialInterval
	bo.MaxElapsedTime = 100 * initialInterval

	// After the path is sent, the ACME server will access our server.
	// Repeatedly check the server for an updated status on our request.
	operation := func() error {
		authz, err := core.Authorizations.Get(chlng.AuthorizationURL)
		if err != nil {
			return backoff.Permanent(err)
		}

		valid, err := checkAuthorizationStatus(authz)
		if err != nil {
			return backoff.Permanent(err)
		}

		if valid {
			return nil
		}

		return errors.New("the server didn't respond to our request")
	}

	return backoff.Retry(operation, bo)
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

	err = client.Challenge.SetDNS01Provider(&plainDnsProvider{}, dns01.AddDNSTimeout(6*time.Minute))
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

func TestDaoqi(t *testing.T) {
	conn, err := tls.Dial("tcp", "1panel.xyz:443", nil)
	if err != nil {
		panic(err)
	}

	conn2 := conn.ConnectionState()

	fmt.Println(conn2.PeerCertificates[0].NotBefore)

	fmt.Println(conn.ConnectionState().PeerCertificates[0].AuthorityKeyId)
	fmt.Println(string(conn.ConnectionState().PeerCertificates[0].SubjectKeyId))
}

func Test111(t *testing.T) {
	out, err := cmd.Exec("docker exec -i 1Panel-nginx1.23.1-AiCt curl http://127.0.0.1/nginx_status")
	if err != nil {
		panic(err)
	}
	outArray := strings.Split(out, " ")
	fmt.Println(outArray)
	fmt.Println(outArray[8])
}
