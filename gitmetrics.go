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
	"sync/atomic"
	"time"
)

var rootDir string
var openFilesLimiter = make(chan int, 1024)
var group = &sync.WaitGroup{}
var mutex = &sync.Mutex{}
var filesProcessed uint64
var extSloc = map[string]uint64{}
var extCount = map[string]uint64{}

//TODO avoid links
func init() {
	switch len(os.Args) {
	case 2:
		rootDir = os.Args[1]
	case 1:
		rootDir = "."
	default:
		fmt.Println("Usage: gitmetrics <dir>")
		os.Exit(0)
	}
}

func main() {
	var started = time.Now()
	group.Add(1)
	go handleDir(rootDir)
	stop := false
	stopped := make(chan struct{})
	go printProcessingState(&stop, stopped)
	group.Wait()
	stop = true
	<-stopped
	printReport(started)
}

func printReport(started time.Time) {
	fmt.Printf("Elapsed %v\n", time.Since(started))
	fmt.Printf("%10v %10v %10v\n", "ext", "sloc", "count")
	fmt.Printf("%10v %10v %10v\n", "---", "---", "---")
	for _, pair := range util.SortMapByValue(extSloc) {
		fmt.Printf("%10v %10v %10v\n", pair.Key, util.AbrvInteger(extSloc[pair.Key]), extCount[pair.Key])
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
			extension := extractExtension(fileInfo.Name())
			extSloc[extension] += uint64(count)
			extCount[extension] += 1
			mutex.Unlock()
		}
	}

}

func openOrWait(path string) (*os.File, error) {
	openFilesLimiter <- 1
	defer func() { <-openFilesLimiter }()
	file, err := os.Open(path)
	if err == nil {
		atomic.AddUint64(&filesProcessed, 1)
	}
	return file, err
}

func printProcessingState(stop *bool, stopped chan struct{}) {
	for !*stop {
		fmt.Printf("Files processed %10v", filesProcessed)
		time.Sleep(333)
		for i := 0; i < 25; i++ {
			fmt.Print("\r")
		}
	}
	fmt.Println("")
	stopped <- struct{}{}
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
