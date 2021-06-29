package nettest

import (
	"fmt"
	"io/fs"
	"os"
	"time"
)

var (
	tmpPath = "/tmp"
)

func getTmpDir(prefix string) (string, func(), error) {
	var dirName string
	for {
		dirName = fmt.Sprintf("%s/%s_%d", prefix, tmpPath, time.Now().UnixNano())
		if _, err := os.Stat(dirName); os.IsNotExist(err) {
			if err := os.Mkdir(dirName, fs.ModeDir); err != nil {
				return "", nil, err
			}
			return dirName, func() { removeTmpDir(dirName) }, nil
		}
	}
}

func removeTmpDir(name string) {
	_ = os.RemoveAll(name)
}
