package stoppablenetlistener

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func helloHttp(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}

func getRequest(port int) (gorequest.Response, []error) {
	url := fmt.Sprintf("http://localhost:%v/", port)
	resp, _, err := gorequest.New().Get(url).
		End()
	return resp, err
}

func TestListener_BadPort(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	_, err := New(-1)
	assert.NotNil(err, "Should have failed with negative port")

	_, err = New(65536)
	assert.NotNil(err, "StoppableNetListener created with out-of-bound port")
}

// We can ensure a timeout occurs just by waiting longer than the Timeout limit
// This is what the test does
func TestListener_Timeout(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	snl, err := New(32232)
	assert.Nil(err, "Failed with valid port:", err)

	snl.Timeout = 100 * time.Millisecond
	go snl.Accept()

	time.Sleep(300 * time.Millisecond)

	snl.Stop()
}

func TestListener(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	require := require.New(t)

	port := 8081
	snl, err := New(port)
	snl.Timeout = 100 * time.Millisecond
	assert.Nil(err, "Failed to create StoppableNetListener")
	require.NotNil(snl)

	http.HandleFunc("/", helloHttp)
	server := http.Server{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Serve(snl)
	}()

	resp, errors := getRequest(port)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	payload := buf.String()
	assert.Nil(errors, "Had errors")
	assert.Equal(200, resp.StatusCode, "Failed GET request")
	assert.Equal("OK", payload, "Failed GET request")

	snl.Stop()
	wg.Wait()
}
