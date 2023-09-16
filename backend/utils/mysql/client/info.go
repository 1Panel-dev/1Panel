package client

import (
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/common"
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

func loadNameByDB(name, version string) string {
	nameItem := common.ConvertToPinyin(name)
	if strings.HasPrefix(version, "5.6") {
		if len(nameItem) <= 16 {
			return nameItem
		}
		return strings.TrimSuffix(nameItem[:10], "_") + "_" + common.RandStr(5)
	}
	if len(nameItem) <= 32 {
		return nameItem
	}
	return strings.TrimSuffix(nameItem[:25], "_") + "_" + common.RandStr(5)
}

func randomPassword(user string) string {
	passwdItem := user
	if len(user) > 6 {
		passwdItem = user[:6]
	}
	return passwdItem + "@" + common.RandStrAndNum(8)
}
