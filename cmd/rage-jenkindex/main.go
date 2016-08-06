package main 

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var inHash = flag.Int("hash", 0, "the hash to lookup")
var dictionary = map[int]string{}

func main() {
	flag.Parse()
	if *inHash == 0 {
		flag.Usage()
		return
	}

	hashes, err := Asset("data/hashes")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(hashes))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		hash, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}

		dictionary[hash] = parts[1]
	}

	if value, ok := dictionary[*inHash]; ok {
		fmt.Println(value)
		return
	}

	if value, ok := dictionary[reverseEndian(*inHash)]; ok {
		fmt.Println(value)
		return
	}

	os.Exit(1)
}

func reverseEndian(in int) int {
	b := []int{in & 0xff, (in >> 8) & 0xff, (in >> 16) & 0xff, (in >> 24) & 0xff}
	return int((b[0] << 24) + (b[1] << 16) + (b[2] << 8) + (b[3]))
}
