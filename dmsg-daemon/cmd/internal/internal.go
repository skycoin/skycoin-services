package internal

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
)

type clientStatus struct {
	Pk     cipher.PubKey
	Online bool
}

func Run(ctx context.Context, tick time.Duration, out io.Writer, dmsgClients []dmsg.Addr) {
	log.SetOutput(out)
	// Calling NewTicker method
	d := time.NewTicker(tick)

	for {
		select {
		case <-ctx.Done():
			log.Fatal()
		case <-d.C:
			err := DmsgClientTest(ctx, dmsgClients)
			if err != nil {
				log.Fatal()
			}
		}
	}
}

func DmsgClientTest(ctx context.Context, dmsgClients []dmsg.Addr) error {
	wg := new(sync.WaitGroup)
	wg.Add(len(dmsgClients))

	clientStatuses := make([]clientStatus, len(dmsgClients))

	c, err := InitClient()
	if err != nil {
		return err
	}

	for i, addr := range dmsgClients {
		go func(addr dmsg.Addr, c *dmsg.Client, i int) {
			err := ConnDmsgClient(ctx, addr, c)
			log.Print(err)
			log.Print(addr)

			if err != nil {
				clientStatuses[i] = clientStatus{
					Pk:     addr.PK,
					Online: false,
				}
			}
			clientStatuses[i] = clientStatus{
				Pk:     addr.PK,
				Online: false,
			}
			wg.Done()
		}(addr, c, i)
	}

	wg.Wait()
	log.Printf("%v", clientStatuses)
	return nil
}

func ConnDmsgClient(ctx context.Context, addr dmsg.Addr, c *dmsg.Client) error {

	// dial responder via DMSG
	conn, err := c.DialStream(ctx, addr)
	if err != nil {
		return err
	}

	// close stream
	err = conn.Close()
	if err != nil {
		return err
	}

	// close client
	_ = c.Close() //nolint:errcheck
	return nil
}
