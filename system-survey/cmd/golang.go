package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var getGoVersionFunc = getGoVersion

type golangVersion struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// GetGolangVersion returns new o
func GetGolangVersion() (*golangVersion, error) {
	out, err := getGoVersionFunc()
	if err != nil {
		return nil, err
	}
	gv := new(golangVersion)
	gv.fromRaw(out)
	return gv, err
}

// String returns golangVersion in the human readable format.
func (g golangVersion) String() string {
	return fmt.Sprintf("version=%s, os=%s, arch=%s", g.Version, g.OS, g.Arch)
}

// JSON returns golangVersion in the machine readable JSON format.
func (g golangVersion) JSON() []byte {
	b, _ := json.Marshal(g)
	return b
}

// fromRaw parses s string and split output of `go version` command into necessary parts.
func (g *golangVersion) fromRaw(s string) {
	// Determine golang version
	parts := strings.Split(s[13:], " ")
	g.Version = parts[0]

	// Determine os and arch.
	parts = strings.Split(parts[1], "/")
	g.OS, g.Arch = parts[0], regexp.MustCompile(`\r?\n`).ReplaceAllString(parts[1], "")
}

// getGoVersion executes `go version` command in the system and returns it's output.
func getGoVersion() (string, error) {
	b, err := exec.Command("go", "version").Output()
	return string(b), err
}
