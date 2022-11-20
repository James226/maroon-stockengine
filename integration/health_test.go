package integration_test

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {

	resp, err := http.Get("http://localhost:10000/health")

	assert.NoError(t, err)

	if err == nil {
		assert.Equal(t, 200, resp.StatusCode)

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		assert.NoError(t, err)
		assert.Equal(t, "Healthy", string(body))
	}
}
