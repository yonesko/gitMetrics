package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	rootDir := "/Users/glebio/go/src/gitMetrics"
	result := map[string]uint64{}
	sloc(rootDir, &result)
	fmt.Println(result)
}

//add lines to result for regular files
//and run recursively for dirs
func sloc(dirname string, result *map[string]uint64) {
	file, err := os.Open(dirname)
	if err != nil {
		log.Println("cant open " + err.Error() + " " + dirname)
		return
	}
	infos, err := file.Readdir(0)
	if err != nil {
		log.Println("cant open " + err.Error() + " " + dirname)
		return
	}

	for _, f := range infos {
		fName := dirname + "/" + f.Name()
		if f.IsDir() {
			sloc(fName, result)
		} else {
			regFile, err := os.Open(fName)
			if err != nil {
				log.Println("cant open " + err.Error() + " " + fName)
				continue
			}
			counter, err := lineCounter(regFile)
			if err != nil {
				log.Println("cant lineCounter " + err.Error() + " " + fName)
				continue
			}
			(*result)["13"] += uint64(counter)
		}
	}

}

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
