package cmd

import (
	"log"
	"net"
)

type NetworkInfo struct {
	PublicIPv4 string   `json:"public_ipv4"`
	PublicIPv6 string   `json:"public_ipv6"`
	LocalIPv4  string   `json:"local_ipv4"`
	LocalIPv6  string   `json:"local_ipv6"`
	Addresses  []string `json:"addresses"`
}

func GetNetworkInfo() (NetworkInfo, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	ni := NetworkInfo{PublicIPv4: localAddr.IP.String()}
	return ni, err
}
