package ssl

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"io"
	"net/http"
	"net/url"
)

type domainError struct {
	Domain string
	Error  error
}

type zeroSSLRes struct {
	Success    bool   `json:"success"`
	EabKid     string `json:"eab_kid"`
	EabHmacKey string `json:"eab_hmac_key"`
}

type KeyType = certcrypto.KeyType

const (
	KeyEC256   = certcrypto.EC256
	KeyEC384   = certcrypto.EC384
	KeyRSA2048 = certcrypto.RSA2048
	KeyRSA3072 = certcrypto.RSA3072
	KeyRSA4096 = certcrypto.RSA4096
)

func GetPrivateKey(priKey crypto.PrivateKey, keyType KeyType) ([]byte, error) {
	var (
		marshal []byte
		block   *pem.Block
		err     error
	)

	switch keyType {
	case KeyEC256, KeyEC384:
		key := priKey.(*ecdsa.PrivateKey)
		marshal, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, err
		}
		block = &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: marshal,
		}
	case KeyRSA2048, KeyRSA3072, KeyRSA4096:
		key := priKey.(*rsa.PrivateKey)
		marshal = x509.MarshalPKCS1PrivateKey(key)
		block = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: marshal,
		}
	}

	return pem.EncodeToMemory(block), nil
}

func NewRegisterClient(acmeAccount *model.WebsiteAcmeAccount) (*AcmeClient, error) {
	var (
		priKey crypto.PrivateKey
		err    error
	)

	if acmeAccount.PrivateKey != "" {
		switch KeyType(acmeAccount.KeyType) {
		case KeyEC256, KeyEC384:
			block, _ := pem.Decode([]byte(acmeAccount.PrivateKey))
			priKey, err = x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
		case KeyRSA2048, KeyRSA3072, KeyRSA4096:
			block, _ := pem.Decode([]byte(acmeAccount.PrivateKey))
			priKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
		}

	} else {
		priKey, err = certcrypto.GeneratePrivateKey(KeyType(acmeAccount.KeyType))
		if err != nil {
			return nil, err
		}
	}

	myUser := &AcmeUser{
		Email: acmeAccount.Email,
		Key:   priKey,
	}
	config := newConfig(myUser, acmeAccount.Type)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	var reg *registration.Resource
	if acmeAccount.Type == "zerossl" || acmeAccount.Type == "google" {
		if acmeAccount.Type == "zerossl" {
			var res *zeroSSLRes
			res, err = getZeroSSLEabCredentials(acmeAccount.Email)
			if err != nil {
				return nil, err
			}
			if res.Success {
				acmeAccount.EabKid = res.EabKid
				acmeAccount.EabHmacKey = res.EabHmacKey
			} else {
				return nil, fmt.Errorf("get zero ssl eab credentials failed")
			}
		}

		eabOptions := registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  acmeAccount.EabKid,
			HmacEncoded:          acmeAccount.EabHmacKey,
		}
		reg, err = client.Registration.RegisterWithExternalAccountBinding(eabOptions)
		if err != nil {
			return nil, err
		}
	} else {
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, err
		}
	}
	myUser.Registration = reg

	acmeClient := &AcmeClient{
		User:   myUser,
		Client: client,
		Config: config,
	}

	return acmeClient, nil
}

func newConfig(user *AcmeUser, accountType string) *lego.Config {
	config := lego.NewConfig(user)
	switch accountType {
	case "letsencrypt":
		config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	case "zerossl":
		config.CADirURL = "https://acme.zerossl.com/v2/DV90"
	case "buypass":
		config.CADirURL = "https://api.buypass.com/acme/directory"
	case "google":
		config.CADirURL = "https://dv.acme-v02.api.pki.goog/directory"
	}

	config.UserAgent = "1Panel"
	config.Certificate.KeyType = certcrypto.RSA2048
	return config
}

func getZeroSSLEabCredentials(email string) (*zeroSSLRes, error) {
	baseURL := "https://api.zerossl.com/acme/eab-credentials-email"
	params := url.Values{}
	params.Add("email", email)
	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var result zeroSSLRes
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
