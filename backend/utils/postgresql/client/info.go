package client

import (
	_ "github.com/jackc/pgx/v5/stdlib"
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
	Username   string `json:"username"`
	Timeout    uint   `json:"timeout"` // second
}

type SyncDBInfo struct {
	Name           string `json:"name"`
	From           string `json:"from"`
	PostgresqlName string `json:"postgresqlName"`
	Format         string `json:"format"`
	Username       string `json:"username"`
	Password       string `json:"password"`
}
type Status struct {
	Uptime              string `json:"uptime"`
	Version             string `json:"version"`
	MaxConnections      string `json:"max_connections"`
	Autovacuum          string `json:"autovacuum"`
	CurrentConnections  string `json:"current_connections"`
	HitRatio            string `json:"hit_ratio"`
	SharedBuffers       string `json:"shared_buffers"`
	BuffersClean        string `json:"buffers_clean"`
	MaxwrittenClean     string `json:"maxwritten_clean"`
	BuffersBackendFsync string `json:"buffers_backend_fsync"`
}
