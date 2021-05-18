package dmsgtest

import (
	"context"
	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"golang.org/x/net/nettest"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type Result struct {
	DMSGMessages [3]struct {
		ResponderAccepted string `json:"responder_accepted"`
		InitiatorAccepted string `json:"initiator_accepted"`
		ServerPK          string `json:"server_pk"`
	} `json:"dmsg_messages"`
}

func Run() Result {
	// ports to listen by clients. can be any free port
	var initPort, respPort uint16 = 1563, 1563

	// instantiate discovery
	//dc := disc.NewHTTP("http://dmsg.discovery.skywire.skycoin.com")

	res := Result{}
	for i := 0; i < 3; i++ {
		respPK, respSK := cipher.GenerateKeyPair()
		initPK, initSK := cipher.GenerateKeyPair()

		dc := disc.NewMock(0)
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
			log.Fatalf("Error listening by initiator on port %d: %v", initPort, err)
		}

		// bind to port and start listening for incoming messages
		respL, err := respC.Listen(respPort)
		if err != nil {
			log.Fatalf("Error listening by responder on port %d: %v", respPort, err)
		}

		// dial responder via DMSG
		initTp, err := initC.DialStream(context.Background(), dmsg.Addr{PK: respPK, Port: respPort})
		if err != nil {
			log.Fatalf("Error dialing responder: %v", err)
		}

		// Accept connection. `AcceptStream` returns an object exposing `stream` features
		// thus, `Accept` could also be used here returning `net.Conn` interface. depends on your needs
		respTp, err := respL.AcceptStream()
		if err != nil {
			log.Fatalf("Error accepting inititator: %v", err)
		}

		// initiator writes to it's stream
		payload := strconv.Itoa(rand.Int())
		_, err = initTp.Write([]byte(payload))
		if err != nil {
			log.Fatalf("Error writing to initiator's stream: %v", err)
		}

		// responder reads from it's stream
		recvBuf := make([]byte, len(payload))
		_, err = respTp.Read(recvBuf)
		if err != nil {
			log.Fatalf("Error reading from responder's stream: %v", err)
		}

		res.DMSGMessages[i].ResponderAccepted = string(recvBuf)
		log.Printf("Responder accepted: %s", res.DMSGMessages[i].ResponderAccepted)

		// responder writes to it's stream
		_, err = respTp.Write([]byte(payload))
		if err != nil {
			log.Fatalf("Error writing response: %v", err)
		}

		// initiator reads from it's stream
		initRecvBuf := make([]byte, len(payload))
		_, err = initTp.Read(initRecvBuf)
		if err != nil {
			log.Fatalf("Error reading response: %v", err)
		}

		res.DMSGMessages[i].InitiatorAccepted = string(initRecvBuf)
		log.Printf("Initiator accepted: %s", res.DMSGMessages[i].InitiatorAccepted)

		res.DMSGMessages[i].ServerPK = srv.LocalPK().String()

		// close stream
		if err := initTp.Close(); err != nil {
			log.Fatalf("Error closing initiator's stream: %v", err)
		}

		// close stream
		if err := respTp.Close(); err != nil {
			log.Fatalf("Error closing responder's stream: %v", err)
		}

		// close listener
		if err := initL.Close(); err != nil {
			log.Fatalf("Error closing initiator's listener: %v", err)
		}

		// close listener
		if err := respL.Close(); err != nil {
			log.Fatalf("Error closing responder's listener: %v", err)
		}

		// close client
		if err := initC.Close(); err != nil {
			log.Fatalf("Error closing initiator: %v", err)
		}

		// close client
		if err := respC.Close(); err != nil {
			log.Fatalf("Error closing responder: %v", err)
		}
	}

	return res
}
