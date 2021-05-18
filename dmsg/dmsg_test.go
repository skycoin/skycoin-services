package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/dmsghttp"
	"github.com/skycoin/dmsg/dmsgtest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var respLisPort = 1560

func TestDaemon(t *testing.T) {
	t.Cleanup(clearLogFile)
	c, pk := getHttpClient(t)
	res, _ := c.Post(
		fmt.Sprintf("http://%s:%d/system_survey", pk.String(), respLisPort),
		"application/json",
		&bytes.Buffer{},
	)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	res, _ = c.Post(
		fmt.Sprintf("http://%s:%d/system_survey", pk.String(), respLisPort),
		"application/json",
		&bytes.Buffer{},
	)

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func getHttpClient(t *testing.T) (http.Client, cipher.PubKey) {
	env := dmsgtest.NewEnv(t, 0)
	if err := env.Startup(0, 1, 2, nil); err != nil {
		t.Fatal(err)
	}

	clients := env.AllClients()
	initC, respC := clients[0], clients[1]
	go respC.Serve(context.Background())

	respLis, err := respC.Listen(uint16(respLisPort))
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if err = http.Serve(respLis, getRouter()); err != nil {
			t.Error(err)
			return
		}
	}()

	return http.Client{Transport: dmsghttp.MakeHTTPTransport(initC)}, respC.LocalPK()
}
