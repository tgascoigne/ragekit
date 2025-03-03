package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/tgascoigne/ragekit/cmd/jenkins-brute/brutedict"
	"github.com/tgascoigne/ragekit/jenkins"
)

var searches = map[jenkins.Jenkins32][]string{}
var searchesM sync.RWMutex

var outFile io.WriteCloser
var outFileM sync.Mutex

var start = flag.Int("start", 0, "set the start point")
var end = flag.Int("end", 12, "set the end point")
var words = flag.String("words", "", "a word dictionary")

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(0)
	}

	inFile, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		hash, err := strconv.ParseUint(line, 10, 32)
		if err != nil {
			panic(err)
		}
		searches[jenkins.Jenkins32(hash)] = make([]string, 0)
	}

	inFile.Close()

	if _, err := os.Stat(flag.Arg(1)); os.IsNotExist(err) {
		outFile, err = os.Create(flag.Arg(1))
		if err != nil {
			panic(err)
		}
	} else {
		outFile, err = os.OpenFile(flag.Arg(1), os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
	}
	defer outFile.Close()

	var dictChan chan string

	if *words != "" {
		wordDict := readDict(*words)
		dict := brutedict.CamelCase{brutedict.NewWordDict(wordDict, *start, *end)}
		dictChan = dict.Chan()
	} else {
		dict := brutedict.New(true, true, true, *start, *end)
		dictChan = dict.Chan()
	}

	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(dictChan, &wg)
	}

	wg.Wait()
}

func readDict(dict string) []string {
	result := make([]string, 0)
	dictFile, err := os.Open(dict)
	if err != nil {
		panic(err)
	}
	defer dictFile.Close()

	scanner := bufio.NewScanner(dictFile)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		result = append(result, line)
	}
	fmt.Println("Read dictionary ", dict)
	return result
}

func worker(s <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var hashFunc = jenkins.New()
	var hit bool
	var hash jenkins.Jenkins32

	for str := range s {
		hashFunc.UpdateArray([]uint8(str))
		hash = hashFunc.HashJenkins32()
		hashFunc.Reset()

		hit = false
		searchesM.RLock()
		_, hit = searches[hash]
		searchesM.RUnlock()

		if hit {
			searchesM.Lock()
			searches[hash] = append(searches[hash], str)
			searchesM.Unlock()

			outFileM.Lock()
			fmt.Printf("%v:%v\n", hash.Uint32(), str)
			fmt.Fprintf(outFile, "%v:%v\n", hash.Uint32(), str)
			outFileM.Unlock()
		}

	}
}
