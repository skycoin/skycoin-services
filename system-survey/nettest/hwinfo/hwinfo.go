package hwinfo

import (
	"github.com/jaypipes/ghw"
)

type Result struct {
	Info  *ghw.HostInfo `json:"info"`
	Error error         `json:"error"`
}

func Run() Result {
	res := Result{}
	res.Info, res.Error = ghw.Host()
	return res
}
