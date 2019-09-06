package main

import (
	"bytes"
	"fmt"
	"gitmetrics/sort"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

const openFilesLimit = 1024

var rootDir string

func init() {
	if len(os.Args) == 2 {
		rootDir = os.Args[1]
	} else {
		fmt.Println("Usage: gitmetrics dir")
		os.Exit(0)
	}
}

func main() {
	result := map[string]uint64{}
	started := time.Now()
	group := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	openFilesLimiter := make(chan int, openFilesLimit)
	group.Add(1)
	go handleDir(rootDir, result, group, mutex, openFilesLimiter)
	group.Wait()
	printReport(result, started)
}

func printReport(result map[string]uint64, started time.Time) {
	fmt.Printf("Elapsed %v\n", time.Since(started))
	fmt.Println("Lines of code by extension:")
	pairs := sort.SortMapByValue(result)
	for _, pair := range pairs[:uint64(math.Min(15, float64(len(pairs))))] {
		fmt.Printf("%v %v\n", pair.Key, pair.Val)
	}
	if len(pairs) > 15 {
		fmt.Printf("and %v more\n", len(pairs)-15)
	}
}

//add lines to result for regular files
//and run recursively for dirs
func handleDir(dirname string, result map[string]uint64, group *sync.WaitGroup, mutex *sync.Mutex, openFilesLimiter chan int) {
	defer func() { group.Done() }()
	fileInfos, err := filesInDir(dirname, openFilesLimiter)
	if err != nil {
		log.Println("can't open " + err.Error() + " " + dirname)
		return
	}

	for _, fileInfo := range fileInfos {
		path := dirname + "/" + fileInfo.Name()
		if fileInfo.IsDir() {
			group.Add(1)
			go handleDir(path, result, group, mutex, openFilesLimiter)
		} else {
			regFile, err := openOrWait(path, openFilesLimiter)
			if err != nil {
				log.Println("Can't open " + err.Error() + " " + path)
				continue
			}
			defer func() { err = regFile.Close() }()
			count, err := countLines(regFile)
			if err != nil {
				log.Println("Can't countLines " + err.Error() + " " + path)
				continue
			}
			mutex.Lock()
			result[extractExtension(fileInfo.Name())] += uint64(count)
			mutex.Unlock()
		}
	}

}

func openOrWait(path string, openFilesLimiter chan int) (*os.File, error) {
	openFilesLimiter <- 1
	defer func() { <-openFilesLimiter }()
	return os.Open(path)
}

func extractExtension(fileName string) string {
	if index := strings.LastIndexByte(fileName, '.'); index >= 0 {
		return fileName[index:]
	}
	return "without-extension"
}

func filesInDir(dirname string, openFilesLimiter chan int) (infos []os.FileInfo, err error) {
	file, err := openOrWait(dirname, openFilesLimiter)
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
