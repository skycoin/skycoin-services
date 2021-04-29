package skywireself

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
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

	var rPK cipher.PubKey
	err := rPK.Set("020011587bf42a45b15f40d6783f5e5320a69a97a7298382103b754f2e3b6b63e9")
	require.NoError(t, err)

	conf := visorconfig.V1{
		Common: &visorconfig.Common{
			PK: pk,
			SK: sk,
		},
		// dmsg-discovery
		Dmsg: &snet.DmsgConfig{
			Discovery:     skyenv.DefaultDmsgDiscAddr,
			SessionsCount: 1,
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

	var addedT []uuid.UUID
	for _, tType := range transportTypes {
		tr, err := v.AddTransport(rPK, tType, true, 0)
		require.NoError(t, err)
		addedT = append(addedT, tr.ID)
	}

	t.Run("skywire_services_test", func(t *testing.T) {

		eSum, err := v.ExtraSummary()
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, eSum.Health.TransportDiscovery)
		require.Equal(t, http.StatusOK, eSum.Health.AddressResolver)

	})

	t.Run("transport_types_test", func(t *testing.T) {

		tps, err := v.DiscoverTransportsByPK(rPK)
		require.NoError(t, err)

		var workingT []uuid.UUID
		t.Logf("%v", addedT)
		for _, tp := range tps {
			if compare(addedT, tp.Entry.ID) {
				t.Logf("%v", tp)
				require.Equal(t, true, tp.IsUp)
				workingT = append(workingT, tp.Entry.ID)
			}
		}
		require.Equal(t, 1, len(workingT))
	})

	t.Run("vpn_client_test", func(t *testing.T) {

		// Stary vpn-client
		err := v.StartApp(skyenv.VPNClientName)
		require.NoError(t, err)

		err = v.StopApp(skyenv.VPNClientName)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.RemoveAll("apps"))
		}()
	})
}

func compare(slice []uuid.UUID, id uuid.UUID) bool {
	for _, item := range slice {
		if item == id {
			return true
		}
	}
	return false
}
