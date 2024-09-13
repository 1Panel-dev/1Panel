package files

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
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

func (t TarGzArchiver) CompressPro(withDir bool, src, dst, secret, exclusionRules string) error {
	workdir := src
	srcItem := "."
	if withDir {
		workdir = path.Dir(src)
		srcItem = path.Base(src)
	}
	commands := ""

	exMap := make(map[string]struct{})
	exStr := ""
	excludes := strings.Split(exclusionRules, ";")
	excludes = append(excludes, "*.sock")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		if _, ok := exMap[exclude]; ok {
			continue
		}
		exStr += " --exclude "
		exStr += exclude
		exMap[exclude] = struct{}{}
	}

	if len(secret) != 0 {
		commands = fmt.Sprintf("tar -zcf - %s | openssl enc -aes-256-cbc -salt -pbkdf2 -k '%s' -out %s", srcItem, secret, dst)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar zcf %s %s %s", dst, exStr, srcItem)
		global.LOG.Debug(commands)
	}
	return cmd.ExecCmdWithDir(commands, workdir)
}

func (t TarGzArchiver) ExtractPro(src, dst string, secret string) error {
	if _, err := os.Stat(path.Dir(dst)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(dst), os.ModePerm); err != nil {
			return err
		}
	}

	commands := ""
	if len(secret) != 0 {
		commands = fmt.Sprintf("openssl enc -d -aes-256-cbc -salt -pbkdf2 -k '%s' -in %s | tar -zxf - > /root/log", secret, src)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar zxvf %s", src)
		global.LOG.Debug(commands)
	}
	return cmd.ExecCmdWithDir(commands, dst)
}
