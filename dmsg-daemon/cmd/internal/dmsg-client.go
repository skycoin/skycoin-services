package internal

import (
	"context"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/skycoin/skywire/pkg/skyenv"
)

// InitClient creates a dmsg client
func InitClient() (*dmsg.Client, error) {

	if lvl, err := logging.LevelFromString("error"); err == nil {
		logging.SetLevel(lvl)
	}

	cPK, cSK := cipher.GenerateKeyPair()

	// instantiate clients
	initC := dmsg.NewClient(cPK, cSK, disc.NewHTTP(skyenv.DefaultDmsgDiscAddr), nil)
	go initC.Serve(context.Background())

	return initC, nil
}
