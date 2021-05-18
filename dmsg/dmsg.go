package main

import (
	"context"
	"flag"
	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/cmdutil"
	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skycoin/src/util/logging"
	"io"
	"net/http"
	"os"
)

var (
	dmsgDisc    = "http://dmsg.discovery.skywire.skycoin.com"
	dmsgPort    = uint(80)
	pk, sk      = cipher.GenerateKeyPair()
	log         *logging.Logger
	logFileName = "uuid.log"
)

func init() {
	log = logging.MustGetLogger("dmsg_daemon")
}

func configureLogger() {
	var f io.Writer
	if _, err := os.Stat(logFileName); err == nil {
		f, _ = os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	} else if os.IsNotExist(err) {
		f, _ = os.Create(logFileName)
	}

	logging.SetOutputTo(f)
}

func main() {
	configureLogger()

	flag.StringVar(&dmsgDisc, "disc", dmsgDisc, "dmsg discovery address")
	flag.UintVar(&dmsgPort, "port", dmsgPort, "dmsg port to serve from")
	flag.Var(&sk, "sk", "dmsg secret key")
	flag.Parse()

	// Get daemon UUID
	UUID := getUUID()
	log.WithField("UUID", UUID).Info("Daemon starting...")

	// Instantiate discovery server.
	dc := disc.NewHTTP(dmsgDisc)

	ctx, cancel := cmdutil.SignalContext(context.Background(), log)
	defer cancel()

	// Create new client.
	client := dmsg.NewClient(pk, sk, dc, nil)
	defer func() { log.WithError(client.Close()).Error() }()
	go client.Serve(context.Background())

	select {
	case <-ctx.Done():
		log.WithError(ctx.Err()).Warn()
		return
	case <-client.Ready():
	}

	// Listen connections on port `dmsgPort`.
	lis, err := client.Listen(uint16(dmsgPort))
	if err != nil {
		log.WithError(err).Fatal()
	}

	go func() {
		<-ctx.Done()
		log.WithError(lis.Close()).Error()
	}()

	log.WithField("dmsg_addr", lis.Addr().String()).Info("Serving...")

	// Handle incoming HTTP requests over dmsg.
	log.WithError(http.Serve(lis, getRouter())).Fatal()
}
