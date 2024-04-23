package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/go-sql-driver/mysql"
)

type DBInfo struct {
	Type     string `json:"type"`
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
	Type      string `json:"type"`
	Version   string `json:"version"`
	Format    string `json:"format"`
	TargetDir string `json:"targetDir"`
	FileName  string `json:"fileName"`

	Timeout uint `json:"timeout"` // second
}

type RecoverInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Version    string `json:"version"`
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

func ConnWithSSL(ssl, skipVerify bool, clientKey, clientCert, rootCert string) (string, error) {
	if !ssl {
		return "", nil
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipVerify,
	}
	if len(rootCert) != 0 {
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM([]byte(rootCert)); !ok {
			global.LOG.Error("append certs from pem failed")
			return "", errors.New("unable to append root cert to pool")
		}
		tlsConfig.RootCAs = pool
	}
	if len(clientCert) != 0 && len(clientKey) != 0 {
		cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
		if err != nil {
			return "", err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if err := mysql.RegisterTLSConfig("cloudsql", tlsConfig); err != nil {
		global.LOG.Errorf("register tls config failed, err: %v", err)
		return "", err
	}
	return "&tls=cloudsql", nil
}
