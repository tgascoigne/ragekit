package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/tgascoigne/ragekit/jenkins"
)

var exhaustive = flag.Bool("exhaustive", false, "exhaustive hash (hash all substrings)")
var index = flag.Bool("index", false, "output in jenkindex form")
var representation = flag.String("repr", "uint", "output representation (uint, int, hex, none)")

var hashFunc = jenkins.New()

func main() {
	flag.Parse()
	var doHash func(string)

	if *exhaustive {
		doHash = hashSubstrings
	} else {
		doHash = hashString
	}

	if flag.NArg() > 0 {
		hash := flag.Arg(0)
		doHash(hash)
		os.Exit(0)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		doHash(line)
	}
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

func hashString(s string) {
	if *representation == "none" {
		fmt.Println(s)
		return
	}

	hashFunc.UpdateArray([]uint8(s))
	hash := hashFunc.HashJenkins32()
	hashFunc.Reset()

	formatted := format(hash)

	if *index {
		fmt.Printf("%v:%v\n", formatted, s)
	} else {
		fmt.Println(formatted)
	}
}

func hashSubstrings(s string) {
	hashString(s)
	for i := 0; i < len(s); i++ {
		for j := i + 1; j <= len(s); j++ {
			hashString(s[i:j])
		}
	}
}
