package main

import (
	"context"
	"fmt"
	"github.com/dmnyu/go-tika/tika"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ScanResult struct {
	WorkerID int
	Mimes    map[string]Mime
}

func scan() {
	resultChannel := make(chan ScanResult)

	for i, chunk := range workerDirs {
		go scanDirectory(chunk, resultChannel, i+1)
	}

	results := []ScanResult{}
	for range workerDirs {
		resultChunk := <-resultChannel
		results = append(results, resultChunk)
	}

	printMimes(mergeScanReults(results))
}

func mergeScanReults(scanResults []ScanResult) map[string]Mime {
	mergedResults := map[string]Mime{}
	for _, scanResult := range scanResults {
		for mime, info := range scanResult.Mimes {
			if containsMime(mime, mergedResults) {
				m := mergedResults[mime]
				m.Size = m.Size + info.Size
				exts := m.Extensions
				for e, i := range info.Extensions {
					if containsExtension(e, exts) {
						exts[e] = exts[e] + i
					} else {
						exts[e] = i
					}
				}
				mergedResults[mime] = m
			} else {
				mergedResults[mime] = info
			}
		}
	}
	return mergedResults
}

func scanDirectory(dirs []string, resultChan chan ScanResult, workerID int) {
	var tikaServer *tika.Server
	var tikaClient *tika.Client
	var port = strconv.Itoa(9998 - workerID)
	tikaServer, err := tika.NewServer("tika-server-1.28.5.jar", port)
	if err != nil {
		log.Fatal(err)
	}
	err = tikaServer.Start(context.Background())

	if err != nil {
		log.Fatal(err)
	}
	defer tikaServer.Stop()
	fmt.Printf("Worker %d started Tika Server %s\n", workerID, tikaServer.URL())

	tikaClient = tika.NewClient(nil, tikaServer.URL())
	dirMimes := map[string]Mime{}
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() {
				//fmt.Println("Scanning", path)
				ext := strings.ToLower(filepath.Ext(info.Name()))
				reader, err := os.Open(path)
				mime, err := tikaClient.Detect(context.Background(), reader)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Worker %d detected file \"%s\" as mime %s\n", workerID, path, mime)
				if containsMime(mime, dirMimes) {
					m := dirMimes[mime]
					m.Size = m.Size + info.Size()
					exts := m.Extensions
					if containsExtension(ext, exts) {
						m.Extensions[ext] = m.Extensions[ext] + 1
					} else {
						m.Extensions[ext] = 1
					}
				} else {

					if err != nil {
						panic(err)
					}
					dirMimes[mime] = Mime{0, info.Size(), map[string]int{ext: 1}}
				}
			} else {
				fmt.Printf("Worker %d scanning %s\n", workerID, path)
			}
			return nil
		})
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	fmt.Printf("Worker %d complete\n", workerID)
	resultChan <- ScanResult{
		WorkerID: workerID,
		Mimes:    dirMimes,
	}
}

func chunkDirs(dirs []string) [][]string {
	var divided [][]string
	chunkSize := (len(dirs) + workers - 1) / workers

	for i := 0; i < len(dirs); i += chunkSize {
		end := i + chunkSize

		if end > len(dirs) {
			end = len(dirs)
		}

		divided = append(divided, dirs[i:end])
	}
	return divided
}
