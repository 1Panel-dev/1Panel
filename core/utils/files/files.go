package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	httpUtil "github.com/1Panel-dev/1Panel/core/utils/http"
)

func CopyFile(src, dst string, withName bool) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if path.Base(src) != path.Base(dst) && !withName {
		dst = path.Join(dst, path.Base(src))
	}
	if _, err := os.Stat(path.Dir(dst)); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(path.Dir(dst), os.ModePerm)
		}
	}
	target, err := os.OpenFile(dst+"_temp", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer target.Close()

	if _, err = io.Copy(target, source); err != nil {
		return err
	}
	if err = os.Rename(dst+"_temp", dst); err != nil {
		return err
	}
	return nil
}

func CopyItem(isDir, withName bool, src, dst string) error {
	if path.Base(src) != path.Base(dst) && !withName {
		dst = path.Join(dst, path.Base(src))
	}
	srcInfo, err := os.Stat(path.Dir(src))
	if err != nil {
		return err
	}
	if _, err := os.Stat(path.Dir(dst)); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(path.Dir(dst), srcInfo.Mode())
		}
	}
	cmdStr := fmt.Sprintf(`cp -rf %s %s`, src, dst+"/")
	if !isDir {
		cmdStr = fmt.Sprintf(`cp -f %s %s`, src, dst+"/")
	}
	stdout, err := cmd.Exec(cmdStr)
	if err != nil {
		return fmt.Errorf("handle %s failed, stdout: %s, err: %v", cmdStr, stdout, err)
	}
	return nil
}

func HandleTar(sourceDir, targetDir, name, exclusionRules string, secret string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	excludes := strings.Split(exclusionRules, ",")
	excludeRules := ""
	excludes = append(excludes, "*.sock")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		excludeRules += " --exclude " + exclude
	}
	path := ""
	if strings.Contains(sourceDir, "/") {
		itemDir := strings.ReplaceAll(sourceDir[strings.LastIndex(sourceDir, "/"):], "/", "")
		aheadDir := sourceDir[:strings.LastIndex(sourceDir, "/")]
		if len(aheadDir) == 0 {
			aheadDir = "/"
		}
		path += fmt.Sprintf("-C %s %s", aheadDir, itemDir)
	} else {
		path = sourceDir
	}

	commands := ""

	if len(secret) != 0 {
		extraCmd := "| openssl enc -aes-256-cbc -salt -k '" + secret + "' -out"
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read -zcf %s %s %s %s", " -"+excludeRules, path, extraCmd, targetDir+"/"+name)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read -zcf %s %s %s", targetDir+"/"+name, excludeRules, path)
		global.LOG.Debug(commands)
	}
	stdout, err := cmd.ExecWithTimeOut(commands, 24*time.Hour)
	if err != nil {
		if len(stdout) != 0 {
			global.LOG.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
			return fmt.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
		}
	}
	return nil
}

func HandleUnTar(sourceFile, targetDir string, secret string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}
	commands := ""
	if len(secret) != 0 {
		extraCmd := "openssl enc -d -aes-256-cbc -k '" + secret + "' -in " + sourceFile + " | "
		commands = fmt.Sprintf("%s tar -zxvf - -C %s", extraCmd, targetDir+" > /dev/null 2>&1")
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar zxvfC %s %s", sourceFile, targetDir)
		global.LOG.Debug(commands)
	}

	stdout, err := cmd.ExecWithTimeOut(commands, 24*time.Hour)
	if err != nil {
		global.LOG.Errorf("do handle untar failed, stdout: %s, err: %v", stdout, err)
		return errors.New(stdout)
	}
	return nil
}

func DownloadFileWithProxy(url, dst string) error {
	_, resp, err := httpUtil.HandleGet(url, http.MethodGet, constant.TimeOut5m)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create download file [%s] error, err %s", dst, err.Error())
	}
	defer out.Close()

	reader := bytes.NewReader(resp)
	if _, err = io.Copy(out, reader); err != nil {
		return fmt.Errorf("save download file [%s] error, err %s", dst, err.Error())
	}
	return nil
}
