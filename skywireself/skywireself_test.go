package skywireself

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/skycoin/skywire/pkg/restart"
	"github.com/skycoin/skywire/pkg/servicedisc"
	"github.com/skycoin/skywire/pkg/skyenv"
	"github.com/skycoin/skywire/pkg/snet/directtp/tptypes"
	"github.com/skycoin/skywire/pkg/syslog"
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

	dmsgDiscAddr := skyenv.TestDmsgDiscAddr
	serviceDiscAddr := skyenv.TestServiceDiscAddr

	// TODO: make sure to setup linode after switching to prod
	var rPK cipher.PubKey
	// PK of visor on linode
	err := rPK.Set("0359272a223ca2c988bd30cb91820f53e802f06120a45dc4c7fd91c1fd246f299b")
	require.NoError(t, err)

	conf := initConfig("skywire-config.json", t)

	conf.SetLogger(logging.NewMasterLogger())

	defer func() {
		require.NoError(t, os.RemoveAll("local"))
	}()

	v, ok := visor.NewVisor(conf, restart.CaptureContext())
	require.True(t, ok)

	transportTypes := []string{
		tptypes.STCPR,
		tptypes.SUDPH,
		dmsg.Type,
	}

	var addedT []uuid.UUID
	for _, tType := range transportTypes {
		tr, err := v.AddTransport(rPK, tType, false, 0)
		require.NoError(t, err)
		addedT = append(addedT, tr.ID)
	}

	t.Run("skywire_services_test", func(t *testing.T) {

		eSum, err := v.ExtraSummary()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, eSum.Health.TransportDiscovery)
		require.Equal(t, http.StatusOK, eSum.Health.AddressResolver)

		// to check if dmsg discovery is working
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_, err = disc.NewHTTP(dmsgDiscAddr).AvailableServers(ctx)
		require.NoError(t, err)

		// to check if service discovery is working
		conf := servicedisc.Config{
			Type:     servicedisc.ServiceTypeVisor,
			PK:       pk,
			SK:       sk,
			Port:     uint16(5505),
			DiscAddr: serviceDiscAddr,
		}

		log := logging.MustGetLogger("appdisc")
		_, err = servicedisc.NewClient(log, conf).Services(ctx)
		require.NoError(t, err)
	})

	t.Run("transport_types_test", func(t *testing.T) {

		tps, err := v.DiscoverTransportsByPK(rPK)
		require.NoError(t, err)

		var workingT []uuid.UUID
		for _, tp := range tps {
			if compare(addedT, tp.Entry.ID) {
				require.Equal(t, true, tp.IsUp)
				workingT = append(workingT, tp.Entry.ID)
			}
		}
		require.Equal(t, 2, len(workingT))
	})

	t.Run("vpn_client_test", func(t *testing.T) {

		err := v.SetAppPK(skyenv.VPNClientName, rPK)
		require.NoError(t, err)

		// Start vpn-client
		err = v.StartApp(skyenv.VPNClientName)
		require.NoError(t, err)

		sum, err := v.GetAppConnectionsSummary(skyenv.VPNClientName)
		require.NoError(t, err)

		if err == nil && len(sum) > 0 {
			require.Equal(t, true, sum[0].IsAlive)
		}

		err = v.StopApp(skyenv.VPNClientName)
		require.NoError(t, err)
	})

	t.Run("close_client", func(t *testing.T) {

		for _, tType := range addedT {
			err := v.RemoveTransport(tType)
			require.NoError(t, err)
		}
		// _ = v.Close()
	})

	defer func() {
		require.NoError(t, os.RemoveAll("apps"))
		require.NoError(t, os.RemoveAll("dmsgpty"))
		require.NoError(t, os.RemoveAll("transport_logs"))
		require.NoError(t, os.RemoveAll("skywire-config.json"))
	}()
}

func compare(slice []uuid.UUID, id uuid.UUID) bool {
	for _, item := range slice {
		if item == id {
			return true
		}
	}
	return false
}

func initConfig(confPath string, t *testing.T) *visorconfig.V1 {
	var r io.Reader
	mLog := initLogger("skywire_selftest", "")
	f, err := os.Open(confPath) //nolint:gosec
	if err != nil {
		require.NoError(t, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			require.NoError(t, err)
		}
	}()
	r = f
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		require.NoError(t, err)
	}

	conf, err := visorconfig.Parse(mLog, confPath, raw)
	if err != nil {
		require.NoError(t, err)
	}

	return conf
}

func initLogger(tag string, syslogAddr string) *logging.MasterLogger {
	log := logging.NewMasterLogger()

	if syslogAddr != "" {
		hook, err := syslog.SetupHook(syslogAddr, tag)
		if err != nil {
			log.WithError(err).Error("Failed to connect to the syslog daemon.")
		} else {
			log.AddHook(hook)
			log.Out = ioutil.Discard
		}
	}
	return log
}
