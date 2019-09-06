package main

import (
	"bytes"
	"fmt"
	"gitmetrics/util"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var rootDir string
var openFilesLimiter = make(chan int, 1024)
var group = &sync.WaitGroup{}
var mutex = &sync.Mutex{}
var extSloc = map[string]uint64{}

//TODO state line
//TODO avoid links
func init() {
	if len(os.Args) == 2 {
		rootDir = os.Args[1]
	} else {
		fmt.Println("Usage: gitmetrics dir")
		os.Exit(0)
	}
}

func main() {
	var started = time.Now()
	group.Add(1)
	go handleDir(rootDir)
	group.Wait()
	printReport(started)
}

func printReport(started time.Time) {
	fmt.Printf("Elapsed %v\n", time.Since(started))
	fmt.Printf("%10v %10v\n", "ext", "sloc")
	fmt.Printf("%10v %10v\n", "---", "---")
	for _, pair := range util.SortMapByValue(extSloc) {
		fmt.Printf("%10v %10v\n", pair.Key, util.PrettyBig(pair.Val))
	}
}

//add lines to extSloc for regular files
//and run recursively for dirs
func handleDir(dirname string) {
	defer func() { group.Done() }()
	fileInfos, err := filesInDir(dirname)
	if err != nil {
		log.Println("can't open " + err.Error() + " " + dirname)
		return
	}

	for _, fileInfo := range fileInfos {
		path := dirname + "/" + fileInfo.Name()
		if fileInfo.IsDir() {
			group.Add(1)
			go handleDir(path)
		} else {
			regFile, err := openOrWait(path)
			if err != nil {
				log.Println("Can't open " + err.Error() + " " + path)
				continue
			}
			count, err := countLines(regFile)
			err = regFile.Close()
			if err != nil {
				log.Println("Can't countLines " + err.Error() + " " + path)
				continue
			}
			mutex.Lock()
			extSloc[extractExtension(fileInfo.Name())] += uint64(count)
			mutex.Unlock()
		}
	}

}

func openOrWait(path string) (*os.File, error) {
	openFilesLimiter <- 1
	defer func() { <-openFilesLimiter }()
	return os.Open(path)
}

func extractExtension(fileName string) string {
	if index := strings.LastIndexByte(fileName, '.'); index >= 0 {
		return fileName[index:]
	}
	return ""
}

func filesInDir(dirname string) (infos []os.FileInfo, err error) {
	file, err := openOrWait(dirname)
	if err != nil {
		return nil, err
	}
	defer func() { err = file.Close() }()
	infos, err = file.Readdir(0)
	return
}

//TODO test buff size
func countLines(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
