package traceroutetest

import (
	"encoding/json"
	"github.com/aeden/traceroute"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"time"
)

const (
	// SerializeJSON result will be serialized to JSON.
	SerializeJSON = iota
	// SerializeByte result will be serialized by encoder.Serialize.
	SerializeByte
)

func Trace(in Input, serializeType int) ([]byte, error) {
	result := Result{StartTime: time.Now().Unix()}

	opts := &traceroute.TracerouteOptions{}
	opts.SetPort(in.DestinationPort)
	opts.SetMaxHops(in.MaxHops)
	opts.SetRetries(in.Retries)
	opts.SetTimeoutMs(in.MaxLatency)

	res, err := traceroute.Traceroute(in.DestinationIP, opts)
	if err != nil {
		return nil, err
	}

	for _, h := range res.Hops {
		result.Hops = append(result.Hops, hop{
			Success:     h.Success,
			Address:     h.AddressString(),
			Host:        h.Host,
			N:           h.N,
			ElapsedTime: h.ElapsedTime,
			TTL:         h.TTL,
		})
	}

	var r []byte

	switch serializeType {
	case SerializeJSON:
		r = encoder.Serialize(result)
	case SerializeByte:
		r, _ = json.Marshal(&result)
	default:
		panic("unknown serialization type")
	}

	return r, nil
}
