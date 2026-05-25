package files

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

type ZipArchiver struct {
}

func NewZipArchiver() ShellArchiver {
	return &ZipArchiver{}
}

func (z ZipArchiver) Extract(ctx context.Context, filePath, dstDir string, secret string) error {
	if err := checkCmdAvailability("unzip"); err != nil {
		return err
	}
	return cmd.NewCommandMgr(cmd.WithContext(ctx)).Run("unzip", "-qo", filePath, "-d", dstDir)
}

func (z ZipArchiver) Compress(ctx context.Context, sourcePaths []string, dstFile string, _ string) error {
	var err error
	tmpFile := path.Join(global.Dir.TmpDir, fmt.Sprintf("%s%s.zip", common.RandStr(50), time.Now().Format(constant.DateTimeSlimLayout)))
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
	cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(baseDir), cmd.WithContext(ctx))
	args := append([]string{"-qr", tmpFile}, relativePaths...)
	if err = cmdMgr.Run("zip", args...); err != nil {
		return err
	}
	if err = op.Mv(tmpFile, dstFile); err != nil {
		return err
	}
	return nil
}
