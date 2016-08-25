package stoppableNetListener

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gurupras/gocommons"
	"github.com/parnurzeal/gorequest"
	"github.com/stretchr/testify/assert"
)

func helloHttp(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}

func getRequest() (gorequest.Response, []error) {
	url := "http://localhost:8081/"
	resp, _, err := gorequest.New().Get(url).
		End()
	return resp, err
}

func TestListener_BadPort(t *testing.T) {
	result := gocommons.InitResult("TestListener-BadPort")

	var err error
	_, err = New(-1)
	assert.True(t, err != nil, "StoppableNetListener created with negative port")

	_, err = New(65536)
	assert.True(t, err != nil, "StoppableNetListener created with out-of-bound port")

	gocommons.HandleResult(t, true, result)
}

// We can ensure a timeout occurs just by waiting longer than the Timeout limit
// This is what the test does
func TestListener_Timeout(t *testing.T) {
	result := gocommons.InitResult("TestListener-Timeout")

	var err error
	var snl *StoppableNetListener

	snl, err = New(32232)
	assert.True(t, err == nil, "StoppableNetListener could not be created", err)

	snl.Timeout = 100 * time.Millisecond
	go snl.Accept()

	time.Sleep(1 * time.Second)

	snl.Stop()
	gocommons.HandleResult(t, true, result)
}

func TestListener(t *testing.T) {
	result := gocommons.InitResult("TestListener")

	snl, err := New(8081)
	assert.Nil(t, err, "Failed to create StoppableNetListener")

	http.HandleFunc("/", helloHttp)
	server := http.Server{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Serve(snl)
	}()

	resp, errors := getRequest()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	payload := buf.String()
	assert.True(t, errors == nil, "Had errors")
	assert.Equal(t, 200, resp.StatusCode, "Failed GET request")
	assert.Equal(t, "OK", payload, "Failed GET request")

	snl.Stop()
	wg.Wait()

	gocommons.HandleResult(t, true, result)
}
