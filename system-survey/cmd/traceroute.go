package cmd

import (
	"fmt"
	"github.com/aeden/traceroute"
	"time"
)

type TracerouteInput struct {
	DestinationIP   string `json:"destination_ip"`
	DestinationPort int    `json:"destination_port"`
	MaxHops         int    `json:"max_hops"`
	Retries         int    `json:"retries"`
	Timeout         int    `json:"timeout"`
	FirstHop        int    `json:"first_hop"`
	PacketSize      int    `json:"packet_size"`
}

func (input TracerouteInput) String() string {
	return fmt.Sprintf("Destination: %s:%d\n", input.DestinationIP, input.DestinationPort) +
		fmt.Sprintln("Max hops:", input.MaxHops) +
		fmt.Sprintln("Retries:", input.Retries) +
		fmt.Sprintln("Timeout:", input.Timeout) +
		fmt.Sprintln("First hop:", input.FirstHop) +
		fmt.Sprint("Packet size:", input.PacketSize)
}

type TracerouteHop struct {
	Success     bool          `json:"success"`
	Address     string        `json:"address"`
	Host        string        `json:"host"`
	N           int           `json:"n"`
	ElapsedTime time.Duration `json:"elapsed_time"`
	TTL         int           `json:"ttl"`
}

func (hop TracerouteHop) String() string {
	return fmt.Sprintf("Success: %v, Address: %s, Host: %s, N: %d, Elasped time: %d, TTL: %d", hop.Success,
		hop.Address, hop.Host, hop.N, hop.ElapsedTime, hop.TTL)
}

type TraceRouteResult struct {
	InputData TracerouteInput `json:"input_data"`
	Time      int64           `json:"time"`
	TotalHops int             `json:"total_hops"`
	Hops      []TracerouteHop `json:"hops"`
}

func (result TraceRouteResult) String() string {
	out := fmt.Sprintf("Time: %d\nTotal hops:%d\n%s\nHops:\n", result.Time, result.TotalHops, result.InputData)
	for i, h := range result.Hops {
		out += fmt.Sprintf("[%d] %s\n", i+1, h)
	}
	return out
}

// Traceroute ...
func Traceroute(input TracerouteInput) (TraceRouteResult, error) {
	options := &traceroute.TracerouteOptions{}
	options.SetPort(input.DestinationPort)
	options.SetMaxHops(input.MaxHops)
	options.SetRetries(input.Retries)
	options.SetTimeoutMs(input.Timeout)
	options.SetPacketSize(input.FirstHop)
	options.SetPacketSize(input.PacketSize)

	timeNow := time.Now().Unix()
	trResult, err := traceroute.Traceroute(input.DestinationIP, options)
	if err != nil {
		return TraceRouteResult{}, err
	}

	result := TraceRouteResult{
		InputData: input,
		Time:      timeNow,
		TotalHops: len(trResult.Hops),
	}

	for _, hop := range trResult.Hops {
		result.Hops = append(result.Hops, TracerouteHop{
			Success:     hop.Success,
			Address:     hop.AddressString(),
			Host:        hop.Host,
			N:           hop.N,
			ElapsedTime: hop.ElapsedTime,
			TTL:         hop.TTL,
		})
	}

	return result, nil
}
