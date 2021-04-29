package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"github.com/urfave/cli/v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func initCLI() *cli.App {
	filesList = processDirAndGenerateMeta(".")

	app := cli.NewApp()
	app.Name = "manifest"
	app.Usage = "create manifest files in current directory"
	app.Version = "1.0.0"
	addCLICommands(app)

	app.Action = func(cnx *cli.Context) error {
		cli.ShowAppHelpAndExit(cnx, 0)
		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	cli.VersionFlag = &cli.BoolFlag{
		Name:  "print-version",
		Usage: "print version",
	}

	return app
}

func addCLICommands(app *cli.App) {
	app.Commands = []*cli.Command{
		{
			Name:      "init",
			Usage:     "initialize tool environment by create the .cxo folder",
			UsageText: "create the manifest folder .cxo in current directory",
			Action: func(cnx *cli.Context) error {
				err := createFolder(".cxo")
				if err == nil {
					fmt.Println("Create .cxo foler in current directory: ")
				}
				return err
			},
		},
		{
			Name:      "commit",
			Usage:     "commit all the files' metadatum into the .cxo file",
			UsageText: "commit all the metadata files into the .cxo folder",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "print-json",
					Value: false,
					Usage: "print files in the directory in json ",
				},
				&cli.BoolFlag{
					Name:  "meta",
					Value: false,
					Usage: "add metadata section in json",
				},
			},
			Action: func(cnx *cli.Context) error {
				metaFlag := false
				if cnx.Bool("meta") {
					if !cnx.Bool("print-json") {
						cli.ShowAppHelpAndExit(cnx, 0)
					}
					metaFlag = true
				}
				cxoPath := currentDir + "/.cxo/"
				if !isFolderExist(cxoPath) {
					fmt.Println("please use 'manifest init' command before 'manifest commit'")
					os.Exit(1)
				}

				cxoFile, err := createFolderAndFile("./.cxo/checkpoints/", manifestCXOFolder, ".cxo")
				if err != nil {
					return err
				}
				defer cxoFile.Close()

				// manifest .cxo file
				var manifestOuputBody ManifestOuputBody
				manifestOuputBody.ManifestBody = *getManifestBody(filesList)
				manifestOuputBody.ManifestHeader = *getManifestDirectoryHeader(&manifestOuputBody.ManifestBody)
				manifestOuputBody.FileList = *getFileList(filesList)

				serializedOuputBody := encoder.Serialize(manifestOuputBody)
				_, err = cxoFile.Write(serializedOuputBody)
				if err != nil {
					return err
				}
				if cnx.Bool("print-json") {
					printFilesInJson(filesList, &manifestOuputBody.ManifestHeader, metaFlag)
				}

				// manifest meta and temp files
				err = generateMetaAndTempFiles()
				if err != nil {
					return err
				}

				return nil
			},
		},
	}
}

func main() {

	currentDir = getCurrentDir()
	app := initCLI()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func processDirAndGenerateMeta(dir string) *FilesInfoList {
	var FilesAndDirectories FilesInfoList
	var directories []string
	var directoriesSize []int
	var files []string
	var filesSize []int
	var filesHash []HashVariable
	var tempFileHash HashVariable
	var filesMetaList ManifestDirectMetaList
	var ChunksList [][]ChunkHash
	var filesCreateDate []string

	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				if info.Name() == ".cxo" {
					return filepath.SkipDir
				}
				directories = append(directories, path)
				dirSize, err := getDirectorySize(path)
				if err != nil {
					return err
				}
				directoriesSize = append(directoriesSize, dirSize)
			} else if info.Name() != appName {
				files = append(files, path)
				filesSize = append(filesSize, int(info.Size()))
				filesCreateDate = append(filesCreateDate, timespecToDate(info.Sys().(*syscall.Stat_t).Ctim))
				tempFileHash = HashVariable{[]byte("base64,sha256"), []byte(hashFileAndEncoding(path))}
				filechunks, err := getFileChunks(path)
				if err != nil {
					return err
				}
				ChunksList = append(ChunksList, *filechunks)
				filesHash = append(filesHash, tempFileHash)

				filesMetaList = append(filesMetaList, getFileMeta(path))
			}

			return nil
		})
	if err != nil {
		log.Fatal(err)
	}

	FilesAndDirectories.directoryNames = directories
	FilesAndDirectories.fileNames = files
	FilesAndDirectories.fileSizes = filesSize
	FilesAndDirectories.diretorySizes = directoriesSize
	FilesAndDirectories.filesHashlist = filesHash
	FilesAndDirectories.filesMetaList = filesMetaList
	FilesAndDirectories.filesChunksList = ChunksList
	FilesAndDirectories.filesCreationDateList = filesCreateDate
	return &FilesAndDirectories
}

