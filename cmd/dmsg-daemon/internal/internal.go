package internal

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skycoin/src/util/logging"
)

var dmsgClients []dmsg.Addr

type clientStatus struct {
	Pk     cipher.PubKey
	Online bool
}

func Run(ctx context.Context, tick time.Duration, out io.Writer, csvPath string, l *logging.Logger) {
	log.SetOutput(out)
	// Calling NewTicker method
	d := time.NewTicker(tick)
	dmsgClients := getDmsgClients(l, csvPath)

	for {
		select {
		case <-ctx.Done():
			log.Fatal()
		case <-d.C:
			err := dmsgClientTest(ctx, dmsgClients)
			if err != nil {
				log.Fatal()
			}
		}
	}
}

func dmsgClientTest(ctx context.Context, dmsgClients []dmsg.Addr) error {
	wg := new(sync.WaitGroup)
	wg.Add(len(dmsgClients))

	clientStatuses := make([]clientStatus, len(dmsgClients))

	c, err := InitClient()
	if err != nil {
		return err
	}

	for i, addr := range dmsgClients {
		go func(addr dmsg.Addr, c *dmsg.Client, i int) {
			err := connDmsgClient(ctx, addr, c)
			if err != nil {
				clientStatuses[i] = clientStatus{
					Pk:     addr.PK,
					Online: false,
				}
			} else {
				clientStatuses[i] = clientStatus{
					Pk:     addr.PK,
					Online: true,
				}
			}
			wg.Done()
		}(addr, c, i)
	}

	wg.Wait()
	log.Printf("%v", clientStatuses)
	return nil
}

func connDmsgClient(ctx context.Context, addr dmsg.Addr, c *dmsg.Client) (err error) {

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
	return
}

func readCSV(fileName string) ([][]string, error) {

	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	r := csv.NewReader(f)

	// skip first line
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}

func getDmsgClients(log *logging.Logger, csvPath string) []dmsg.Addr {

	records, err := readCSV(csvPath)
	if err != nil {
		log.Errorf("csv error: %v", err)
	}

	for _, record := range records {
		var dcPK cipher.PubKey
		err := dcPK.Set(record[0])
		if err != nil {
			log.Errorf("pk error: %v", err)
		}
		var addr dmsg.Addr
		if record[1] != "" {
			p, err := strconv.Atoi(record[1])
			if err != nil {
				log.Errorf("port error: %v", err)
			}
			addr = dmsg.Addr{
				PK:   dcPK,
				Port: uint16(p),
			}
		} else {
			addr = dmsg.Addr{
				PK: dcPK,
			}
		}
		dmsgClients = append(dmsgClients, addr)
	}
	return dmsgClients
}
