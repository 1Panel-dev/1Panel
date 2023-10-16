package files

import (
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type ShellArchiver interface {
	Compress() error
	Extract(dstDir string) error
}

func NewShellArchiver(compressType CompressType) (*TarArchiver, error) {
	switch compressType {
	case Tar:
		if err := checkCmdAvailability("tar"); err != nil {
			return nil, err
		}
		return NewTarArchiver(compressType), nil
	default:
		return nil, buserr.New("unsupported compress type")
	}
}

func checkCmdAvailability(cmdStr string) error {
	if cmd.Which(cmdStr) {
		return nil
	}
	return buserr.WithName(cmdStr, constant.ErrCmdNotFound)
}
