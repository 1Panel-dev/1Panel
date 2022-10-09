package docker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

func NewClient() (Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return Client{}, err
	}

	return Client{
		cli: cli,
	}, nil
}

func (c Client) ListAllContainers() ([]types.Container, error) {
	var options types.ContainerListOptions
	containers, err := c.cli.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func (c Client) ListContainersByName(names []string) ([]types.Container, error) {
	var options types.ContainerListOptions
	options.All = true
	if len(names) > 0 {
		var array []filters.KeyValuePair
		for _, n := range names {
			array = append(array, filters.Arg("name", n))
		}
		options.Filters = filters.NewArgs(array...)
	}
	containers, err := c.cli.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func (c Client) ExecCommand(context context.Context, name string, command []string) {
	execConfig := types.ExecConfig{Tty: true, AttachStdout: true, AttachStderr: false, Cmd: command}
	respIdExecCreate, err := c.cli.ContainerExecCreate(context, name, execConfig)
	if err != nil {
		fmt.Println(err)
	}
	respId, err := c.cli.ContainerExecAttach(context, respIdExecCreate.ID, types.ExecStartCheck{})
	if err != nil {
		fmt.Println(err)
	}

	//text, _ := respId.Reader.ReadString('\n')
	//fmt.Printf("%s\n", text)
	scanner := bufio.NewScanner(respId.Reader)
	//text, _ := resp.Reader.ReadString('\n')
	//log.Print(text)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	//
	//respId, err := c.cli.ContainerExecAttach(context, respIdExecCreate.ID, types.ExecStartCheck{})
	//if err != nil {
	//	fmt.Println(err)
	//}

}
