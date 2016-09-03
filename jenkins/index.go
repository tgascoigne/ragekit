package jenkins

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var Index []string

type IndexByHash []string

// Len is part of sort.Interface.
func (idx IndexByHash) Len() int {
	return len(idx)
}

// Swap is part of sort.Interface.
func (idx IndexByHash) Swap(i, j int) {
	idx[i], idx[j] = idx[j], idx[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (idx IndexByHash) Less(i, j int) bool {
	hashI, _ := splitEntry(idx[i])
	hashJ, _ := splitEntry(idx[j])
	return hashI < hashJ
}

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

	ReadIndex(fd)
	return nil
}

func ReadIndex(reader io.Reader) {
	Index = make([]string, 0)
	scanner := bufio.NewScanner(reader)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		Index = append(Index, line)
	}
}

func splitEntry(s string) (uint32, string) {
	parts := strings.SplitN(s, ":", 2)
	hash, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		panic(err)
	}

	return uint32(hash), parts[1]
}

func Lookup(j Jenkins32) string {
	if Index == nil {
		panic("index not loaded")
	}

	var binSearch func(s []string, offset int, hash uint32) int
	binSearch = func(s []string, offset int, hash uint32) int {
		n := len(s)
		i := n / 2
		if n == 0 || i == 0 {
			return -1
		}

		h, _ := splitEntry(s[i])
		if h < hash {
			offset = offset + i
			return binSearch(s[i:], offset, hash)
		} else if h > hash {
			return binSearch(s[:i], offset, hash)
		} else if h == hash {
			return offset + i
		}
		return -1
	}

	index := binSearch(Index, 0, uint32(j))

	if index == -1 {
		return ""
	}

	return Index[index]
}
