package util

import (
	"fmt"
	"os"
	"testing"
)

func TestCountLines(t *testing.T) {
	file, err := os.Open("/tmp/big")
	if err != nil {
		panic(err)
	}
	countLines, err := CountLines(file)
	if err != nil {
		panic(err)
	}

	fmt.Println(countLines)
}
