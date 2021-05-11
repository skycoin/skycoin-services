package commands

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/buildinfo"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/cmdutil"
	"github.com/skycoin/skycoin-services/dmsg-daemon/cmd/internal"
	"github.com/skycoin/skycoin-services/dmsg-daemon/cmd/internal/api"
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/skycoin/skywire/pkg/skyenv"

	"github.com/spf13/cobra"
)

const defaultTick = 10 * time.Second

var testClients = []string{
	"020011587bf42a45b15f40d6783f5e5320a69a97a7298382103b754f2e3b6b63e9",
	"02001728a88c27b6fa73ebc969dccdbcbd1d4393f267ea10fff2ed8d5eccaca0a4",
	"02004a94f317f3a7f857b4531906e72a0691bf1853e07d17e6632e40240bb11ee1",
	"02004aa9e2caea09fa20d9fb701737627e8df14a0c3ed082416f23857465982757",
}

var (
	sf          cmdutil.ServiceFlags
	addr        string
	tick        time.Duration
	dmsgClients []dmsg.Addr
)

func init() {
	sf.Init(rootCmd, "dmsg_daemon", "")

	rootCmd.Flags().StringVarP(&addr, "addr", "a", ":9090", "address to bind to")
	rootCmd.Flags().DurationVar(&tick, "entry-timeout", defaultTick, "discovery entry timeout")
}

var rootCmd = &cobra.Command{
	Use:   "dmsg-daemon",
	Short: "Dmsg daemon service",
	Run: func(_ *cobra.Command, _ []string) {
		if _, err := buildinfo.Get().WriteTo(os.Stdout); err != nil {
			log.Printf("Failed to output build info: %v", err)
		}

		log := sf.Logger()
		dmsgClients := test(log)

		ctx, cancel := cmdutil.SignalContext(context.Background(), log)
		defer cancel()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		defer func() {
			signal.Stop(signalChan)
			cancel()
		}()

		go func() {
			select {
			case <-signalChan:
				log.Printf("Got SIGINT/SIGTERM, exiting.")
				cancel()
				os.Exit(1)
			case <-ctx.Done():
				log.Printf("Done.")
				os.Exit(1)
			}
		}()

		go internal.Run(ctx, tick, os.Stdout, dmsgClients)

		a := api.NewApi(log)

		log.WithField("addr", addr).Info("Serving discovery API...")
		go func() {
			if err := serve(addr, a); err != nil {
				log.Errorf("serve: %v", err)
				cancel()
			}
		}()
		<-ctx.Done()

	},
}

// Execute executes root CLI command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func serve(addr string, handler http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: handler}
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func test(log *logging.Logger) []dmsg.Addr {
	var snPK cipher.PubKey
	// convert the pk from string to cipher.PubKey
	err := snPK.Set(skyenv.DefaultSetupPK)
	if err != nil {
		log.Errorf("serve: %v", err)
	}

	dmsgClients = append(dmsgClients, dmsg.Addr{PK: snPK, Port: skyenv.DmsgSetupPort})
	for _, c := range testClients {
		var dcPK cipher.PubKey
		err := dcPK.Set(c)
		if err != nil {
			log.Errorf("serve: %v", err)
		}
		dmsgClients = append(dmsgClients, dmsg.Addr{PK: dcPK})
	}
	return dmsgClients
}
