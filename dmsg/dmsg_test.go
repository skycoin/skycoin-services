package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/dmsghttp"
	"github.com/skycoin/dmsg/dmsgtest"
	dmsgtest2 "github.com/skycoin/skycoin-services/system-survey/cmd/dmsgtest"
	"github.com/skycoin/skycoin-services/system-survey/cmd/httptest"
	"github.com/skycoin/skycoin-services/system-survey/cmd/traceroutetest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var respLisPort = 1560

func TestDaemon(t *testing.T) {
	t.Cleanup(clearLogFile)

	b, _ := json.Marshal(map[string]interface{}{
		"apps": []string{"wget", "go", "git"},
		"dmsg": &dmsgtest2.Input{
			Tries:      3,
			InitPort:   1563,
			RespPort:   1563,
			DiscServer: "local",
		},
		"traceroute": &traceroutetest.Input{
			DestinationPort: 80,
			DestinationIP:   "8.8.8.8",
			Retries:         10,
			MaxHops:         30,
			MaxLatency:      10,
		},
		"http": &httptest.Input{Addr: "0.0.0.0:8888"},
	})

	c, pk := getHttpClient(t)
	res, _ := c.Post(
		fmt.Sprintf("http://%s:%d/system_survey", pk.String(), respLisPort),
		"application/json",
		bytes.NewBuffer(b),
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
