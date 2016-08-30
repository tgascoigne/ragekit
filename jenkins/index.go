package jenkins

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type IndexType map[Jenkins32][]string

var Index map[Jenkins32][]string

const (
	IndexFileEnv = "RAGEKIT_HASH_INDEX"
)

func ReadIndexFromEnv() error {
	indexFile := os.Getenv(IndexFileEnv)
	fmt.Printf("Reading index from %v\n", indexFile)
	fd, err := os.Open(indexFile)
	if err != nil {
		return err
	}
	defer fd.Close()

	return ReadIndex(fd)
}

func ReadIndex(reader io.Reader) error {
	Index = make(IndexType)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		hash, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return err
		}

		if _, ok := Index[Jenkins32(hash)]; !ok {
			Index[Jenkins32(hash)] = []string{}
		}
		Index[Jenkins32(hash)] = append(Index[Jenkins32(hash)], parts[1])
	}
	return nil
}

func Lookup(j Jenkins32) []string {
	if Index == nil {
		panic("index not loaded")
	}

	if results, ok := Index[j]; ok {
		return results
	}
	return []string{}
}
