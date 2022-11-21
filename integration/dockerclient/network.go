package dockerclient

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Network struct {
	ID string
}

type NetworkOperations struct {
	cli *client.Client
}

func (n NetworkOperations) Create(ctx context.Context, name string) (*Network, error) {
	network, err := getNetwork(ctx, n.cli, name)
	if err != nil {
		return nil, err
	}

	if network != nil {
		return network, nil
	}

	newNetwork, err := n.cli.NetworkCreate(ctx, name, types.NetworkCreate{
		Attachable:     true,
		CheckDuplicate: true,
	})
	if err != nil {
		panic(err)
	}

	return &Network{ID: newNetwork.ID}, nil
}

func getNetwork(ctx context.Context, cli *client.Client, name string) (*Network, error) {
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if network.Name == name {
			return &Network{ID: network.ID}, nil
		}
	}

	return nil, nil
}
