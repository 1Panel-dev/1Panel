package binary

import (
	"io"
	"os"
	"os/exec"

	"github.com/1Panel-dev/1Panel/global"
)

func StartTTY() {
	cmd := "gotty"
	params := []string{"--permit-write", "bash"}
	go func() {
		c := exec.Command(cmd, params...)
		c.Env = append(c.Env, os.Environ()...)
		c.Stdout = io.Discard
		c.Stderr = io.Discard
		if err := c.Run(); err != nil {
			global.LOG.Error(err)
		}
	}()
}
