package manager

import (
	"errors"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
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
	return cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second), cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout(name, args...)
}
