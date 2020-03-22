package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func main() {
	sizeStr := flag.String("size", "1MB", "the size of the file")
	outputPath := flag.String("output", "", "where to save the file")

	flag.Parse()

	out, err := os.Create(*outputPath)

	size := getSize(*sizeStr)

	if err != nil {
		panic(err)
	}

	defer out.Close()

	fmt.Println("Writing", size, "bytes to", *outputPath)
	b := make([]byte, size)

	for i := 0; i < size; i++ {
		rnd := rand.Int() % 256
		b[i] = byte(rnd)

	}
	_, err = out.Write(b)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done writing", size, "bytes to", *outputPath)
}

func getSize(str string) int {
	// B, KB, MB, GB
	var t float64
	if strings.HasSuffix(str, "GB") {
		t = 1073741824
		str = strings.Replace(str, "GB", "", strings.Index(str, "GB"))
	} else if strings.HasSuffix(str, "MB") {
		str = strings.Replace(str, "MB", "", strings.Index(str, "MB"))
		t = 1048576
	} else if strings.HasSuffix(str, "KB") {
		str = strings.Replace(str, "KB", "", strings.Index(str, "KB"))
		t = 1024
	} else {
		t = 1
		str = strings.Replace(str, "B", "", strings.Index(str, "B"))
	}

	nr, err := strconv.ParseFloat(str, 32)
	if err != nil {
		panic(err)
	}

	return int(math.Floor(nr * t))
}
