package integration_test

import (
	"archive/tar"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	_ "github.com/microsoft/go-mssqldb"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	network, err := GetNetworkId(ctx, cli, "stock-engine-integration")
	if err != nil {
		panic(err)
	}

	if network == "" {
		newNetwork, err := cli.NetworkCreate(ctx, "stock-engine-integration", types.NetworkCreate{
			Attachable:     true,
			CheckDuplicate: true,
		})
		if err != nil {
			panic(err)
		}
		network = newNetwork.ID
	}

	if err := PullImage(ctx, cli, "mcr.microsoft.com/azure-sql-edge"); err != nil {
		panic(err)
	}

	sqlContainer, err := StartContainer(cli, ctx, "sql-server", "mcr.microsoft.com/azure-sql-edge", 1433, network, []string{
		"ACCEPT_EULA=Y",
		"MSSQL_SA_PASSWORD=yourStrong!Password",
	})
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlserver", "Server=localhost;User Id=sa;Password=yourStrong!Password;pooling=true;Encrypt=True;TrustServerCertificate=True;")
	if err != nil {
		panic(err)
	}

	err = VerifySqlConnection(ctx, db)
	if err != nil {
		panic(err)
	}

	if err := BuildImage(ctx, cli, "app-container"); err != nil {
		panic(err)
	}

	appContainer, err := StartContainer(cli, ctx, "app-container", "app-container", 10000, network, []string{
		"SQL_CONNECTIONSTRING=Server=sql-server;User Id=sa;Password=yourStrong!Password;pooling=true;Encrypt=True;TrustServerCertificate=True;",
	})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = StopContainer(ctx, cli, appContainer, true)
	_ = StopContainer(ctx, cli, sqlContainer, false)

	os.Exit(code)
}

func VerifySqlConnection(ctx context.Context, db *sql.DB) error {
	var err error

	for i := 0; i < 60; i++ {
		err = db.PingContext(ctx)
		if err == nil {
			return nil
		}

		if i%10 == 0 {
			log.Printf("Failed to connect to sql server on attempt %d: %s", i, err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

func GetNetworkId(ctx context.Context, cli *client.Client, name string) (string, error) {
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return "", err
	}

	for _, network := range networks {
		if network.Name == name {
			return network.ID, nil
		}
	}

	return "", nil
}

func PullImage(ctx context.Context, cli *client.Client, image string) error {
	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	defer reader.Close()
	_, err = io.Copy(os.Stdout, reader)
	return err
}

func BuildImage(ctx context.Context, cli *client.Client, image string) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	CopyAllFiles("../", "", tw)

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	build, err := cli.ImageBuild(ctx, dockerFileTarReader, types.ImageBuildOptions{
		Context:    dockerFileTarReader,
		Dockerfile: "Dockerfile",
		Tags:       []string{image},
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
			fileReader, err := os.Open(filepath.Join(path, item.Name()))
			if err != nil {
				log.Fatal(err, " :unable to open file")
			}
			fileBytes, err := io.ReadAll(fileReader)
			if err != nil {
				log.Fatal(err, " :unable to read file")
			}

			tarHeader := &tar.Header{
				Name: relativePath + "/" + item.Name(),
				Size: int64(len(fileBytes)),
			}
			err = tw.WriteHeader(tarHeader)
			if err != nil {
				log.Fatal(err, " :unable to write tar header")
			}
			_, err = tw.Write(fileBytes)
			if err != nil {
				log.Fatal(err, " :unable to write tar body")
			}
		}
	}
}

func StartContainer(cli *client.Client, ctx context.Context, containerName string, image string, port int, network string, env []string) (string, error) {
	err := RemoveContainer(ctx, cli, containerName, false)

	portValue := nat.Port(fmt.Sprintf("%d/tcp", port))

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:    image,
		Hostname: containerName,
		ExposedPorts: nat.PortSet{
			portValue: struct{}{},
		},
		Env: env,
		Tty: false,
	}, &container.HostConfig{
		NetworkMode: container.NetworkMode(network),
		PortBindings: nat.PortMap{
			portValue: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(port),
				},
			},
		},
	}, nil, nil, containerName)
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	containerId := resp.ID

	return containerId, nil
}

func StopContainer(ctx context.Context, cli *client.Client, containerId string, logOutput bool) error {
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

func GetContainerId(ctx context.Context, cli *client.Client, containerName string) (string, error) {
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

func RemoveContainer(ctx context.Context, cli *client.Client, container string, logOutput bool) error {
	containerId, err := GetContainerId(ctx, cli, container)
	if err != nil {
		return err
	}

	err = StopContainer(ctx, cli, containerId, logOutput)
	if err != nil {
		return err
	}

	err = cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}
