package golang

import (
	"os/exec"
)

type Result struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Error   error  `json:"error"`
}

func Run() Result {
	// Lookup for command `go`.
	path, err := exec.LookPath("go")
	if err != nil {
		return Result{Error: err}
	}

	// Execute `go version` command and get Result.
	b, err := exec.Command("go", "version").Output()
	if err != nil {
		return Result{Error: err, Path: path}
	}

	return Result{Error: nil, Path: path, Version: string(b)}
}
