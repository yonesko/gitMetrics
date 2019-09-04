package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var rootDir = flag.String("dir", ".", "root dir")

func main() {
	flag.Parse()
	result := map[string]uint64{}
	started := time.Now()
	fmt.Printf("Started: %v\n", started.Format("15:04:05"))
	group := &sync.WaitGroup{}
	group.Add(1)
	go sloc(*rootDir, result, group)
	group.Wait()
	fmt.Printf("Finished: %v %v\n", time.Now().Format("15:04:05"), time.Since(started))
	printReport(result)
}

func printReport(result map[string]uint64) {
	//TODO sort by value
	fmt.Println("Result:")
	for k, v := range result {
		fmt.Printf("%v %v\n", k, v)
	}
}

//add lines to result for regular files
//and run recursively for dirs
func sloc(dirname string, result map[string]uint64, group *sync.WaitGroup) {
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
			sloc(path, result, group)
		} else {
			//TODO read parallel
			regFile, err := os.Open(path)
			if err != nil {
				log.Println("Can't open " + err.Error() + " " + path)
				continue
			}
			defer func() { err = regFile.Close() }()
			counter, err := lineCounter(regFile)
			if err != nil {
				log.Println("Can't lineCounter " + err.Error() + " " + path)
				continue
			}
			result[extractExtension(fileInfo.Name())] += uint64(counter)
		}
	}

}

func extractExtension(fileName string) string {
	if index := strings.LastIndexByte(fileName, '.'); index >= 0 {
		return fileName[index:]
	}
	return "without-extension"
}

func filesInDir(dirname string) (infos []os.FileInfo, err error) {
	file, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer func() { err = file.Close() }()
	infos, err = file.Readdir(0)
	return
}

//TODO test buff size
func lineCounter(r io.Reader) (int, error) {
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
