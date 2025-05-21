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

type RarArchiver struct {
}

func NewRarArchiver() ShellArchiver {
	return &RarArchiver{}
}

func (z RarArchiver) Extract(filePath, dstDir string, _ string) error {
	if err := checkCmdAvailability("unrar"); err != nil {
		return err
	}
	return cmd.RunDefaultBashCf("unrar x -y -o+ %q %q", filePath, dstDir)
}

func (z RarArchiver) Compress(sourcePaths []string, dstFile string, _ string) (err error) {
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

	cmdArgs := append([]string{"a", "-r", tmpFile}, relativePaths...)
	cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(baseDir))
	if err = cmdMgr.Run("rar", cmdArgs...); err != nil {
		return err
	}

	if err = op.Mv(tmpFile, dstFile); err != nil {
		return err
	}
	return nil
}
