package files

import (
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type ShellArchiver interface {
	Extract(filePath, dstDir string, secret string) error
	Compress(sourcePaths []string, dstFile string, secret string) error
}

func NewShellArchiver(compressType CompressType) (ShellArchiver, error) {
	switch compressType {
	case Tar:
		if err := checkCmdAvailability("tar"); err != nil {
			return nil, err
		}
		return NewTarArchiver(compressType), nil
	case TarGz:
		return NewTarGzArchiver(), nil
	case Zip:
		if err := checkCmdAvailability("zip"); err != nil {
			return nil, err
		}
		return NewZipArchiver(), nil
	default:
		return nil, buserr.New("unsupported compress type")
	}
}

func checkCmdAvailability(cmdStr string) error {
	if cmd.Which(cmdStr) {
		return nil
	}
	return buserr.WithName("ErrCmdNotFound", cmdStr)
}
