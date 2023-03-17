package ssl

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type domainError struct {
	Domain string
	Error  error
}

func GetPrivateKey(priKey crypto.PrivateKey) []byte {
	rsaKey := priKey.(*rsa.PrivateKey)
	derStream := x509.MarshalPKCS1PrivateKey(rsaKey)
	block := &pem.Block{
		Type:  "privateKey",
		Bytes: derStream,
	}
	return pem.EncodeToMemory(block)
}

func NewRegisterClient(email string) (*AcmeClient, error) {
	priKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	myUser := &AcmeUser{
		Email: email,
		Key:   priKey,
	}
	config := newConfig(myUser)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	myUser.Registration = reg

	acmeClient := &AcmeClient{
		User:   myUser,
		Client: client,
		Config: config,
	}

	return acmeClient, nil
}

func NewPrivateKeyClient(email string, privateKey string) (*AcmeClient, error) {
	block, _ := pem.Decode([]byte(privateKey))
	priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	myUser := &AcmeUser{
		Email: email,
		Key:   priKey,
	}
	config := newConfig(myUser)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	reg, err := client.Registration.ResolveAccountByKey()
	if err != nil {
		return nil, err
	}
	myUser.Registration = reg

	acmeClient := &AcmeClient{
		User:   myUser,
		Client: client,
		Config: config,
	}

	return acmeClient, nil
}

func newConfig(user *AcmeUser) *lego.Config {
	config := lego.NewConfig(user)
	config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	//config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	config.UserAgent = "acm_go/0.0.1"
	config.Certificate.KeyType = certcrypto.RSA2048
	return config
}
