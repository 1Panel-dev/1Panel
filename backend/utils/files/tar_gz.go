package files

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type TarGzArchiver struct {
}

func NewTarGzArchiver() ShellArchiver {
	return &TarGzArchiver{}
}

func (t TarGzArchiver) Extract(filePath, dstDir string, secret string) error {
	var err error
	commands := ""
	if len(secret) != 0 {
		extraCmd := fmt.Sprintf("openssl enc -d -aes-256-cbc -k '%s' -in '%s' | ", secret, filePath)
		commands = fmt.Sprintf("%s tar -zxvf - -C '%s' > /dev/null 2>&1", extraCmd, dstDir)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar -zxvf '%s' -C '%s' > /dev/null 2>&1", filePath, dstDir)
		global.LOG.Debug(commands)
	}
	if err = cmd.ExecCmd(commands); err != nil {
		return err
	}
	return nil
}

func (t TarGzArchiver) Compress(sourcePaths []string, dstFile string, secret string) error {
	var itemDirs []string
	for _, item := range sourcePaths {
		itemDirs = append(itemDirs, fmt.Sprintf("\"%s\"", filepath.Base(item)))
	}
	itemDir := strings.Join(itemDirs, " ")
	aheadDir := filepath.Dir(sourcePaths[0])
	if len(aheadDir) == 0 {
		aheadDir = "/"
	}
	commands := ""
	if len(secret) != 0 {
		extraCmd := fmt.Sprintf("| openssl enc -aes-256-cbc -salt -k '%s' -out '%s'", secret, dstFile)
		commands = fmt.Sprintf("tar -zcf - -C \"%s\" %s %s", aheadDir, itemDir, extraCmd)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar -zcf \"%s\" -C \"%s\" %s", dstFile, aheadDir, itemDir)
		global.LOG.Debug(commands)
	}
	if err := cmd.ExecCmd(commands); err != nil {
		return err
	}
	return nil
}
