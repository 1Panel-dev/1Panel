package files

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
)

type ZipArchiver struct {
}

func NewZipArchiver() ShellArchiver {
	return &ZipArchiver{}
}

func (z ZipArchiver) Extract(filePath, dstDir string, secret string) error {
	if err := checkCmdAvailability("unzip"); err != nil {
		return err
	}
	return cmd.ExecCmd(fmt.Sprintf("unzip -qo %s -d %s", filePath, dstDir))
}

func (z ZipArchiver) Compress(sourcePaths []string, dstFile string, _ string) error {
	var err error
	tmpFile := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("%s%s.zip", common.RandStr(50), time.Now().Format(constant.DateTimeSlimLayout)))
	op := NewFileOp()
	defer func() {
		_ = op.DeleteFile(tmpFile)
		if err != nil {
			_ = op.DeleteFile(dstFile)
		}
	}()
	baseDir := path.Dir(sourcePaths[0])
	relativePaths := make([]string, len(sourcePaths))
	for i, sp := range sourcePaths {
		relativePaths[i] = path.Base(sp)
	}
	cmdStr := fmt.Sprintf("zip -qr %s  %s", tmpFile, strings.Join(relativePaths, " "))
	if err = cmd.ExecCmdWithDir(cmdStr, baseDir); err != nil {
		return err
	}
	if err = op.Mv(tmpFile, dstFile); err != nil {
		return err
	}
	return nil
}
