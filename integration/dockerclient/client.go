package dockerclient

import (
	"github.com/docker/docker/client"
)

type DockerClient struct {
	cli        *client.Client
	Networks   NetworkOperations
	Images     ImageOperations
	Containers ContainerOperations
}

func NewClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerClient{
		cli:        cli,
		Networks:   NetworkOperations{cli},
		Images:     ImageOperations{cli},
		Containers: ContainerOperations{cli},
	}, nil
}

func (c *DockerClient) Close() error {
	return c.cli.Close()
}
