package client

import (
	mathRand "math/rand"
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
	num := []rune("1234567890")
	uppercase := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	lowercase := []rune("abcdefghijklmnopqrstuvwxyz")
	special := []rune(".%@!~_-")

	b := make([]rune, 10)
	for i := 0; i < 2; i++ {
		b[i] = lowercase[mathRand.Intn(len(lowercase))]
	}
	for i := 2; i < 4; i++ {
		b[i] = uppercase[mathRand.Intn(len(uppercase))]
	}
	b[4] = special[mathRand.Intn(len(special))]
	for i := 5; i < 9; i++ {
		b[i] = num[mathRand.Intn(len(num))]
	}

	for i := len(b) - 1; i > 0; i-- {
		j := mathRand.Intn(i + 1)
		b[i], b[j] = b[j], b[i]
	}
	return passwdItem + "-" + (string(b))
}
