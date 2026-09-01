package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestHeaderParse(t *testing.T) {

	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	value, ok := headers.Get("Host")
	assert.Equal(t, "localhost:42069", value)
	assert.True(t, ok)
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Valid single header with extra space
	headers = NewHeaders()
	data = []byte("Host:            localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	value, ok = headers.Get("Host")
	assert.Equal(t, "localhost:42069", value)
	assert.True(t, ok)
	assert.Equal(t, 34, n)
	assert.False(t, done)

	// Test: Valid two headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nUser-Agent: curl/7.81.0\r\n\r\n")
	nf, donef, errf := headers.Parse(data)
	require.NoError(t, errf)
	require.NotNil(t, headers)
	value, ok = headers.Get("Host")
	assert.Equal(t, "localhost:42069", value)
	assert.True(t, ok)
	assert.Equal(t, 23, nf)
	assert.False(t, donef)

	ns, dones, errs := headers.Parse(data[nf:])
	require.NoError(t, errs)
	require.NotNil(t, headers)
	value, ok = headers.Get("User-Agent")
	assert.Equal(t, "curl", value)
	assert.True(t, ok)
	assert.Equal(t, 25, ns)
	assert.False(t, dones)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with capital header key
	headers = NewHeaders()
	data = []byte("HOST: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	value, ok = headers.Get("Host")
	assert.Equal(t, "localhost:42069", value)
	assert.True(t, ok)
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid Character in Header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid existing header
	headers = NewHeaders()
	data = []byte("Host: localhost:42069, localhost:8080\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	value, ok = headers.Get("Host")
	assert.Equal(t, "localhost:42069, localhost:8080", value)
	assert.True(t, ok)
	assert.Equal(t, 39, n)
	assert.False(t, done)

}
