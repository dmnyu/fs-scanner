package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/dmnyu/go-tika/tika"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Mime struct {
	Count      int
	Size       int64
	Extensions map[string]int
}

var mimes = map[string]Mime{}

func containsMime(ext string) bool {
	for k, _ := range mimes {
		if k == ext {
			return true
		}
	}
	return false
}

func containsExtension(ext string, exts map[string]int) bool {
	for e, _ := range exts {
		if e == ext {
			return true
		}
	}
	return false
}

func mapToString(m map[string]int) string {
	out := ""
	for k, v := range m {
		if out == "" {
			out = fmt.Sprintf("%s:%d ", k, v)
		} else {
			out = fmt.Sprintf("%s, %s:%d ", out, k, v)
		}
	}
	return out[0 : len(out)-1]
}

var tikaServer *tika.Server
var tikaClient *tika.Client

func main() {
	fmt.Println("START")
	root := os.Args[1]

	fmt.Println("DOWNLOAD")
	err := tika.DownloadServer(context.Background(), tika.Version1285, "tika-server-1.28.5.jar")
	if err != nil {
		panic(err)
	}

	fmt.Println("INIT")
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

	fmt.Println("SCAN")
	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			reader, err := os.Open(path)
			mime, err := tikaClient.Detect(context.Background(), reader)
			if err != nil {
				panic(err)
			}

			if containsMime(mime) {
				m := mimes[mime]
				m.Count = m.Count + 1
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
				mimes[mime] = Mime{1, info.Size(), map[string]int{ext: 1}}
			}
		} else {
			fmt.Println("Scanning", path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	var sumSize float64 = 0.0

	for _, v := range mimes {
		sumSize = sumSize + float64(v.Size)
	}

	outfile, _ := os.Create("output.tsv")
	defer outfile.Close()
	writer := bufio.NewWriter(outfile)

	for k, v := range mimes {
		pct := (float64(v.Size) * 100.00) / sumSize
		writer.WriteString(fmt.Sprintf("%s\t%d\t%d\t%s\t\"%.2f\"\n", k, v.Count, v.Size, mapToString(v.Extensions), pct))
		writer.Flush()
	}
}
