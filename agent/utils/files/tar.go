package files

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type TarArchiver struct {
	Cmd          string
	CompressType CompressType
}

func NewTarArchiver(compressType CompressType) ShellArchiver {
	return &TarArchiver{
		Cmd:          "tar",
		CompressType: compressType,
	}
}

func (t TarArchiver) Extract(ctx context.Context, FilePath string, dstDir string, secret string) error {
	return t.ExtractWithOptions(ctx, FilePath, dstDir, secret, false)
}

func (t TarArchiver) ExtractWithOptions(ctx context.Context, filePath, dstDir, _ string, preserveOwner bool) error {
	if err := checkCmdAvailability(t.Cmd); err != nil {
		return err
	}
	args := []string{t.getOptionStr("extract"), filePath, "-C", dstDir}
	if preserveOwner {
		args = append([]string{"--same-owner", "--same-permissions"}, args...)
	}
	return cmd.NewCommandMgr(cmd.WithContext(ctx)).Run(t.Cmd, args...)
}

func (t TarArchiver) Compress(ctx context.Context, sourcePaths []string, dstFile string, _ string) error {
	if len(sourcePaths) == 0 {
		return fmt.Errorf("source paths cannot be empty")
	}
	if err := checkCmdAvailability(t.Cmd); err != nil {
		return err
	}
	baseDir := filepath.Dir(sourcePaths[0])
	args := []string{t.getOptionStr("compress"), dstFile, "-C", baseDir}
	for _, sourcePath := range sourcePaths {
		args = append(args, filepath.Base(sourcePath))
	}
	return cmd.NewCommandMgr(cmd.WithContext(ctx)).Run(t.Cmd, args...)
}

func (t TarArchiver) getOptionStr(option string) string {
	switch t.CompressType {
	case Tar:
		if option == "compress" {
			return "-cf"
		}
		return "-xf"
	case Gz, Tgz, TarGz:
		if option == "compress" {
			return "-zcf"
		}
		return "-zxf"
	case Bz2, TarBz2:
		if option == "compress" {
			return "-jcf"
		}
		return "-jxf"
	case Xz, TarXz:
		if option == "compress" {
			return "-Jcf"
		}
		return "-Jxf"
	}
	return ""
}
