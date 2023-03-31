package main

import (
	"context"
	"fmt"
	"github.com/google/go-tika/tika"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Ext struct {
	Count int
	Size  int64
	Mime  string
}

var extensions = map[string]Ext{}

func contains(ext string) bool {
	for k, _ := range extensions {
		if k == ext {
			return true
		}
	}
	return false
}

var tikaServer *tika.Server
var tikaClient *tika.Client

func init() {
	var err error
	tikaServer, err = tika.NewServer("tika-server-1.28.5.jar", "")
	if err != nil {
		log.Fatal(err)
	}
	err = tikaServer.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer tikaServer.Stop()

	tikaClient = tika.NewClient(nil, tikaServer.URL())

}

func main() {
	root := os.Args[1]

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if contains(ext) {
				e := extensions[ext]
				e.Count = e.Count + 1
				e.Size = e.Size + info.Size()
				extensions[ext] = e
			} else {
				reader, err := os.Open(path)
				if err != nil {
					panic(err)
				}
				mime, err := tikaClient.Detect(context.Background(), reader)
				if err != nil {
					panic(err)
				}

				extensions[ext] = Ext{1, info.Size(), mime}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	var sumSize int64 = 0

	for _, v := range extensions {
		sumSize = sumSize + v.Size
	}

	for k, v := range extensions {
		pct := (v.Size * 100) / sumSize
		fmt.Println(k, v.Count, v.Size, v.Mime, pct)
	}
}
