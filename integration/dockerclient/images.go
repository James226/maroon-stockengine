package dockerclient

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Image struct {
	Name string
}

type ImageOperations struct {
	cli *client.Client
}

func (i ImageOperations) Pull(ctx context.Context, name string) (*Image, error) {
	reader, err := i.cli.ImagePull(ctx, name, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}

	defer reader.Close()
	_, err = io.Copy(os.Stdout, reader)
	return &Image{name}, err
}

func (i ImageOperations) Build(ctx context.Context, name string, path string) (*Image, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	CopyAllFiles(path, "", tw)

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	build, err := i.cli.ImageBuild(ctx, dockerFileTarReader, types.ImageBuildOptions{
		Context:    dockerFileTarReader,
		Dockerfile: "Dockerfile",
		Tags:       []string{name},
		Remove:     true})
	if err != nil {
		return nil, err
	}

	defer build.Body.Close()
	_, err = io.Copy(os.Stdout, build.Body)

	if err != nil {
		return nil, err
	}

	return &Image{name}, nil
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
