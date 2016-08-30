package jenkins

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

var Index []string

type byHash []string

func (a byHash) Len() int      { return len(a) }
func (a byHash) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byHash) Less(i, j int) bool {
	hash1, _ := splitEntry(Index[i])
	hash2, _ := splitEntry(Index[j])
	return hash1 < hash2
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
	for scanner.Scan() {
		line := scanner.Text()
		Index = append(Index, line)
	}
	sort.Sort(byHash(Index))

	fd, err := os.Create("jenkindex.sorted")
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	for _, s := range Index {
		fmt.Fprintln(fd, s)
	}
}

/*func ReadIndex(reader io.Reader) error {
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
		// this forces a copy of parts[1]
		// I think there was some wierdness going on with the GC when value wasn't copied
		value := parts[1] + " "
		value = value[:len(value)-1]
		Index[Jenkins32(hash)] = append(Index[Jenkins32(hash)], value)
	}
	return nil
}
*/

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

	var binSearch func(s []string, hash uint32) int
	binSearch = func(s []string, hash uint32) int {
		n := len(s)
		i := n / 2

		h, _ := splitEntry(Index[i])
		if h < hash {
			return binSearch(s[i+1:], hash)
		} else if h > hash {
			return binSearch(s[:i-1], hash)
		} else if h == hash {
			return i
		}
		return -1
	}

	index := binSearch(Index, uint32(j))

	if index == -1 {
		fmt.Printf("couldn't find it'")
		return ""
	}

	return Index[index]
}
