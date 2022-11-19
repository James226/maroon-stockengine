package integration_test

import (
	"archive/tar"
	"bytes"
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	if err := BuildImage(ctx, cli); err != nil {
		panic(err)
	}

	containerId, err := StartContainer(cli, ctx)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	if err := StopContainer(cli, ctx, containerId); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func BuildImage(ctx context.Context, cli *client.Client) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	CopyAllFiles("../", "", tw)

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	build, err := cli.ImageBuild(ctx, dockerFileTarReader, types.ImageBuildOptions{
		Context:    dockerFileTarReader,
		Dockerfile: "Dockerfile",
		Tags:       []string{"foo"},
		Remove:     true})
	if err != nil {
		return err
	}

	defer build.Body.Close()
	_, err = io.Copy(os.Stdout, build.Body)

	if err != nil {
		return err
	}

	return nil
}

func CopyAllFiles(path string, relativePath string, tw *tar.Writer) {
	items, _ := os.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			CopyAllFiles(filepath.Join(path, item.Name()), filepath.Join(relativePath, item.Name()), tw)
		} else {
			dockerFileReader, err := os.Open(filepath.Join(path, item.Name()))
			if err != nil {
				log.Fatal(err, " :unable to open file")
			}
			readDockerFile, err := io.ReadAll(dockerFileReader)
			if err != nil {
				log.Fatal(err, " :unable to read file")
			}

			tarHeader := &tar.Header{
				Name: filepath.Join(relativePath, item.Name()),
				Size: int64(len(readDockerFile)),
			}
			err = tw.WriteHeader(tarHeader)
			if err != nil {
				log.Fatal(err, " :unable to write tar header")
			}
			_, err = tw.Write(readDockerFile)
			if err != nil {
				log.Fatal(err, " :unable to write tar body")
			}
		}
	}
}

func StartContainer(cli *client.Client, ctx context.Context) (string, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "foo",
		ExposedPorts: nat.PortSet{
			"10000/tcp": struct{}{},
		},
	}, &container.HostConfig{
		AutoRemove: true,
		PortBindings: nat.PortMap{
			"10000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "10000",
				},
			},
		},
	}, nil, nil, "Foo")
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	containerId := resp.ID
	return containerId, nil
}

func StopContainer(cli *client.Client, ctx context.Context, containerId string) error {
	err := cli.ContainerStop(ctx, containerId, nil)
	if err != nil {
		return err
	}
	return nil
}
