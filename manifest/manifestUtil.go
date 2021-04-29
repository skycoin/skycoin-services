package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"time"
)

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}

func getFileMeta(filename string) FileMeta {

	var result FileMeta

	fileInfo, _ := os.Stat(filename)
	fStat := fileInfo.Sys().(*syscall.Stat_t)
	result.LastModified = uint64(fileInfo.ModTime().Unix())
	result.UnixPermission = fileInfo.Mode().String()
	sec, _ := fStat.Ctim.Unix()
	result.CreateAt = uint64(sec)

	return result
}

func createFolderAndFile(baseFolder string, subFolder string, fileExt string) (*os.File, error) {
	err := os.MkdirAll(baseFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}
	FileName := currentDir + subFolder + strconv.FormatInt(time.Now().Unix(), 10) + fileExt

	File, err := os.OpenFile(FileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return File, nil
}

func getDirectorySize(directory string) (int, error) {
	totalSize := 0
	err := filepath.Walk(directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				totalSize = +int(info.Size())

			}
			return nil
		})
	return totalSize, err
}

func createFolder(folderName string) error {
	var err error
	if isFolderExist(folderName) {
		return nil
	}
	err = os.Mkdir(folderName, 0777)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chmod(folderName, 0777)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func hashFileAndEncoding(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func isFolderExist(path string) bool {

	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

func SortByteArrays(src [][]byte) {
	sort.Slice(src, func(i, j int) bool { return bytes.Compare(src[i], src[j]) < 0 })
}

func timespecToDate(ts syscall.Timespec) string {
	res := time.Unix(int64(ts.Sec), int64(ts.Nsec)).Format("2006-01-02")
	return res
}
