package files

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

type TarGzArchiver struct {
}

func NewTarGzArchiver() ShellArchiver {
	return &TarGzArchiver{}
}

func (t TarGzArchiver) Extract(ctx context.Context, filePath, dstDir string, secret string) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination dir: %w", err)
	}
	if len(secret) != 0 {
		return runTarGzDecryptToDir(cmd.NewCommandMgr(cmd.WithIgnoreExist1()), filePath, dstDir, secret, false)
	} else {
		return runTarGzExtractToDir(cmd.NewCommandMgr(cmd.WithIgnoreExist1()), filePath, dstDir)
	}
}

func (t TarGzArchiver) Compress(ctx context.Context, sourcePaths []string, dstFile string, secret string) error {
	tmpFile := path.Join(global.Dir.TmpDir, fmt.Sprintf("%s%s.tar.gz", common.RandStr(50), time.Now().Format("20060102150405")))
	op := NewFileOp()
	var err error
	defer func() {
		_ = op.DeleteFile(tmpFile)
		if err != nil {
			_ = op.DeleteFile(dstFile)
		}
	}()

	var itemDirs []string
	for _, item := range sourcePaths {
		itemDirs = append(itemDirs, filepath.Base(item))
	}
	aheadDir := filepath.Dir(sourcePaths[0])
	if len(aheadDir) == 0 {
		aheadDir = "/"
	}
	if len(secret) != 0 {
		err = runTarGzEncryptToFile(cmd.NewCommandMgr(cmd.WithContext(ctx)), tmpFile, secret, append([]string{"-C", aheadDir}, itemDirs...)...)
	} else {
		err = runTarGzToFile(cmd.NewCommandMgr(cmd.WithContext(ctx)), tmpFile, append([]string{"-C", aheadDir}, itemDirs...)...)
	}
	if err != nil {
		return err
	}
	if err = op.Mv(tmpFile, dstFile); err != nil {
		return err
	}
	return nil
}

func runTarGzToFile(cmdMgr *cmd.CommandHelper, dst string, tarArgs ...string) error {
	args := append([]string{"-zcf", dst}, tarArgs...)
	return cmdMgr.Run("tar", args...)
}

func runTarGzEncryptToFile(cmdMgr *cmd.CommandHelper, dst, secret string, tarArgs ...string) error {
	args := append([]string{"-zcf", "-"}, tarArgs...)
	_, err := cmdMgr.RunPipe(
		cmd.PipeCommand{Name: "tar", Args: args},
		cmd.PipeCommand{Name: "openssl", Args: []string{"enc", "-aes-256-cbc", "-salt", "-pass", "env:BACKUP_SECRET", "-out", dst}, Env: []string{"BACKUP_SECRET=" + secret}},
	)
	return err
}

func runTarGzExtractToDir(cmdMgr *cmd.CommandHelper, src, dst string) error {
	return cmdMgr.Run("tar", "-zxvf", src, "-C", dst)
}

func runTarGzDecryptToDir(cmdMgr *cmd.CommandHelper, src, dst, secret string, withSalt bool) error {
	opensslArgs := []string{"enc", "-d", "-aes-256-cbc"}
	if withSalt {
		opensslArgs = append(opensslArgs, "-salt")
	}
	opensslArgs = append(opensslArgs, "-pass", "env:BACKUP_SECRET", "-in", src)
	_, err := cmdMgr.RunPipe(
		cmd.PipeCommand{Name: "openssl", Args: opensslArgs, Env: []string{"BACKUP_SECRET=" + secret}},
		cmd.PipeCommand{Name: "tar", Args: []string{"-zxf", "-", "-C", dst}},
	)
	return err
}
