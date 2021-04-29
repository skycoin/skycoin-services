package main

import (
	"bufio"
	"bytes"
	crtRand "crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/skycoin/skycoin/src/cipher/encoder"
	"github.com/stretchr/testify/require"
)

const (
	// buffer size for reading file
	bufSize = 1024
	// write to file randomly with max number of byte
	randomFileSizeMax = 1024 * 1024 * 10
	// number of files generated into 3 directories
	numOfTestFiles = 15
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	currentDir = getCurrentDir()
	os.MkdirAll("./testdata", os.ModePerm)
	return func(t *testing.T) {
		os.RemoveAll("./testdata")
	}
}

func TestManifestWithGeneratedData(t *testing.T) {
	var funcs []func() error
	funcs = append(funcs, generateTestData1)
	funcs = append(funcs, generateTestData2)
	funcs = append(funcs, generateTestData3)
	funcs = append(funcs, generateTestData4)

	testNames := []string{"3 files", "3 directories", "random files", "3 levels of subfolder"}

	for i, fun := range funcs {
		t.Run(testNames[i], func(t *testing.T) {
			configTestCase := setupTestCase(t)
			defer configTestCase(t)

			err := fun()
			if err != nil {
				fmt.Printf("could not generate test files: %v", err)
				os.Exit(1)
			}

			makeManifestAndMove2TestFolder()

			changeDir("./testdata")
			defer changeDir("..")

			fList := processDirAndGenerateMeta(".")
			execManifestCmd()

			manifestFileDirList := getTestDataManifest()

			testManifestOutput(t, manifestFileDirList, fList)

		})
	}

}

// test for 3 files in testdata directory
func generateTestData1() error {
	var err error
	filePrefix := currentDir + "/testdata/test_level_0_"
	var testfileNames []string
	for i := 0; i < 3; i++ {
		testfileNames = append(testfileNames, fmt.Sprintf("%s%d", filePrefix, i))
	}

	err = writeFilesRandomly(testfileNames)

	return err
}

// test for 3 files in first level subfolder of testdata directory
func generateTestData2() error {
	var err error
	filePrefix := currentDir + "/testdata/test_level_0_"
	var testfileNames []string
	var testdirNames []string

	for i := 0; i < 3; i++ {
		testfileNames = append(testfileNames, fmt.Sprintf("%s%d", filePrefix, i))
		testdirNames = append(testdirNames, "./testdata/"+getRandomString())
		err = createFolder(testdirNames[i])
		if err != nil {
			return err
		}
	}

	for _, dir := range testdirNames {
		for i := 1; i <= numOfTestFiles; i++ {
			filename := fmt.Sprintf("%s%s%d", dir, "/test_level_1_", i)
			testfileNames = append(testfileNames, filename)
		}
	}

	err = writeFilesRandomly(testfileNames)

	return err
}

// test for random number of files in testdata directory
func generateTestData3() error {
	filePath := currentDir + "/testdata/"

	return getRandomFiles(filePath)
}

// test for 3 levels of subfolder of testdata directory
func generateTestData4() error {
	var err error
	var testdirNames []string

	testdirNames = append(testdirNames, "./testdata/"+getRandomString()+"/")
	for i := 0; i < 3; i++ {
		testdirNames = append(testdirNames, testdirNames[i]+getRandomString()+"/")
		err = createFolder(testdirNames[i])
		if err != nil {
			return err
		}
	}

	for i := 0; i < 3; i++ {
		err = getCertainNumOfFiles(testdirNames[i], i+1, i)
	}

	return err
}

func writeFilesRandomly(fileNames []string) error {
	var err error
	var testfiles []*os.File

	mathRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	numOfFiles := len(fileNames)

	for i := 0; i < numOfFiles; i++ {
		file, err := os.Create(fileNames[i])
		if err != nil {
			break
		}
		testfiles = append(testfiles, file)
		defer testfiles[i].Close()
	}

	for i := 0; i < numOfFiles; i++ {
		size := mathRand.Int63n(randomFileSizeMax)
		mathRand.Seed(time.Now().UnixNano())
		fb := bufio.NewWriter(testfiles[i])
		defer fb.Flush()
		buf := make([]byte, bufSize)
		for i := size; i > 0; i -= bufSize {
			if _, err = crtRand.Read(buf); err != nil {
				return err
			}
			bR := bytes.NewReader(buf)
			if _, err = io.Copy(fb, bR); err != nil {
				return err
			}
		}
	}

	return err
}

func getRandomFiles(dir string) error {
	var err error
	var testfileNames []string
	filePrefix := dir + "test_level_0_"

	rand.Seed(time.Now().UnixNano())
	numFiles := rand.Intn(21)
	for i := 0; i < numFiles; i++ {
		testfileNames = append(testfileNames, fmt.Sprintf("%s%d", filePrefix, i+1))
	}

	if len(testfileNames) > 0 {
		err = writeFilesRandomly(testfileNames)
	}

	return err
}

