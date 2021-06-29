package cmd

import (
	"github.com/matishsiao/goInfo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetHwinfo(t *testing.T) {
	gi := goInfo.GetInfo()
	hi := GetHwinfo()

	assert.Equal(t, gi.GoOS, hi.OSInfo.Goos)
	assert.Equal(t, gi.Kernel, hi.OSInfo.Kernel)
	assert.Equal(t, gi.Core, hi.OSInfo.Core)
	assert.Equal(t, gi.Platform, hi.OSInfo.Platform)
	assert.Equal(t, gi.Hostname, hi.OSInfo.Hostname)
	assert.Equal(t, gi.CPUs, hi.OSInfo.CPUs)
}
