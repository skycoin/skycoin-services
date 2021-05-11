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

	"github.com/skycoin/dmsg/buildinfo"
	"github.com/skycoin/dmsg/cmdutil"
	"github.com/skycoin/skycoin-services/cmd/dmsg-daemon/internal"
	"github.com/skycoin/skycoin-services/cmd/dmsg-daemon/internal/api"

	"github.com/spf13/cobra"
)

const defaultTick = 60 * time.Second

var (
	sf      cmdutil.ServiceFlags
	addr    string
	csvPath string
	tick    time.Duration
)

func init() {
	sf.Init(rootCmd, "dmsg_daemon", "")

	rootCmd.Flags().StringVarP(&addr, "addr", "a", ":9090", "address to bind to")
	rootCmd.Flags().DurationVar(&tick, "entry-timeout", defaultTick, "discovery entry timeout")
	rootCmd.Flags().StringVarP(&csvPath, "csvPath", "c", "dmsg-clients.csv", "path of csv file")
}

var rootCmd = &cobra.Command{
	Use:   "dmsg-daemon",
	Short: "Dmsg daemon service",
	Run: func(_ *cobra.Command, _ []string) {
		if _, err := buildinfo.Get().WriteTo(os.Stdout); err != nil {
			log.Printf("Failed to output build info: %v", err)
		}

		log := sf.Logger()

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

		go internal.Run(ctx, tick, os.Stdout, csvPath, log)

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
