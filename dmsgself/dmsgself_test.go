package dmsgtest

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skycoin/dmsg"
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/stretchr/testify/require"

	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
)

func TestClient_RemoteClients(t *testing.T) {
	logging.SetLevel(logrus.ErrorLevel)
	const snPort = uint16(22)
	var snPK cipher.PubKey
	pk := "0324579f003e6b4048bae2def4365e634d8e0e3054a20fc7af49daf2a179658557"
	err := snPK.Set(pk)
	require.NoError(t, err)
	dmsgDisc := "http://dmsg.discovery.skywire.skycoin.com"

	t.Run("dmsg_self_test", func(t *testing.T) {

		// generate keys for client
		cPK, cSK := cipher.GenerateKeyPair()

		// instantiate clients
		initC := dmsg.NewClient(cPK, cSK, disc.NewHTTP(dmsgDisc), nil)
		go initC.Serve(context.Background())

		time.Sleep(time.Second)

		// dial responder via DMSG
		initTp, err := initC.DialStream(context.Background(), dmsg.Addr{PK: snPK, Port: snPort})
		require.NoError(t, err)

		// close stream
		err = initTp.Close()
		require.NoError(t, err)

		// close client
		err = initC.Close()
		require.NoError(t, err)
	})
}