func getFileChunks(filepath string) (*[]ChunkHash, error) {

	var fileData []ChunkHash
	var size uint64

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bf := make([]byte, chunkSize)
	hs := sha256.New()

	for {
		size = chunkSize
		readTotal, err := file.Read(bf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if readTotal == 0 {
			break
		}

		for readTotal < chunkSize {
			size = uint64(readTotal)
			readTotal = readTotal + copy(bf[readTotal:], []byte{0x0000})
		}

		hs.Write(bf)
		hash := hs.Sum(nil)
		hs.Reset()
		fileData = append(fileData, ChunkHash{size, hash})
	}

	return &fileData, nil
}

func printFilesInJson(fList *FilesInfoList, dirHeader *ManifestDirectoryHeader, metaflag bool) {
	var dirmeta DirectoryMetaList
	var filemeta FileDataList

	for indx, fn := range (*fList).fileNames {
		fh := (*fList).filesHashlist[indx].Hash
		fs := (*fList).fileSizes[indx]
		meta := (*fList).filesMetaList[indx]
		fileInfo := FileData{fn, fs, fh, &meta}
		if !metaflag {
			fileInfo = FileData{fn, fs, fh, nil}
		}

		filemeta = append(filemeta, fileInfo)

	}

	for indx, dn := range (*fList).directoryNames {
		ds := (*fList).diretorySizes[indx]
		dirInfo := DirectoryMeta{dn, ds}
		dirmeta = append(dirmeta, dirInfo)
	}

	sort.Sort(filemeta)
	sort.Sort(dirmeta)
	metadata := struct {
		DirectoryHeader ManifestDirectoryHeader `json:"directory header"`
		Directories     []DirectoryMeta         `json:"directories"`
		Files           []FileData              `json:"files"`
	}{*dirHeader, dirmeta, filemeta}

	jsons, err := json.MarshalIndent(metadata, "", "   ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsons))
}

func getManifestBody(fList *FilesInfoList) *ManifestDirectoryBody {

	var result ManifestDirectoryBody
	var fileHashList FileHashList

	for indx, fname := range (*fList).fileNames {
		fsize := (*fList).fileSizes[indx]
		fhash := (*fList).filesHashlist[indx]
		fileChunks := (*fList).filesChunksList[indx]
		for _, chunk := range fileChunks {
			fileHashList.ChunksHashes = append(fileHashList.ChunksHashes, chunk.Hash)
		}
		fileHashList.FileHash = fhash
		fullname := currentDir + "/" + fname
		paths, fileName := filepath.Split(fullname)
		manifestFile := ManifestFile{
			Path:       []byte(paths),
			FileName:   []byte(fileName),
			Size:       int64(fsize),
			HashList:   fileHashList,
			MetaString: []byte{},
		}
		fileHashList.ChunksHashes = nil
		result.ManifestFileList = append(result.ManifestFileList, manifestFile)
	}

	for indx, dirname := range (*fList).directoryNames {
		dirsize := (*fList).diretorySizes[indx]
		fullDirname := currentDir + "/" + dirname
		manifestFile := ManifestFile{
			Path:       []byte(fullDirname),
			FileName:   nil,
			Size:       int64(dirsize),
			HashList:   FileHashList{},
			MetaString: []byte{},
		}
		result.ManifestFileList = append(result.ManifestFileList, manifestFile)
	}

	return &result
}

func getManifestDirectoryHeader(body *ManifestDirectoryBody) *ManifestDirectoryHeader {
	var result ManifestDirectoryHeader
	dataSize := 0

	for _, manifile := range (*body).ManifestFileList {
		if manifile.FileName != nil {
			dataSize += int(manifile.Size)
		}
	}
	segLenth := unsafe.Sizeof((*body).ManifestFileList)
	version := []byte(versionNo)
	sequenceid := getSequenceId()
	createat := uint64(time.Now().Unix())
	bodySegmentLength := uint64(segLenth)
	bodyDataFileSize := uint64(dataSize)

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	result = ManifestDirectoryHeader{
		VersionString:     version,
		SequenceId:        sequenceid,
		Creator:           user.Name,
		CreatedAt:         createat,
		BodySegmentLength: bodySegmentLength,
		BodyDataFileSize:  bodyDataFileSize,
		MetaDataTags:      KeysValuesList{},
		ChunkSize:         chunkSize,
	}

	manifestMeta.ManifestHeaderMeta = *getManifestHeaderMetaData(&result)
	return &result
}

func getManifestHeaderMetaData(header *ManifestDirectoryHeader) *ManifestHeaderMetaData {
	var result ManifestHeaderMetaData

	creationTime := (*header).CreatedAt

	filename, err := getPreviousManifest((*header).SequenceId)
	if err != nil {
		panic(err)
	}
	previousManifest := filename

	serializedheader := encoder.Serialize(*header)
	h := sha256.New()
	h.Write(serializedheader)
	id := base64.StdEncoding.EncodeToString(h.Sum(nil))

	result = ManifestHeaderMetaData{
		CreationTime:     creationTime,
		Creator:          (*header).Creator,
		PreviousManifest: previousManifest,
		SequenceId:       (*header).SequenceId,
		UniqueId:         id,
	}
	return &result
}

func getPreviousManifest(currentSequenctId uint64) (string, error) { // ToDo
	// var manifestOuputBody ManifestOuputBody
	// cxoFolderName := currentDir + "/.cxo/checkpoints/"
	// var filename string
	// files, _ := ioutil.ReadDir(cxoFolderName)
	// for _, file := range files {
	// 	filename = file.Name()
	// 	if strings.HasSuffix(filename, ".cxo") {
	// 		cxoFile, err := os.Open(cxoFolderName + filename)
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 		defer cxoFile.Close()
	// 		fileBytes, err := ioutil.ReadAll(cxoFile)
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 		_, err = encoder.DeserializeRaw(fileBytes, &manifestOuputBody)
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 		if currentSequenctId == manifestOuputBody.ManifestHeader.SequenceId+1 {
	// 			break
	// 		}
	// 	}
	// }
	return "previousManifest", nil
}

func getSequenceId() uint64 {
	cxoFolderName := currentDir + "/.cxo/checkpoints/"
	files, _ := ioutil.ReadDir(cxoFolderName)
	count := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".cxo") {
			count++
		}
	}
	return uint64(count)
}

func getFileList(fList *FilesInfoList) *FileList {
	var result FileList

	fileList := getFileItemList(fList)
	result.FileItemList = *fileList
	listHeader := getFileListHeader(fList)
	result.Header = *listHeader
	return &result
}

func getFileListHeader(fList *FilesInfoList) *FileListHeader {
	var result FileListHeader
	var fileListRef []FileItemRef
	var tempFileRef FileItemRef
	var fileChunkHashList []FileChunksHash
	var tempFileChunkHash FileChunksHash
	var fileFullName string
	var fileSize uint64

	for indx, fileHash := range (*fList).filesHashlist {
		fileFullName = (*fList).fileNames[indx]
		path, fileName := filepath.Split(fileFullName)
		tempFileRef.Name = fileName
		fileSize = uint64((*fList).fileSizes[indx])
		tempFileRef.Size = fileSize
		tempFileRef.Path = path
		tempFileRef.Hash = fileHash.Hash
		fileListRef = append(fileListRef, tempFileRef)

		tempFileChunkHash.FileHash = fileHash.Hash
		tempFileChunkHash.FileSize = fileSize
		filech := (*fList).filesChunksList[indx]
		tempFileChunkHash.ChunksHashList = append(tempFileChunkHash.ChunksHashList, filech...)
		fileChunkHashList = append(fileChunkHashList, tempFileChunkHash)
	}

	result.FileListRef = fileListRef
	result.FileChunkHashList = fileChunkHashList
	result.ChunkHashSetList = *getChunkHashSetList(fList)
	return &result
}

func getFileItemList(fList *FilesInfoList) *[]FileItem {
	var result []FileItem
	var tempFileItem FileItem
	var tempFileHeader FileItemHeader

	for indx, fileHash := range (*fList).filesHashlist {
		tempFileItem.ChunksHashList = (*fList).filesChunksList[indx]
		tempFileHeader.Id = fileHash.Hash
		tempFileHeader.SequenceId = getSequenceId()
		tempFileHeader.CreationDate = (*fList).filesCreationDateList[indx]
		tempFileHeader.Size = chunkSize
		tempFileHeader.MetaDatum = KeysValuesList{}
		tempFileItem.Header = tempFileHeader
		result = append(result, tempFileItem)
	}

	return &result
}

func getChunkHashSetList(fList *FilesInfoList) *ChunkHashSetList {
	var result ChunkHashSetList
	var chunkSet ChunkHashSet
	var hashSet [][]byte
	var setMetaList []ChunkHashSetMeta

	result.Id = sha256.New().Sum(nil)
	chunkSet.Id = sha256.New().Sum(nil)
	for _, fileHash := range (*fList).filesChunksList {
		chunkSet.ChunkHashList = append(chunkSet.ChunkHashList, fileHash...)
		chunkSet.Count = chunkSet.Count + int64(len(fileHash))
		for _, ch := range fileHash {
			chunkSet.Size = chunkSet.Size + int64(ch.Size)
			hashSet = append(hashSet, ch.Hash)
		}
	}

	setMetaList = append(setMetaList, *getChunkHashSetMeta(&chunkSet))
	manifestMeta.ChunkHashSetMetaList = setMetaList

	byteArr := encoder.Serialize(chunkSet)
	h := sha256.New()
	h.Write(byteArr)
	chunkSet.Id = h.Sum(nil)

	SortByteArrays(hashSet)

	countBytes := encoder.Serialize(chunkSet.Count)
	countBytes = append(countBytes, encoder.Serialize(hashSet)...)

	h3 := sha256.New()
	h3.Write(countBytes)
	manifestTemp.HashSet.Id = h3.Sum(nil)
	manifestTemp.HashSet.HashSet = hashSet

	result.HashSetList = append(result.HashSetList, chunkSet)

	byteArr2 := encoder.Serialize(result)
	h2 := sha256.New()
	h2.Write(byteArr2)
	result.Id = h2.Sum(nil)

	manifestMeta.ChunkHashSetListMeta = *getChunkHashSetListMeta(&result, &setMetaList)
	return &result
}

func getChunkHashSetListMeta(setList *ChunkHashSetList, setMeraList *[]ChunkHashSetMeta) *ChunkHashSetListMeta {
	var result ChunkHashSetListMeta
	result.ListId = (*setList).Id

	for i, set := range (*setList).HashSetList {
		result.ChunkSetIdList = append(result.ChunkSetIdList, set.Id)
		result.ChunkSetSizeList = append(result.ChunkSetSizeList, (*setMeraList)[i].ChunkSetDataSize)
		result.ChunkSetDataSizeList = append(result.ChunkSetDataSizeList, set.Size)
		result.HashCountTotal = result.HashCountTotal + set.Count
	}

	return &result
}

func getChunkHashSetMeta(chunkSet *ChunkHashSet) *ChunkHashSetMeta {
	var result ChunkHashSetMeta

	result.Count = (*chunkSet).Count
	result.Id = (*chunkSet).Id
	result.ChunkSetDataSize = (*chunkSet).Size
	sz := unsafe.Sizeof(*chunkSet)
	result.ChunkSetSize = int64(sz)

	return &result
}

func generateMetaAndTempFiles() error {
	metaFile, err := createFolderAndFile("./.cxo/meta/", manifestMetaFolder, ".meta")
	if err != nil {
		return err
	}
	defer metaFile.Close()
	serializedMetaBody := encoder.Serialize(manifestMeta)
	_, err = metaFile.Write(serializedMetaBody)
	if err != nil {
		return err
	}

	tempFile, err := createFolderAndFile("./.cxo/temp/", manifestTempFolder, ".temp")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	serializedTempBody := encoder.Serialize(manifestTemp)
	_, err = tempFile.Write(serializedTempBody)
	if err != nil {
		return err
	}
	return nil
}
