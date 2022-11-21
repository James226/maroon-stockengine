package dockerclient

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

type Container struct {
	ID string

	cli *client.Client
}

type ContainerOperations struct {
	cli *client.Client
}

type StartContainerOptions struct {
	Name        string
	Image       *Image
	Port        uint16
	Network     *Network
	Environment map[string]string
}

func mapEnv(m map[string]string) []string {
	var env []string
	for k, v := range m {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func (c ContainerOperations) Start(ctx context.Context, opt StartContainerOptions) (*Container, error) {
	err := removeContainer(ctx, c.cli, opt.Name, false)
	portValue := nat.Port(fmt.Sprintf("%d/tcp", opt.Port))

	resp, err := c.cli.ContainerCreate(ctx, &container.Config{
		Image:    opt.Image.Name,
		Hostname: opt.Name,
		ExposedPorts: nat.PortSet{
			portValue: struct{}{},
		},
		Env: mapEnv(opt.Environment),
		Tty: false,
	}, &container.HostConfig{
		NetworkMode: container.NetworkMode(opt.Network.ID),
		PortBindings: nat.PortMap{
			portValue: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(int(opt.Port)),
				},
			},
		},
	}, nil, nil, opt.Name)
	if err != nil {
		return nil, err
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	containerId := resp.ID

	return &Container{ID: containerId, cli: c.cli}, nil
}

func (c *Container) Stop(ctx context.Context, logOutput bool) error {
	return stopContainer(ctx, c.cli, c.ID, logOutput)
}

func stopContainer(ctx context.Context, cli *client.Client, containerId string, logOutput bool) error {
	err := cli.ContainerStop(ctx, containerId, nil)
	if err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Print(err)
		}
	case <-statusCh:
	}

	if logOutput {
		out, err := cli.ContainerLogs(ctx, containerId, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			log.Print(err)
		}

		stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	}

	return nil
}

func getContainerId(ctx context.Context, cli *client.Client, containerName string) (string, error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return "", err
	}

	dockerContainerName := fmt.Sprintf("/%s", containerName)

	for _, cont := range containers {
		for _, name := range cont.Names {
			if name == dockerContainerName {
				return cont.ID, nil
			}
		}
	}

	return "", nil
}

func removeContainer(ctx context.Context, cli *client.Client, container string, logOutput bool) error {
	containerId, err := getContainerId(ctx, cli, container)
	if err != nil {
		return err
	}

	err = stopContainer(ctx, cli, containerId, logOutput)
	if err != nil {
		return err
	}

	err = cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}
