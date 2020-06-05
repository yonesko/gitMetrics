package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yonesko/gitmetrics/util"
)

var cpuProf = flag.String("cpup", "", "cpu profile file")
var openFilesLimiter = make(chan int, 1024)
var group = &sync.WaitGroup{}
var mutex = &sync.Mutex{}
var filesProcessed uint64
var extSloc = map[string]uint64{}
var extCount = map[string]uint64{}

func main() {
	flag.Parse()
	rootDir := rootDir()
	var started = time.Now()
	if startCPUProfile() {
		defer pprof.StopCPUProfile()
	}
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

func rootDir() string {
	if flag.Arg(0) == "" {
		return "."
	}
	return flag.Arg(0)
}

func startCPUProfile() bool {
	if *cpuProf == "" {
		return false
	}
	file, err := os.Create(*cpuProf)
	if err != nil {
		println("Can't StartCPUProfile: " + err.Error())
		return false
	}
	err = pprof.StartCPUProfile(file)
	if err != nil {
		println("Can't StartCPUProfile: " + err.Error())
		return false
	}
	return true
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
		time.Sleep(333 * time.Millisecond)
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

var countLinesBuffer = make([]byte, 1024*1024)

func countLines(r io.Reader) (int, error) {
	count := 0
	for {
		c, err := r.Read(countLinesBuffer)
		count += bytes.Count(countLinesBuffer[:c], []byte{'\n'})

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
