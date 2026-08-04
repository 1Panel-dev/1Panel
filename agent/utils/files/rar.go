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

type RarArchiver struct {
}

func NewRarArchiver() ShellArchiver {
	return &RarArchiver{}
}

func (z RarArchiver) Extract(ctx context.Context, filePath, dstDir string, _ string) error {
	return z.ExtractWithOptions(ctx, filePath, dstDir, "", false)
}

func (z RarArchiver) ExtractWithOptions(ctx context.Context, filePath, dstDir, _ string, preserveOwner bool) error {
	if err := checkCmdAvailability("unrar"); err != nil {
		return err
	}
	args := []string{"x", "-y", "-o+"}
	if preserveOwner {
		args = append(args, "-oh", "-ol", "-ow")
	}
	args = append(args, filePath, dstDir)
	return cmd.NewCommandMgr(cmd.WithContext(ctx)).Run("unrar", args...)
}

func (z RarArchiver) Compress(ctx context.Context, sourcePaths []string, dstFile string, _ string) (err error) {
	if err = checkCmdAvailability("rar"); err != nil {
		return err
	}
	tmpFile := path.Join(global.Dir.TmpDir, fmt.Sprintf("%s%s.rar", common.RandStr(50), time.Now().Format(constant.DateTimeSlimLayout)))
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

	cmdArgs := append([]string{"a", "-r", "-oh", "-ol", "-ow", tmpFile}, relativePaths...)
	cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(baseDir), cmd.WithContext(ctx))
	if err = cmdMgr.Run("rar", cmdArgs...); err != nil {
		return err
	}

	if err = op.Mv(tmpFile, dstFile); err != nil {
		return err
	}
	return nil
}
