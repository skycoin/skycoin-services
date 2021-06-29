package traceroutetest

// Input traceroute settings.
type Input struct {
	DestinationIP   string `json:"destination_ip"`
	DestinationPort int    `json:"destination_port"`
	MaxLatency      int    `json:"max_latency"`
	MaxHops         int    `json:"max_hops"`
	Retries         int    `json:"retries"`
}
