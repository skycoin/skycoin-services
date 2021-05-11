package internal

import (
	"context"
	"io"
	"log"
	"time"
)

func Run(ctx context.Context, tick time.Duration, out io.Writer) {
	log.SetOutput(out)
	// Calling NewTicker method
	d := time.NewTicker(tick)

	for {
		select {
		case <-ctx.Done():
			log.Fatal()
		case <-d.C:
			err := InitClients()
			if err != nil {
				log.Fatal()
			}
		}
	}
}
