package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

var workers = 4
var workerDirs = [][]string{}

type Mime struct {
	Count      int
	Size       int64
	Extensions map[string]int
}

func containsMime(mime string, mimes map[string]Mime) bool {

	for k, _ := range mimes {
		if k == mime {
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

func printMimes(mimes map[string]Mime) {
	outfile, _ := os.Create("output.tsv")
	defer outfile.Close()
	writer := bufio.NewWriter(outfile)
	writer.WriteString("mime\tcount\tsize\textensions\tpercent\n")

	var sumSize float64 = 0.0
	for _, v := range mimes {
		sumSize = sumSize + float64(v.Size)
	}

	for k, v := range mimes {
		pct := (float64(v.Size) * 100.00) / sumSize
		count := 0
		for _, i := range v.Extensions {
			count = count + i
		}
		fmt.Printf("%s\t%d\t%d\t%s\t\"%.2f\"\n", k, count, v.Size, mapToString(v.Extensions), pct)
		writer.WriteString(fmt.Sprintf("%s\t%d\t%d\t%s\t\"%.2f\"\n", k, count, v.Size, mapToString(v.Extensions), pct))
		writer.Flush()
	}
}

func main() {

	root := os.Args[1]
	subdirs, err := os.ReadDir(root)
	if err != nil {
		panic(err)
	}

	dirsToScan := []string{}

	for _, d := range subdirs {
		if d.IsDir() {
			dirsToScan = append(dirsToScan, filepath.Join(root, d.Name()))
		}
	}

	workerDirs = chunkDirs(dirsToScan)
	scan()
}
