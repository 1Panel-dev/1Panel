package ssl

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/miekg/dns"
	"golang.org/x/crypto/acme"
	"log"
	"net"
	"strings"
	"time"
)

type ManualClient struct {
	client  *acme.Client
	account *model.WebsiteAcmeAccount
	logger  *log.Logger
}

type RequestCertRequest struct {
	WebsiteSSL *model.WebsiteSSL
}

func NewCustomAcmeClient(acmeAccount *model.WebsiteAcmeAccount, logger *log.Logger) (*ManualClient, error) {
	var (
		key crypto.PrivateKey
		err error
	)
	switch KeyType(acmeAccount.KeyType) {
	case KeyEC256, KeyEC384:
		block, _ := pem.Decode([]byte(acmeAccount.PrivateKey))
		key, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	case KeyRSA2048, KeyRSA3072, KeyRSA4096:
		block, _ := pem.Decode([]byte(acmeAccount.PrivateKey))
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}
	if logger == nil {
		logger = log.Default()
	}

	client := &acme.Client{
		Key:          key.(crypto.Signer),
		DirectoryURL: getCaDirURL(acmeAccount.Type, acmeAccount.CaDirURL),
	}
	return &ManualClient{
		client:  client,
		account: acmeAccount,
		logger:  logger,
	}, nil
}

type Resolve struct {
	Key   string
	Value string
	Err   string
}

func (c *ManualClient) GetDNSResolve(ctx context.Context, websiteSSL *model.WebsiteSSL) (map[string]Resolve, error) {
	order, err := c.client.AuthorizeOrder(ctx, acme.DomainIDs(getWebsiteSSLDomains(websiteSSL)...))
	if err != nil {
		return nil, err
	}
	Orders[websiteSSL.ID] = order

	records := make(map[string]Resolve)

	for _, authzURL := range order.AuthzURLs {
		authz, err := c.client.GetAuthorization(ctx, authzURL)
		if err != nil {
			return nil, err
		}
		domain := authz.Identifier.Value

		var dnsChallenge *acme.Challenge
		for _, challenge := range authz.Challenges {
			if challenge.Type == "dns-01" {
				dnsChallenge = challenge
				break
			}
		}

		if dnsChallenge == nil {
			return nil, fmt.Errorf("no DNS-01 challenge found for domain %s", domain)
		}

		txtValue, err := c.client.DNS01ChallengeRecord(dnsChallenge.Token)
		if err != nil {
			return nil, err
		}

		records[domain] = Resolve{
			Key:   fmt.Sprintf("_acme-challenge.%s", domain),
			Value: txtValue,
		}
	}
	return records, nil
}

func queryDNSRecords(domain string) (map[string]string, error) {
	recordName := fmt.Sprintf("_acme-challenge.%s", domain)
	txts, err := net.LookupTXT(recordName)
	if err != nil {
		return nil, err
	}
	records := make(map[string]string)
	if len(txts) > 0 {
		records[recordName] = txts[0]
	}
	return records, nil
}

func (c *ManualClient) handleAuthorization(ctx context.Context, authzURL string, nameservers []string) error {
	authz, err := c.client.GetAuthorization(ctx, authzURL)
	if err != nil {
		return fmt.Errorf("failed to get authorization: %v", err)
	}

	domain := authz.Identifier.Value
	c.logger.Printf("[INFO] [%s] AuthURL: %s", domain, authzURL)

	if authz.Status == acme.StatusValid {
		return nil
	}

	var dnsChallenge *acme.Challenge
	for _, challenge := range authz.Challenges {
		if challenge.Type == "dns-01" {
			dnsChallenge = challenge
			break
		}
	}

	c.logger.Printf("[INFO] [%s] acme: use dns-01 solver", domain)
	if dnsChallenge == nil {
		return fmt.Errorf("no DNS-01 challenge found for domain %s", domain)
	}

	deadline := time.Now().Add(manualDnsTimeout)
	expectedRecord, err := c.client.DNS01ChallengeRecord(dnsChallenge.Token)
	if err != nil {
		return fmt.Errorf("failed to compute DNS challenge record: %v", err)
	}
	c.logger.Printf("[INFO] [%s] acme: Checking TXT record  %s", domain, expectedRecord)

	for {
		c.logger.Printf("[INFO] [%s] acme: Checking DNS record propagation.", domain)
		var currentRecords map[string]string
		var queryErr error
		if len(nameservers) == 0 {
			currentRecords, queryErr = queryDNSRecords(domain)
		} else {
			var errs []string
			for _, nameserver := range nameservers {
				currentRecords, queryErr = queryDNSRecordsWithResolver(ctx, c.logger, domain, nameserver)
				errs = append(errs, fmt.Sprintf("%s: %v", nameserver, queryErr))
			}
			if queryErr != nil && len(errs) > 0 {
				queryErr = fmt.Errorf("all nameservers failed: %s", strings.Join(errs, "; "))
			}
		}
		if currentRecords == nil && queryErr != nil {
			return fmt.Errorf("failed to query DNS records: %v", queryErr)
		}
		recordName := fmt.Sprintf("_acme-challenge.%s", domain)
		providedRecord, exists := currentRecords[recordName]
		if exists && providedRecord == expectedRecord {
			break
		}
		if time.Now().After(deadline) {
			if !exists {
				return fmt.Errorf("TXT record not provided for domain %s after retrying", domain)
			}
			c.logger.Printf("[INFO] [%s] TXT record mismatch for %s: expected %s, got %s\"", domain, domain, expectedRecord, providedRecord)
			return fmt.Errorf("TXT record mismatch for %s: expected %s, got %s", domain, expectedRecord, providedRecord)
		}
		time.Sleep(pollingInterval)
	}

	_, err = c.client.Accept(ctx, dnsChallenge)
	if err != nil {
		return fmt.Errorf("failed to accept challenge: %v", err)
	}
	for {
		time.Sleep(pollingInterval)
		authz, err = c.client.GetAuthorization(ctx, authzURL)
		if err != nil {
			return fmt.Errorf("failed to get authorization while polling: %v", err)
		}
		if authz.Status == acme.StatusValid {
			break
		} else if authz.Status == acme.StatusInvalid {
			return fmt.Errorf("authorization failed for domain %s", domain)
		}
	}
	return nil
}

