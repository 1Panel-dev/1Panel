package ssl

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/acme"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	legoacme "github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certcrypto/compat"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
)

var Orders = make(map[uint]*acme.Order)

// parsePrivateKeyPEM decodes a PEM-encoded private key produced by either
// lego v4 (SEC1 for EC, PKCS#1 for RSA) or lego v5 (PKCS#8 for both). It tries
// PKCS#8 first because that is what v5 and any modern tooling emits, then
// falls back to the legacy formats so account keys persisted under v4 keep
// working without manual conversion.
func parsePrivateKeyPEM(pemBytes []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, buserr.New("invalid PEM block")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("PKCS#8 key does not implement crypto.Signer")
		}
		return signer, nil
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("unsupported private key format")
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

// normalizeKeyType maps legacy v4 KeyType strings (P256/P384/2048/...) to the
// v5 form (EC256/EC384/RSA2048/...) using lego v5's official compat helper, so
// any account/SSL row written under v4 keeps loading without manual conversion.
// Unknown values are returned as-is and let the downstream caller error out.
func normalizeKeyType(stored string) KeyType {
	var k compat.KeyTypeCompat
	if err := k.UnmarshalText([]byte(stored)); err != nil {
		return KeyType(stored)
	}
	return KeyType(k)
}

// AcmeUser implements registration.User. Key is crypto.Signer
// (lego v5 changed the field type from crypto.PrivateKey to crypto.Signer).
type AcmeUser struct {
	Email        string
	Registration *legoacme.ExtendedAccount
	Key          crypto.Signer
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *legoacme.ExtendedAccount {
	return u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.Signer {
	return u.Key
}

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

func NewRegisterClient(acmeAccount *model.WebsiteAcmeAccount, proxy *dto.SystemProxy) (*AcmeClient, error) {
	var (
		priKey crypto.Signer
		err    error
	)

	keyType := normalizeKeyType(acmeAccount.KeyType)

	if acmeAccount.PrivateKey != "" {
		signer, parseErr := parsePrivateKeyPEM([]byte(acmeAccount.PrivateKey))
		if parseErr != nil {
			return nil, parseErr
		}
		priKey = signer
	} else {
		generated, genErr := certcrypto.GeneratePrivateKey(keyType)
		if genErr != nil {
			return nil, genErr
		}
		signer, ok := generated.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("generated key does not implement crypto.Signer")
		}
		priKey = signer
	}

	myUser := &AcmeUser{
		Email: acmeAccount.Email,
		Key:   priKey,
	}
	config := NewConfigWithProxy(myUser, acmeAccount.Type, acmeAccount.CaDirURL, proxy)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	var reg *legoacme.ExtendedAccount
	if acmeAccount.Type == "zerossl" || acmeAccount.Type == "google" || acmeAccount.Type == "freessl" || (acmeAccount.Type == "custom" && acmeAccount.UseEAB) {
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
		reg, err = client.Registration.RegisterWithExternalAccountBinding(ctx, eabOptions)
		if err != nil {
			return nil, err
		}
	} else {
		reg, err = client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
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

func getCaDirURL(accountType, customCaURL string) string {
	var caDirURL string
	switch accountType {
	case "letsencrypt":
		caDirURL = "https://acme-v02.api.letsencrypt.org/directory"
	case "zerossl":
		caDirURL = "https://acme.zerossl.com/v2/DV90"
	case "buypass":
		caDirURL = "https://api.buypass.com/acme/directory"
	case "google":
		caDirURL = "https://dv.acme-v02.api.pki.goog/directory"
	case "freessl":
		caDirURL = "https://acmepro.freessl.cn/v2/DV"
	case "custom":
		caDirURL = customCaURL
	}
	return caDirURL
}

func NewConfigWithProxy(user registration.User, accountType, customCaURL string, systemProxy *dto.SystemProxy) *lego.Config {
	var (
		caDirURL      string
		proxyURL      string
		proxyUser     string
		proxyPassword string
	)
	caDirURL = getCaDirURL(accountType, customCaURL)
	if systemProxy != nil {
		proxyURL = fmt.Sprintf("%s://%s:%s", systemProxy.Type, systemProxy.URL, systemProxy.Port)
		proxyUser = systemProxy.User
		proxyPassword = systemProxy.Password
	}
	// lego v5 removed lego.CertificateConfig.KeyType; the key type is now
	// taken from the user's account for each ObtainRequest call.
	return &lego.Config{
		CADirURL:   caDirURL,
		UserAgent:  "1Panel",
		User:       user,
		HTTPClient: createHTTPClientWithProxy(proxyURL, proxyUser, proxyPassword),
		Certificate: lego.CertificateConfig{
			Timeout: 60 * time.Second,
		},
	}
}

func initCertPool() *x509.CertPool {
	customCACertsPath := os.Getenv("LEGO_CA_CERTIFICATES")
	if customCACertsPath == "" {
		return nil
	}

	useSystemCertPool, _ := strconv.ParseBool(os.Getenv("LEGO_CA_SYSTEM_CERT_POOL"))

	caCerts := strings.Split(customCACertsPath, string(os.PathListSeparator))

	certPool, err := lego.CreateCertPool(caCerts, useSystemCertPool)
	if err != nil {
		panic(fmt.Sprintf("create certificates pool: %v", err))
	}

	return certPool
}

func createHTTPClientWithProxy(proxyURL, username, password string) *http.Client {
	var proxyFunc func(*http.Request) (*url.URL, error)
	if proxyURL != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			proxyFunc = http.ProxyFromEnvironment
		} else {
			if username != "" && password != "" {
				parsedURL.User = url.UserPassword(username, password)
			} else if username != "" {
				parsedURL.User = url.User(username)
			}
			proxyFunc = http.ProxyURL(parsedURL)
		}
	} else {
		proxyFunc = func(_ *http.Request) (*url.URL, error) {
			return nil, nil
		}
	}

	return &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			Proxy: proxyFunc,
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   60 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
			TLSClientConfig: &tls.Config{
				ServerName: os.Getenv("LEGO_CA_SERVER_NAME"),
				RootCAs:    initCertPool(),
			},
		},
	}
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
	defer client.CloseIdleConnections()
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

