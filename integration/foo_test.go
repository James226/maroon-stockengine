package integration_test

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSomething(t *testing.T) {

	resp, err := http.Get("http://localhost:10000/")

	assert.NoErrorf(t, err, "Expected no error")

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	assert.Equal(t, "true", strings.Trim(string(body), "\n"))
}