func (c *ManualClient) createCSR(keyType string, privateKey string, domains []string) ([]byte, crypto.PrivateKey, error) {
	var certKey crypto.PrivateKey
	var err error
	certKey, err = GetPrivateKeyByType(keyType, privateKey)
	if err != nil {
		return nil, nil, err
	}
	template := x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: domains[0]},
		DNSNames: domains,
	}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, certKey)
	if err != nil {
		return nil, nil, err
	}
	return csrBytes, certKey, nil
}

func (c *ManualClient) encodePrivateKey(key crypto.PrivateKey) (string, error) {
	var keyBytes []byte
	var keyType string
	var err error

	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		keyBytes, err = x509.MarshalECPrivateKey(k)
		keyType = "EC PRIVATE KEY"
	case *rsa.PrivateKey:
		keyBytes = x509.MarshalPKCS1PrivateKey(k)
		keyType = "RSA PRIVATE KEY"
	default:
		return "", fmt.Errorf("unsupported key type")
	}

	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  keyType,
		Bytes: keyBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

func (c *ManualClient) RequestCertificate(ctx context.Context, websiteSSL *model.WebsiteSSL) (certificate.Resource, error) {
	var res certificate.Resource
	domains := []string{websiteSSL.PrimaryDomain}
	if websiteSSL.Domains != "" {
		domains = append(domains, strings.Split(websiteSSL.Domains, ",")...)
	}

	c.logger.Printf("[INFO] Requesting certificate for domains: %v\n", domains)
	csr, certKey, err := c.createCSR(websiteSSL.KeyType, websiteSSL.PrivateKey, domains)
	if err != nil {
		return res, err
	}

	order, ok := Orders[websiteSSL.ID]
	if !ok {
		return res, fmt.Errorf("order not found")
	}
	defer delete(Orders, websiteSSL.ID)

	for _, authzURL := range order.AuthzURLs {
		if err := c.handleAuthorization(ctx, authzURL, getNameservers(*websiteSSL)); err != nil {
			return res, err
		}
	}

	c.logger.Printf("[INFO] acme: Validations succeeded; requesting certificates")
	order, err = c.client.WaitOrder(ctx, order.URI)
	if err != nil {
		return res, err
	}

	if order.Status != acme.StatusReady {
		return res, fmt.Errorf("order not ready: %s", order.Status)
	}

	certBytes, certURL, err := c.client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return res, fmt.Errorf("failed to finalize order: %v", err)
	}

	privateKeyPEM, err := c.encodePrivateKey(certKey)
	if err != nil {
		return res, err
	}

	var certPEM []byte
	for _, cert := range certBytes {
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		}
		certPEM = append(certPEM, pem.EncodeToMemory(block)...)
	}
	c.logger.Printf("[INFO] acme: Server responded with a certificate.")
	resource := certificate.Resource{
		Domain:        domains[0],
		CertURL:       certURL,
		CertStableURL: certURL,
		PrivateKey:    []byte(privateKeyPEM),
		Certificate:   certPEM,
		CSR:           csr,
	}
	return resource, nil
}

func getNameservers(websiteSSL model.WebsiteSSL) []string {
	var nameservers []string
	if websiteSSL.Nameserver1 != "" {
		nameservers = append(nameservers, handleNameserver(websiteSSL.Nameserver1))
	}
	if websiteSSL.Nameserver2 != "" {
		nameservers = append(nameservers, handleNameserver(websiteSSL.Nameserver2))
	}
	return nameservers
}

func handleNameserver(nameserver string) string {
	if strings.Contains(nameserver, ":") {
		return nameserver
	}
	return fmt.Sprintf("%s:53", nameserver)
}

func queryDNSRecordsWithResolver(ctx context.Context, logger *log.Logger, domain string, dnsServer string) (map[string]string, error) {
	recordName := fmt.Sprintf("_acme-challenge.%s", domain)
	c := new(dns.Client)
	c.Timeout = 10 * time.Second
	c.Net = "udp"

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(recordName), dns.TypeTXT)
	m.RecursionDesired = true

	r, _, err := c.ExchangeContext(ctx, m, dnsServer)
	if isNetworkError(err) {
		logger.Printf("[WARN] Network error occurred while querying DNS: %v", err)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("DNS query failed: %w", err)
	}

	if r.Rcode == dns.RcodeNameError {
		logger.Printf("[INFO] DNS record does not exist yet (NXDOMAIN)")
		return nil, nil
	}

	if r.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS query failed with code: %s", dns.RcodeToString[r.Rcode])
	}

	records := make(map[string]string)

	for _, answer := range r.Answer {
		if txt, ok := answer.(*dns.TXT); ok {
			if len(txt.Txt) > 0 {
				records[recordName] = txt.Txt[0]
				break
			}
		}
	}

	return records, nil
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "i/o timeout") ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "network is unreachable") ||
		strings.Contains(err.Error(), "no route to host")
}
