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

	Timeout uint `json:"timeout"` // second
}

type CreateInfo struct {
	Name      string `json:"name"`
	Username  string `json:"userName"`
	Password  string `json:"password"`
	SuperUser bool   `json:"superUser"`

	Timeout uint `json:"timeout"` // second
}

type Privileges struct {
	Username  string `json:"userName"`
	SuperUser bool   `json:"superUser"`

	Timeout uint `json:"timeout"` // second
}

type DeleteInfo struct {
	Name     string `json:"name"`
	Username string `json:"userName"`

	ForceDelete bool `json:"forceDelete"`
	Timeout     uint `json:"timeout"` // second
}

type PasswordChangeInfo struct {
	Username string `json:"userName"`
	Password string `json:"password"`

	Timeout uint `json:"timeout"` // second
}

type BackupInfo struct {
	Name      string `json:"name"`
	TargetDir string `json:"targetDir"`
	FileName  string `json:"fileName"`

	Timeout uint `json:"timeout"` // second
}

type RecoverInfo struct {
	Name       string `json:"name"`
	SourceFile string `json:"sourceFile"`
	Username   string `json:"username"`

	Timeout uint `json:"timeout"` // second
}

type SyncDBInfo struct {
	Name           string `json:"name"`
	From           string `json:"from"`
	PostgresqlName string `json:"postgresqlName"`
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
