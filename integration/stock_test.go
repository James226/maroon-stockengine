package integration

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/stretchr/testify/assert"
)

func TestStockEndpoint_NotFound(t *testing.T) {

	resp, err := http.Get("http://localhost:10000/location/01/stock/1234567")

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestStockEndpoint_SingleStockQuantity(t *testing.T) {

	db, err := sql.Open("sqlserver", "Server=localhost;User Id=sa;Password=yourStrong!Password;pooling=true;Encrypt=True;TrustServerCertificate=True;")
	assert.NoError(t, err)

	defer db.Close()

	_, err = db.ExecContext(
		context.Background(),
		"INSERT INTO [Stock] ([Item], [Location], [FinalStock], [Uom]) VALUES ('1234567', '01', 76, 'EA')",
	)

	assert.NoError(t, err)

	resp, err := http.Get("http://localhost:10000/location/01/stock/1234567")

	assert.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	json, err := gabs.ParseJSON(body)
	assert.Nil(t, err)

	assert.Equal(t, "1234567", json.Path("item").Data())
	assert.Equal(t, "01", json.Path("location").Data())
	assert.Equal(t, 76.0, json.Path("quantity").Data())
	assert.Equal(t, "EA", json.Path("uom").Data())

	_, err = db.ExecContext(
		context.Background(),
		"DELETE FROM [Stock] WHERE [Item] = '1234567' AND [Location] = '01')",
	)
}
