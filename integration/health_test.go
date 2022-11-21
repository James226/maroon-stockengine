package integration_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {

	resp, err := http.Get("http://localhost:10000/health")

	assert.Nil(t, err)

	assert.Equal(t, 200, resp.StatusCode)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	assert.NoError(t, err)
	assert.Equal(t, "Healthy", string(body))
}
