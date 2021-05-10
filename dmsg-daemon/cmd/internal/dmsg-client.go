package internal

import (
	"context"
	"time"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skywire/pkg/skyenv"
)

// ConnClient connects to dmsg clients
func ConnClient() error {
	var snPK cipher.PubKey

	// convert the pk from string to cipher.PubKey
	err := snPK.Set(skyenv.DefaultSetupPK)
	if err != nil {
		return err
	}

	cPK, cSK := cipher.GenerateKeyPair()

	// instantiate clients
	initC := dmsg.NewClient(cPK, cSK, disc.NewHTTP(skyenv.DefaultDmsgDiscAddr), nil)
	go initC.Serve(context.Background())

	time.Sleep(time.Second)

	// dial responder via DMSG
	initTp, err := initC.DialStream(context.Background(), dmsg.Addr{PK: snPK, Port: skyenv.DmsgSetupPort})
	if err != nil {
		return err
	}

	// close stream
	err = initTp.Close()
	if err != nil {
		return err
	}

	// close client
	return initC.Close()
}
