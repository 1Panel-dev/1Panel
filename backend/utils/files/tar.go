package files

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
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

func (t TarArchiver) Extract(FilePath string, dstDir string, secret string) error {
	return cmd.ExecCmd(fmt.Sprintf("%s %s \"%s\" -C \"%s\"", t.Cmd, t.getOptionStr("extract"), FilePath, dstDir))
}

func (t TarArchiver) Compress(sourcePaths []string, dstFile string, secret string) error {
	return nil
}

func (t TarArchiver) getOptionStr(Option string) string {
	switch t.CompressType {
	case Tar:
		if Option == "compress" {
			return "cvf"
		} else {
			return "xf"
		}
	}
	return ""
}
