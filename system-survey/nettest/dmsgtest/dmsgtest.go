package dmsgtest

import (
	"context"
	"fmt"
	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"golang.org/x/net/nettest"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type ResultEntry struct {
	ResponderAccepted string `json:"responder_accepted"`
	InitiatorAccepted string `json:"initiator_accepted"`
	ServerPK          string `json:"server_pk"`
	Error             error  `json:"error"`
}

type Input struct {
	DiscServer string `json:"disc_server"`
	InitPort   uint16 `json:"init_port"`
	RespPort   uint16 `json:"resp_port"`
	Tries      uint   `json:"tries"`
}

func Run(inp *Input) []ResultEntry {
	// instantiate discovery
	var dc disc.APIClient
	if inp.DiscServer == "local" {
		dc = disc.NewMock(0)
	} else {
		dc = disc.NewHTTP(inp.DiscServer)
	}

	// ports to listen by clients. can be any free port
	var initPort, respPort = inp.InitPort, inp.RespPort

	var res []ResultEntry
	for i := uint(0); i < inp.Tries; i++ {
		entry := ResultEntry{}

		respPK, respSK := cipher.GenerateKeyPair()
		initPK, initSK := cipher.GenerateKeyPair()

		maxSessions := 10

		// instantiate server
		sPK, sSK := cipher.GenerateKeyPair()
		srvConf := dmsg.ServerConfig{
			MaxSessions:    maxSessions,
			UpdateInterval: 0,
		}
		srv := dmsg.NewServer(sPK, sSK, dc, &srvConf, nil)

		lis, err := nettest.NewLocalListener("tcp")
		if err != nil {
			panic(err)
		}
		go func() { _ = srv.Serve(lis, "") }() //nolint:errcheck
		time.Sleep(time.Second)

		// instantiate clients
		respC := dmsg.NewClient(respPK, respSK, dc, nil)
		go respC.Serve(context.Background())

		initC := dmsg.NewClient(initPK, initSK, dc, nil)
		go initC.Serve(context.Background())

		time.Sleep(time.Second)

		// bind to port and start listening for incoming messages
		initL, err := initC.Listen(initPort)
		if err != nil {
			entry.Error = fmt.Errorf("error listening by initiator on port %d: %v", initPort, err)
			continue
		}

		// bind to port and start listening for incoming messages
		respL, err := respC.Listen(respPort)
		if err != nil {
			entry.Error = fmt.Errorf("error listening by responder on port %d: %v", respPort, err)
			continue
		}

		// dial responder via DMSG
		initTp, err := initC.DialStream(context.Background(), dmsg.Addr{PK: respPK, Port: respPort})
		if err != nil {
			entry.Error = fmt.Errorf("error dialing responder: %v", err)
			continue
		}

		// Accept connection. `AcceptStream` returns an object exposing `stream` features
		// thus, `Accept` could also be used here returning `net.Conn` interface. depends on your needs
		respTp, err := respL.AcceptStream()
		if err != nil {
			entry.Error = fmt.Errorf("error accepting inititator: %v", err)
			continue
		}

		// initiator writes to it's stream
		payload := strconv.Itoa(rand.Int())
		_, err = initTp.Write([]byte(payload))
		if err != nil {
			entry.Error = fmt.Errorf("error writing to initiator's stream: %v", err)
			continue
		}

		// responder reads from it's stream
		recvBuf := make([]byte, len(payload))
		_, err = respTp.Read(recvBuf)
		if err != nil {
			entry.Error = fmt.Errorf("error reading from responder's stream: %v", err)
			continue
		}

		entry.ResponderAccepted = string(recvBuf)
		log.Printf("Responder accepted: %s", entry.ResponderAccepted)

		// responder writes to it's stream
		_, err = respTp.Write([]byte(payload))
		if err != nil {
			entry.Error = fmt.Errorf("error writing response: %v", err)
			continue
		}

		// initiator reads from it's stream
		initRecvBuf := make([]byte, len(payload))
		_, err = initTp.Read(initRecvBuf)
		if err != nil {
			entry.Error = fmt.Errorf("error reading response: %v", err)
			continue
		}

		entry.InitiatorAccepted = string(initRecvBuf)
		log.Printf("Initiator accepted: %s", entry.InitiatorAccepted)

		entry.ServerPK = srv.LocalPK().String()

		_ = initTp.Close()
		_ = respTp.Close()
		_ = initL.Close()
		_ = respL.Close()
		_ = initC.Close()
		_ = respC.Close()

		res = append(res, entry)
	}

	return res
}