func getCertainNumOfFiles(dir string, num int, level int) error {
	var err error
	var testfileNames []string
	filePrefix := dir + "test_level_" + strconv.Itoa(level) + "_"

	for i := 0; i < num; i++ {
		testfileNames = append(testfileNames, fmt.Sprintf("%s%d", filePrefix, i+1))
	}

	if len(testfileNames) > 0 {
		err = writeFilesRandomly(testfileNames)
	}

	return err
}

func getRandomString() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(20) + 1
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func getTestDataManifest() *[]ManifestFile {
	var manifestOuputBody ManifestOuputBody
	cxoFileFolder := currentDir + "/testdata/.cxo/checkpoints/"

	files, _ := ioutil.ReadDir(cxoFileFolder)
	if len(files) != 1 {
		fmt.Println("multiple files exist in cxo folder: ")
		os.Exit(1)
	}
	filename := cxoFileFolder + files[0].Name()
	if !strings.HasSuffix(filename, ".cxo") {
		fmt.Println("'.cxo' file does not exist")
		os.Exit(1)
	}

	cxoFile, err := os.Open(filename)
	if err != nil {
		fmt.Println("failed to open manifest file")
		os.Exit(1)
	}
	defer cxoFile.Close()
	fileBytes, err := ioutil.ReadAll(cxoFile)
	if err != nil {
		fmt.Println("failed to read manifest files")
		os.Exit(1)
	}
	_, err = encoder.DeserializeRaw(fileBytes, &manifestOuputBody)
	if err != nil {
		fmt.Println("failed to deserialize manifest file")
		os.Exit(1)
	}

	return &manifestOuputBody.ManifestBody.ManifestFileList
}

func changeDir(path string) {
	err := os.Chdir(path)
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}
}

func execManifestCmd() {
	cmd := exec.Command("./manifest", "init")
	err := cmd.Run()
	if err != nil {
		fmt.Println("run 'manifest init' failed: ", err)
		os.Exit(1)
	}
	cmd = exec.Command("./manifest", "commit", "-print-json")
	err = cmd.Run()
	if err != nil {
		fmt.Println("run 'manifest commit' failed: ", err)
		os.Exit(1)
	}
}

func makeManifestAndMove2TestFolder() {
	make := exec.Command("make", "build")
	err := make.Run()
	if err != nil {
		fmt.Println("could not make binary for manifest: ", err)
		os.Exit(1)
	}

	err = os.Rename("./manifest", "./testdata/manifest")
	if err != nil {
		fmt.Println("could not move manifest: ", err)
		os.Exit(1)
	}
}

func testManifestOutput(t *testing.T, manifestFileDirList *[]ManifestFile, fList *FilesInfoList) {
	var manifestFileList []ManifestFile
	var manifestDirList []ManifestFile
	var outPutFileNames []string
	var outPutDirNames []string
	var outPutFileHashList []FileHashList
	var testFileHashList []FileHashList
	var chunkHash [][]byte
	for _, filedir := range *manifestFileDirList {
		if filedir.FileName != nil {
			manifestFileList = append(manifestFileList, filedir)
		} else {
			manifestDirList = append(manifestDirList, filedir)
		}
	}

	for _, file := range manifestFileList {
		outPutFileHashList = append(outPutFileHashList, file.HashList)
	}

	for _, file := range manifestFileList {
		outPutFileNames = append(outPutFileNames, string(file.FileName))
	}

	currdir := getCurrentDir()
	var dirNamesTest []string
	var fileNamesTest []string

	for i, fh := range (*fList).filesHashlist {
		file := (*fList).filesChunksList[i]
		for _, hs := range file {
			chunkHash = append(chunkHash, hs.Hash)
		}
		testFileHashList = append(testFileHashList, FileHashList{fh, chunkHash})
		chunkHash = nil
	}

	for _, filename := range (*fList).fileNames {
		_, fn := filepath.Split(filename)
		fileNamesTest = append(fileNamesTest, fn)
	}

	for _, dirname := range (*fList).directoryNames {
		fullpath := currdir + "/" + dirname
		dirNamesTest = append(dirNamesTest, fullpath)
	}

	for _, dirName := range manifestDirList {
		outPutDirNames = append(outPutDirNames, string(dirName.Path))
	}
	require.Equal(t, fileNamesTest, outPutFileNames, "The two file name list should have the same content.")
	require.Equal(t, dirNamesTest, outPutDirNames, "The two directory name list should have the same content.")
	require.Equal(t, testFileHashList, outPutFileHashList, "The two hash list should have the same content.")
}
