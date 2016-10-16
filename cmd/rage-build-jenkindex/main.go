package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func uniquePath(dir, base, ext string) string {
	for i := 0; ; i++ {
		path := fmt.Sprintf("%v/%v_%v.%v", dir, base, i, ext)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
}

func main() {
	fileName := os.Args[1]

	outFile := os.Args[2]

	files := &Files{
		Files: make([]File, 0),
	}

	dirname := filepath.Dir(outFile)
	basename := filepath.Base(outFile)
	ext := filepath.Ext(basename)
	name := strings.TrimSuffix(basename, ext)
	ext = ext[1:]
	outFile = uniquePath(dirname, name, ext)

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

	addHash(basename) // name with ext
	addHash(name)     // name without ext

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

	//	gzipFile := outFile + ".gz"
	//	fmt.Printf("writing %v\n", outFile)
	outWriter, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		panic(err)
	}
	defer outWriter.Close()

	//	gzipWriter := gzip.NewWriter(outWriter)
	//	defer gzipWriter.Close()

	//	gzipWriter.Name = outFile
	//	gzipWriter.Write(bin)
	//	gzipWriter.Flush()

	outWriter.Write(bin)
}
