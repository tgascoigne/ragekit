package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/tgascoigne/ragekit/jenkins"
)

var exhaustive = flag.Bool("exhaustive", false, "exhaustive hash (hash all substrings)")
var index = flag.Bool("index", false, "output in jenkindex form")
var representation = flag.String("repr", "uint", "output representation (uint, int, hex)")

var hashFunc = jenkins.New()

func main() {
	flag.Parse()
	var doHash func(string, chan string)

	if *exhaustive {
		doHash = hashSubstrings
	} else {
		doHash = hashString
	}

	results := make(chan string, 1024*32) // arbitrarily large

	if flag.NArg() > 0 {
		hash := flag.Arg(0)
		doHash(hash, results)
		close(results)

		for str := range results {
			fmt.Println(str)
		}
		os.Exit(0)
	}

	var consGroup sync.WaitGroup
	var prodGroup sync.WaitGroup
	done := make(chan bool)

	consGroup.Add(1)
	go func() {
		defer consGroup.Done()

		for {
			select {
			case str := <-results:
				fmt.Println(str)
			default:
				// buffer is empty. If done has been triggered, quit. Otherwise, loop on
				select {
				case <-done:
					return
				default:
				}
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		prodGroup.Add(1)
		go func() {
			defer prodGroup.Done()
			doHash(line, results)
		}()
	}

	prodGroup.Wait()
	done <- true
	consGroup.Wait()
}

func format(j jenkins.Jenkins32) string {
	switch *representation {
	case "uint":
		return fmt.Sprintf("%v", j.Uint32())
	case "int":
		return fmt.Sprintf("%v", j.Int32())
	case "hex":
		return j.Hex()
	default:
		fmt.Fprintln(os.Stderr, "Unknown representation: ", *representation)
		flag.Usage()
		os.Exit(1)
	}
	return ""
}

func hashString(s string, results chan string) {
	hashFunc.UpdateArray([]uint8(s))
	hash := hashFunc.HashJenkins32()
	hashFunc.Reset()

	formatted := format(hash)

	if *index {
		results <- fmt.Sprintf("%v:%v", formatted, s)
	} else {
		results <- fmt.Sprintf(formatted)
	}
}

func hashSubstrings(s string, results chan string) {
	hashString(s, results)
	for i := 0; i < len(s); i++ {
		for j := i + 1; j <= len(s); j++ {
			hashString(s[i:j], results)
		}
	}
}
