package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

var (
	pathVarName = "PATH"
	pathVarSep  = ":"
)

type AppList map[string][]string

// GetAppList returns new loaded AppList.
func GetAppList(filter string) AppList {
	apps := AppList{}
	for _, path := range strings.Split(os.Getenv(pathVarName), pathVarSep) {
		if path == "" {
			path = "undefined_location"
		}
		apps[path] = getDirEntries(path, filter)
	}
	return apps
}

// JSON renders AppList to the machine readable JSON format and return JSON bytes.
func (a AppList) JSON() []byte {
	b, _ := json.Marshal(a)
	return b
}

// String renders AppList to the human readable format and returns it.
func (a AppList) String() string {
	buf := new(bytes.Buffer)
	for path, entries := range a {
		_, _ = fmt.Fprintf(buf, "%s:\n", path)
		for _, entry := range entries {
			_, _ = fmt.Fprintf(buf, "\t- %s\n", entry)
		}
	}
	return buf.String()
}

// getDirEntries scans directory with location path and returns it's entries.
func getDirEntries(path, filter string) []string {
	// Get all files located in the path directory
	files, err := os.ReadDir(path)

	// Load application filter.
	neededApps := make(map[string]bool)
	if filter != "" {
		for _, app := range strings.Split(filter, ",") {
			neededApps[app] = true
		}
	}
	// Compute applyFilter once before multiple use.
	applyFilter := len(neededApps) > 0

	entriesMu := &sync.Mutex{}
	var entries []string
	// If you get an error while trying to get files, then scan will be skipped
	// and ad nil array will be returned.
	if err == nil {
		wg := &sync.WaitGroup{}
		for _, file := range files {
			file := file
			wg.Add(1)
			go func() {
				defer wg.Done()
				fileName := file.Name()
				if applyFilter {
					if _, ok := neededApps[fileName]; !ok {
						return
					}
				}
				entriesMu.Lock()
				entries = append(entries, fileName)
				entriesMu.Unlock()
			}()
		}
		wg.Wait()
	}
	sort.Strings(entries)
	return entries
}
