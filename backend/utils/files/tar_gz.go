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
		extraCmd := "openssl enc -d -aes-256-cbc -k '" + secret + "' -in " + filePath + " | "
		commands = fmt.Sprintf("%s tar -zxvf - -C %s", extraCmd, dstDir+" > /dev/null 2>&1")
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar -zxvf %s %s", filePath+" -C ", dstDir+" > /dev/null 2>&1")
		global.LOG.Debug(commands)
	}
	if err = cmd.ExecCmd(commands); err != nil {
		return err
	}
	return nil
}

func (t TarGzArchiver) Compress(sourcePaths []string, dstFile string, secret string) error {
	var err error
	path := ""
	itemDir := ""
	for _, item := range sourcePaths {
		itemDir += filepath.Base(item) + " "
	}
	aheadDir := dstFile[:strings.LastIndex(dstFile, "/")]
	if len(aheadDir) == 0 {
		aheadDir = "/"
	}
	path += fmt.Sprintf("- -C %s %s", aheadDir, itemDir)
	commands := ""
	if len(secret) != 0 {
		extraCmd := "| openssl enc -aes-256-cbc -salt -k '" + secret + "' -out"
		commands = fmt.Sprintf("tar -zcf %s %s %s", path, extraCmd, dstFile)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar -zcf %s -C %s %s", dstFile, aheadDir, itemDir)
		global.LOG.Debug(commands)
	}
	if err = cmd.ExecCmd(commands); err != nil {
		return err
	}
	return nil
}
