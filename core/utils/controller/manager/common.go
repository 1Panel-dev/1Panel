package manager

import (
	"errors"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
)

func handlerErr(out string, err error) error {
	if err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	return nil
}

func run(name string, args ...string) (string, error) {
	global.LOG.Debugf("handle with controller `%s %s`", name, strings.Join(args, " "))
	return cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second)).RunWithStdoutBashCf("LANGUAGE=en_US:en %s %s", name, strings.Join(args, " "))
}
