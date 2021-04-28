package skywireself

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/skycoin/skywire/pkg/app/launcher"
	"github.com/skycoin/skywire/pkg/restart"
	"github.com/skycoin/skywire/pkg/routing"
	"github.com/skycoin/skywire/pkg/skyenv"
	"github.com/skycoin/skywire/pkg/snet"
	"github.com/skycoin/skywire/pkg/snet/directtp/tptypes"
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
	pk, sk := cipher.GenerateKeyPair()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(&NextNonceResponse{Edge: pk, NextNonce: 1}))
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
		STCP: &snet.STCPConfig{
			LocalAddr: skyenv.DefaultSTCPAddr,
			PKTable:   nil,
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
			LocalPath:  skyenv.DefaultAppLocalPath,
			BinPath:    skyenv.DefaultAppBinPath,
			ServerAddr: skyenv.DefaultAppSrvAddr,
			Apps: []launcher.AppConfig{
				{
					Name:      skyenv.VPNClientName,
					AutoStart: false,
					Port:      routing.Port(skyenv.VPNClientPort),
				},
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

	v, ok := visor.NewVisor(&conf, restart.CaptureContext())
	require.True(t, ok)

	transportTypes := []string{
		tptypes.STCPR,
		tptypes.SUDPH,
		dmsg.Type,
	}

	for _, tType := range transportTypes {
		_, err := v.AddTransport(pk, tType, true, 0)
		require.NoError(t, err)
	}
	// pks := []cipher.PubKey{
	// 	pk,
	// }

	t.Run("skywire_services_test", func(t *testing.T) {

		eSum, err := v.ExtraSummary()
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, eSum.Health.TransportDiscovery)
		require.Equal(t, http.StatusOK, eSum.Health.AddressResolver)

	})

	t.Run("transport_types_test", func(t *testing.T) {

		tps, err := v.DiscoverTransportsByPK(pk)
		require.NoError(t, err)
		for _, tp := range tps {
			require.Equal(t, true, tp.IsUp)
		}
	})

	t.Run("vpn_client_test", func(t *testing.T) {

		// Stary vpn-client
		err := v.StartApp(skyenv.VPNClientName)
		require.NoError(t, err)

		err = v.StopApp(skyenv.VPNClientName)
		require.NoError(t, err)
	})
}
