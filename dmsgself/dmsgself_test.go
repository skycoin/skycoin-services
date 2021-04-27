package dmsgtest

import (
	"context"
	"testing"
	"time"

	"github.com/skycoin/dmsg"
	"github.com/stretchr/testify/require"

	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skywire/pkg/skyenv"
)

func TestClient_RemoteClients(t *testing.T) {
	var snPK cipher.PubKey

	// convert the pk from string to cipher.PubKey
	err := snPK.Set(skyenv.DefaultSetupPK)
	require.NoError(t, err)

	t.Run("dmsg_self_test", func(t *testing.T) {

		// generate keys for client
		cPK, cSK := cipher.GenerateKeyPair()

		// instantiate clients
		initC := dmsg.NewClient(cPK, cSK, disc.NewHTTP(skyenv.DefaultDmsgDiscAddr), nil)
		go initC.Serve(context.Background())

		time.Sleep(time.Second)

		// dial responder via DMSG
		initTp, err := initC.DialStream(context.Background(), dmsg.Addr{PK: snPK, Port: skyenv.DmsgSetupPort})
		require.NoError(t, err)

		// close stream
		err = initTp.Close()
		require.NoError(t, err)

		// close client
		err = initC.Close()
		require.NoError(t, err)
	})
}
