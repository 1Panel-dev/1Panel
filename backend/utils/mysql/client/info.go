package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type DBInfo struct {
	From     string `json:"from"`
	Database string `json:"database"`
	Address  string `json:"address"`
	Port     uint   `json:"port"`
	Username string `json:"userName"`
	Password string `json:"password"`

	SSL        bool   `json:"ssl"`
	RootCert   string `json:"rootCert"`
	ClientKey  string `json:"clientKey"`
	ClientCert string `json:"clientCert"`
	SkipVerify bool   `json:"skipVerify"`

	Timeout uint `json:"timeout"` // second
}

type CreateInfo struct {
	Name       string `json:"name"`
	Format     string `json:"format"`
	Version    string `json:"version"`
	Username   string `json:"userName"`
	Password   string `json:"password"`
	Permission string `json:"permission"`

	Timeout uint `json:"timeout"` // second
}

type DeleteInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Username   string `json:"userName"`
	Permission string `json:"permission"`

	ForceDelete bool `json:"forceDelete"`
	Timeout     uint `json:"timeout"` // second
}

type PasswordChangeInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Username   string `json:"userName"`
	Password   string `json:"password"`
	Permission string `json:"permission"`

	Timeout uint `json:"timeout"` // second
}

type AccessChangeInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Username      string `json:"userName"`
	Password      string `json:"password"`
	OldPermission string `json:"oldPermission"`
	Permission    string `json:"permission"`

	Timeout uint `json:"timeout"` // second
}

type BackupInfo struct {
	Name      string `json:"name"`
	Format    string `json:"format"`
	TargetDir string `json:"targetDir"`
	FileName  string `json:"fileName"`

	Timeout uint `json:"timeout"` // second
}

type RecoverInfo struct {
	Name       string `json:"name"`
	Format     string `json:"format"`
	SourceFile string `json:"sourceFile"`

	Timeout uint `json:"timeout"` // second
}

type SyncDBInfo struct {
	Name       string `json:"name"`
	From       string `json:"from"`
	MysqlName  string `json:"mysqlName"`
	Format     string `json:"format"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Permission string `json:"permission"`
}

var formatMap = map[string]string{
	"utf8":    "utf8_general_ci",
	"utf8mb4": "utf8mb4_general_ci",
	"gbk":     "gbk_chinese_ci",
	"big5":    "big5_chinese_ci",
}

func VerifyPeerCertFunc(pool *x509.CertPool) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return errors.New("no certificates available to verify")
		}

		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return err
		}

		opts := x509.VerifyOptions{Roots: pool}
		if _, err = cert.Verify(opts); err != nil {
			return err
		}
		return nil
	}
}

func ConnWithSSL(ssl, skipVerify bool, clientKey, clientCert, rootCert string) (string, error) {
	if !ssl {
		return "", nil
	}
	pool := x509.NewCertPool()
	if len(rootCert) != 0 {
		if ok := pool.AppendCertsFromPEM([]byte(rootCert)); !ok {
			return "", errors.New("unable to append root cert to pool")
		}
	}
	cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
	if err != nil {
		return "", err
	}
	if err := mysql.RegisterTLSConfig("cloudsql", &tls.Config{
		RootCAs:               pool,
		Certificates:          []tls.Certificate{cert},
		InsecureSkipVerify:    skipVerify,
		VerifyPeerCertificate: VerifyPeerCertFunc(pool),
	}); err != nil {
		return "", err
	}
	return "&tls=cloudsql", nil
}
