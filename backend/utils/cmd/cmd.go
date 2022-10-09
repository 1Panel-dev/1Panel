package cmd

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
)

func Exec(cmdStr string) (out string, err error) {
	command := exec.CommandContext(context.Background(), "bash", "-c", cmdStr)

	var wg sync.WaitGroup
	wg.Add(1)

	stdout, err := command.StdoutPipe()
	if err != nil {
		return
	}
	readout := bufio.NewReader(stdout)
	go func() {
		defer wg.Done()
		out = getOutput(readout)
	}()

	err = command.Run()
	if err != nil {
		return
	}
	wg.Wait()
	return
}

func getOutput(reader *bufio.Reader) string {
	var sumOutput string
	outputBytes := make([]byte, 200)
	for {
		n, err := reader.Read(outputBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			sumOutput += err.Error()
		}
		output := string(outputBytes[:n])
		sumOutput += output
	}
	return sumOutput
}
