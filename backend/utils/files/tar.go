package files

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"strings"
)

type TarArchiver struct {
	Cmd          string
	FilePath     string
	CompressType CompressType
}

func NewTarArchiver(compressType CompressType) *TarArchiver {
	return &TarArchiver{
		Cmd:          "tar",
		CompressType: compressType,
	}
}

func (t TarArchiver) Compress(SourcePaths []string) error {
	return cmd.ExecCmd(fmt.Sprintf("%s %s %s %s", t.Cmd, t.getOptionStr("compress"), t.FilePath, strings.Join(SourcePaths, " ")))
}

func (t TarArchiver) Extract(dstDir string) error {
	return cmd.ExecCmd(fmt.Sprintf("%s %s %s -C %s", t.Cmd, t.getOptionStr("extract"), t.FilePath, dstDir))
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
