package integration_test

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/james226/dockerclient"

	_ "github.com/microsoft/go-mssqldb"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	c, err := dockerclient.NewClient()
	if err != nil {
		panic(err)
	}

	network, err := c.Networks.Create(ctx, "stock-engine-integration")
	if err != nil {
		panic(err)
	}

	defer c.Close()

	sqlImage, err := c.Images.Pull(ctx, "mcr.microsoft.com/azure-sql-edge")
	if err != nil {
		panic(err)
	}
	sqlContainer, err := c.Containers.Start(ctx, dockerclient.StartContainerOptions{
		Name:    "sql-server",
		Image:   sqlImage,
		Port:    1433,
		Network: network,
		Environment: map[string]string{
			"ACCEPT_EULA":       "Y",
			"MSSQL_SA_PASSWORD": "yourStrong!Password",
		}})
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

	appImage, err := c.Images.Build(ctx, "app-container", "../")
	if err != nil {
		panic(err)
	}

	appContainer, err := c.Containers.Start(ctx, dockerclient.StartContainerOptions{
		Name:    "app-container",
		Image:   appImage,
		Port:    10000,
		Network: network,
		Environment: map[string]string{
			"SQL_CONNECTIONSTRING": "Server=sql-server;User Id=sa;Password=yourStrong!Password;pooling=true;Encrypt=True;TrustServerCertificate=True;",
		}})
	if err != nil {
		panic(err)
	}

	err = VerifyHttp("http://localhost:10000/health")
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = appContainer.Stop(ctx, true)
	_ = sqlContainer.Stop(ctx, false)

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

func VerifyHttp(url string) error {
	var err error

	for i := 0; i < 60; i++ {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}

		if i%10 == 0 {
			log.Printf("Failed to connect to url on attempt %d: %s", i, err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}

	return err
}
