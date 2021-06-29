package main

import (
	"encoding/json"
	torrentlog "github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var magnets = []string{
	"570B38F1BAF89B40903E6066E849B5BD07CC0879",
	"8D3C4677CF4F7F12A0DDFBFE00214C48C7712624",
	"311BCEE679211FF69DB471E2EFD6099A3F06C3D3",
	"602A60BB6B4F061B5E4E7BE9822B67D6CD289E23",
	"C1F4D522DF9D86CDF39B67DB46762C7801F2E646",
	"4E64AAAF48D922DBD93F8B9E4ACAA78C99BC1F40",
	"34B20C595E9D7A800C242DC2918B8B870AFDEC12",
	"135E2BE1F525D30AAC80E413F371AD171E81C68D",
}

func main() {
	tempDir := getTempDir()
	defer clearTempDir(tempDir)

	log.Println("Data dir:", tempDir)

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = tempDir
	cfg.Logger = torrentlog.Discard

	c, _ := torrent.NewClient(cfg)
	defer c.Close()

	mu, wg := sync.Mutex{}, sync.WaitGroup{}

	res := result{TorrentGrab: map[string]bool{}}
	for _, hash := range magnets {
		wg.Add(1)

		hash := hash
		go func() {
			defer wg.Done()

			var val bool

			t, _ := c.AddTorrentInfoHash(metainfo.NewHashFromHex(hash))
			select {
			case _, val = <-t.GotInfo():
				val = !val
			case <-time.After(time.Minute):
			}

			mu.Lock()
			res.TorrentGrab[hash] = val
			mu.Unlock()
		}()
	}

	wg.Wait()

	b, _ := json.Marshal(res)
	log.Printf("%s", b)
}

func clearTempDir(path string) {
	_ = os.RemoveAll(path)
}

func getTempDir() string {
	var dir string

	// Define chars.
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	for {
		b := make([]rune, 10)
		for i := range b {
			b[i] = chars[rand.Intn(len(chars))]
		}

		dir = filepath.Join(os.TempDir(), string(b))

		if _, err := os.Stat(dir); err == nil {
			continue
		} else if os.IsNotExist(err) {
			break
		} else {
			log.Fatal("Failed to create temp dir.")
		}
	}

	return dir
}
