package ssl

import (
	"crypto"
	"encoding/json"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/registration"
	"github.com/pkg/errors"
	"time"
)

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

type AcmeClient struct {
	Config *lego.Config
	Client *lego.Client
	User   *AcmeUser
}

func NewAcmeClient(email, privateKey string) (*AcmeClient, error) {
	if email == "" {
		return nil, errors.New("email can not blank")
	}
	if privateKey == "" {
		client, err := NewRegisterClient(email)
		if err != nil {
			return nil, err
		}
		return client, nil
	} else {
		client, err := NewPrivateKeyClient(email, privateKey)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}

type DnsType string

const (
	DnsPod DnsType = "dnsPod"
)

type DNSParam struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Email     string `json:"email"`
	APIkey    string `json:"APIkey"`
}

func (c *AcmeClient) UseDns(dnsType DnsType, params string) error {

	var param DNSParam
	if err := json.Unmarshal([]byte(params), &param); err != nil {
		return err
	}

	var p challenge.Provider
	var err error
	if dnsType == DnsPod {
		dnsPodConfig := dnspod.NewDefaultConfig()
		dnsPodConfig.LoginToken = param.ID + "," + param.Token
		p, err = dnspod.NewDNSProviderConfig(dnsPodConfig)
		if err != nil {
			return err
		}
	}

	return c.Client.Challenge.SetDNS01Provider(p, dns01.AddDNSTimeout(6*time.Minute))
}
func (c *AcmeClient) UseHTTP() {

}

func (c *AcmeClient) GetSSL(domains []string) (certificate.Resource, error) {

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	certificates, err := c.Client.Certificate.Obtain(request)
	if err != nil {
		return certificate.Resource{}, err
	}

	return *certificates, nil
}
