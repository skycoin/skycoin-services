package netinfo

import (
	"io"
	"net"
	"net/http"
)

type HttpTest struct {
	Port int `json:"port"`
}

// Get returns network information.
func Get() NetworkInfo {
	n := NetworkInfo{}
	n.IPv4Method, n.IPv6Method = "https://v4.ident.me/", "https://v6.ident.me/"
	n.IPv4, n.IPv6 = getIP(n.IPv4Method), getIP(n.IPv6Method)

	// Grab all network interfaces.
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		var addresses []string

		// Grab all iface addresses.
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			addresses = append(addresses, addr.String())
		}

		n.Ifaces = append(n.Ifaces, iface{
			MTU:          i.MTU,
			Name:         i.Name,
			PhysicalAddr: i.HardwareAddr.String(),
			Addresses:    addresses,
		})
	}

	return n
}

// NetworkInfo information about network.
type NetworkInfo struct {
	IPv4       string  `json:"ipv4"`
	IPv4Method string  `json:"ipv4_method"`
	IPv6       string  `json:"ipv6"`
	IPv6Method string  `json:"ipv6_method"`
	Ifaces     []iface `json:"ifaces"`
}

type iface struct {
	MTU          int      `json:"mtu"`
	Name         string   `json:"name"`
	PhysicalAddr string   `json:"physical_addr"`
	Addresses    []string `json:"addresses"`
}

// getIP make GET request to the url, which must return public IP address of the current machine.
func getIP(url string) string {
	res, err := http.Get(url)
	if err != nil {
		return ""
	}

	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return string(b)
}
