package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"math/rand"
	"strings"
	"fmt"
	"unsafe"

	"github.com/tgascoigne/ragekit/jenkins"
)

var dictionary = map[int]string{}

type Jenkins32 []interface{}

type Files struct {
	Files []File
}

type File struct {
	FileName string
	Hashes   map[string]Jenkins32
}

func main() {
	fileName := os.Args[1]

	outFile := os.Args[2]

	files := &Files{
		Files: make([]File, 0),
	}

	ext := filepath.Ext(outFile)
	name := strings.TrimSuffix(outFile, ext)
	fmt.Printf("in file name is %v\n", outFile)
	for {
		if _, err := os.Stat(outFile); err == nil {
			outFile = fmt.Sprintf("%v_%v.%v", name, rand.Int(), ext)
			fmt.Printf("adjusting to %v\n", outFile)
		} else {
			break
		}
	}

	hashFunc := jenkins.New()

	file := File{
		FileName: fileName,
		Hashes:   make(map[string]Jenkins32),
	}

	addHash := func(s string) {
		hashFunc.UpdateArray([]uint8(s))
		hash := hashFunc.Hash()
		hashFunc.Reset()

		values := []interface{}{uint32(hash), *(*int32)(unsafe.Pointer(&hash)), fmt.Sprintf("0x%x", hash)}
		file.Hashes[s] = values
	}

	addHash(fileName)
	basename := filepath.Base(fileName)
	addHash(basename)
	addHash(strings.TrimSuffix(basename, filepath.Ext(basename)))

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		addHash(line)
	}

	files.Files = append(files.Files, file)

	bin, err := json.MarshalIndent(files, "", "\t")
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(outFile, bin, 0744)
}
