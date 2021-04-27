package skywireself

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/skycoin/skywire/pkg/app/launcher"
	"github.com/skycoin/skywire/pkg/restart"
	"github.com/skycoin/skywire/pkg/skyenv"
	"github.com/skycoin/skywire/pkg/snet"
	"github.com/skycoin/skywire/pkg/visor"
	"github.com/skycoin/skywire/pkg/visor/visorconfig"
)

// NextNonceResponse represents a ServeHTTP response for json encoding
type NextNonceResponse struct {
	Edge      cipher.PubKey `json:"edge"`
	NextNonce Nonce         `json:"next_nonce"`
}

// Nonce is used to sign requests in order to avoid replay attack
type Nonce uint64

func TestSkywireSelf(t *testing.T) {
	t.Run("skywire_services_test", func(t *testing.T) {
		pk, sk := cipher.GenerateKeyPair()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewEncoder(w).Encode(NextNonceResponse{Edge: pk, NextNonce: 1}))
		}))
		defer srv.Close()

		conf := visorconfig.V1{
			Common: &visorconfig.Common{
				PK: pk,
				SK: sk,
			},
			// dmsg-discovery
			Dmsg: &snet.DmsgConfig{
				Discovery:     skyenv.DefaultDmsgDiscAddr,
				SessionsCount: 10,
			},
			// transport discovery
			// address-resolver
			Transport: &visorconfig.V1Transport{
				Discovery:       skyenv.DefaultTpDiscAddr,
				AddressResolver: skyenv.DefaultAddressResolverAddr,
				LogStore: &visorconfig.V1LogStore{
					Type: visorconfig.MemoryLogStore,
				},
				TrustedVisors: nil,
			},
			Routing: &visorconfig.V1Routing{
				SetupNodes:         nil,
				RouteFinder:        skyenv.DefaultRouteFinderAddr,
				RouteFinderTimeout: 0,
			},
			// service discovery
			Launcher: &visorconfig.V1Launcher{
				LocalPath: "local",
				BinPath:   "apps",
				Apps: []launcher.AppConfig{
					{Name: "foo", Port: 1},
					{Name: "bar", AutoStart: true, Port: 2},
				},
				Discovery: &visorconfig.V1AppDisc{
					UpdateInterval: visorconfig.Duration(skyenv.AppDiscUpdateInterval),
					ServiceDisc:    skyenv.DefaultServiceDiscAddr,
				},
			},
		}

		conf.SetLogger(logging.NewMasterLogger())

		defer func() {
			require.NoError(t, os.RemoveAll("local"))
		}()

		_, ok := visor.NewVisor(&conf, restart.CaptureContext())
		require.True(t, ok)
	})
}