// GetPrivateKeyByType returns a crypto.Signer (lego v5 changed from crypto.PrivateKey).
func GetPrivateKeyByType(keyType, sslPrivateKey string) (crypto.Signer, error) {
	kType := normalizeKeyType(keyType)
	if sslPrivateKey == "" {
		generated, genErr := certcrypto.GeneratePrivateKey(kType)
		if genErr != nil {
			return nil, genErr
		}
		signer, ok := generated.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("generated key does not implement crypto.Signer")
		}
		return signer, nil
	}
	return parsePrivateKeyPEM([]byte(sslPrivateKey))
}

func getWebsiteSSLDomains(websiteSSL *model.WebsiteSSL) []string {
	domains := []string{websiteSSL.PrimaryDomain}
	if websiteSSL.Domains != "" {
		domains = append(domains, strings.Split(websiteSSL.Domains, ",")...)
	}
	return domains
}

const (
	maxRetryAttempts = 3
	retryDelayOn503  = 30 * time.Second
)

// isHTTP503Error checks if an error is an HTTP 503 Service Unavailable error.
func isHTTP503Error(err error) bool {
	if err == nil {
		return false
	}

	// Check for golang.org/x/crypto/acme.Error (used in manual_client.go).
	var acmeErr *acme.Error
	if errors.As(err, &acmeErr) {
		return acmeErr.StatusCode == http.StatusServiceUnavailable
	}

	// Check error message for 503 (fallback for lego library errors).
	errMsg := err.Error()
	return strings.Contains(errMsg, "503") ||
		strings.Contains(errMsg, "Service busy")
}
