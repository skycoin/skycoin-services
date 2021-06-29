package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jaypipes/ghw"
	"github.com/matishsiao/goInfo"
)

type Hwinfo struct {
	*ghw.HostInfo
	OSInfo struct {
		Goos     string `json:"goos"`
		Kernel   string `json:"kernel"`
		Core     string `json:"core"`
		Platform string `json:"platform"`
		Hostname string `json:"hostname"`
		CPUs     int    `json:"cpus"`
	} `json:"os_info"`
	osInfo *goInfo.GoInfoObject
}

func (h *Hwinfo) JSON() []byte {
	b, _ := json.Marshal(h)
	return b
}

func (h *Hwinfo) String() string {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, "%s\n", h.osInfo.String())
	_, _ = fmt.Fprintf(buf, h.HostInfo.String())
	return buf.String()
}

func GetHwinfo() *Hwinfo {
	hi := &Hwinfo{}

	gi := goInfo.GetInfo()
	hi.osInfo = gi
	hi.OSInfo.Goos = gi.GoOS
	hi.OSInfo.Kernel = gi.Kernel
	hi.OSInfo.Core = gi.Core
	hi.OSInfo.Platform = gi.Platform
	hi.OSInfo.Hostname = gi.Hostname
	hi.OSInfo.CPUs = gi.CPUs
	hi.HostInfo, _ = ghw.Host(ghw.WithDisableWarnings())
	return hi
}
