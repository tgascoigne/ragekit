package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tgascoigne/ragekit/jenkins"
)

var dictionary = map[int]string{}

type Jenkins32 uint32

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

	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		bin, err := json.Marshal(files)
		if err != nil {
			panic(err)
		}

		ioutil.WriteFile(outFile, bin, 0744)
	}

	binData, err := ioutil.ReadFile(outFile)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(binData, files)
	if err != nil {
		panic(err)
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

		file.Hashes[s] = Jenkins32(hash)
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

	bin, err := json.Marshal(files)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(outFile, bin, 0744)
}
