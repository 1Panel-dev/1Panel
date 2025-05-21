package files

import (
	"fmt"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

type X7zArchiver struct {
}

func NewX7zArchiver() ShellArchiver {
	return &X7zArchiver{}
}

func (z X7zArchiver) Extract(filePath, dstDir string, _ string) error {
	if err := checkCmdAvailability("7z"); err != nil {
		return err
	}
	return cmd.RunDefaultBashCf("7z x -y -o%q %q", dstDir, filePath)
}

func (z X7zArchiver) Compress(sourcePaths []string, dstFile string, _ string) (err error) {
	if err = checkCmdAvailability("7z"); err != nil {
		return err
	}
	tmpFile := path.Join(global.Dir.TmpDir, fmt.Sprintf("%s%s.7z", common.RandStr(50), time.Now().Format(constant.DateTimeSlimLayout)))
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

	cmdArgs := append([]string{"a", "-r", tmpFile}, relativePaths...)
	cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(baseDir))
	if err = cmdMgr.Run("7z", cmdArgs...); err != nil {
		return err
	}

	if err = op.Mv(tmpFile, dstFile); err != nil {
		return err
	}
	return nil
}
