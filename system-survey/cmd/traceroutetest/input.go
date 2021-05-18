package traceroutetest

// Input traceroute settings.
type Input struct {
	DestinationIP   string
	DestinationPort int
	MaxLatency      int
	MaxHops         int
	Retries         int
}
