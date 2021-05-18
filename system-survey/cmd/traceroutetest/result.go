package traceroutetest

import "time"

// Result of the Trace operation.
type Result struct {
	StartTime int64 `json:"start_time"`
	Hops      []hop `json:"hops"`
}

// hop
type hop struct {
	Success     bool          `json:"success"`
	Address     string        `json:"address"`
	Host        string        `json:"host"`
	N           int           `json:"n"`
	ElapsedTime time.Duration `json:"elapsed_time"`
	TTL         int           `json:"ttl"`
}
