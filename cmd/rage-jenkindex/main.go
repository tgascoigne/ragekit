package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/tgascoigne/ragekit/jenkins"
)

var inHash = flag.Uint("hash", 0, "the hash to lookup")
var interactive = flag.Bool("interactive", false, "launch an interactive cli")

func main() {
	flag.Parse()
	if *inHash == 0 && !*interactive {
		flag.Usage()
		return
	}

	jenkins.ReadIndexFromEnv()

	if *inHash != 0 {
		result := jenkins.Lookup(jenkins.Jenkins32(*inHash))
		fmt.Println(result)
	}

	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			text, _ := reader.ReadString('\n')
			base := 10
			if text[:2] == "0x" {
				base = 16
				text = text[2:]
			}
			hash, err := strconv.ParseUint(text[:len(text)-1], base, 32)
			if err != nil {
				fmt.Printf("Error parsing input: %v\n", err)
				continue
			}
			fmt.Printf("searching for %v...\n", text)
			result := jenkins.Lookup(jenkins.Jenkins32(hash))
			fmt.Printf("result: %v\n", result)
		}
	}

	os.Exit(1)
}

func reverseEndian(in jenkins.Jenkins32) jenkins.Jenkins32 {
	b := []jenkins.Jenkins32{in & 0xff, (in >> 8) & 0xff, (in >> 16) & 0xff, (in >> 24) & 0xff}
	return jenkins.Jenkins32((b[0] << 24) + (b[1] << 16) + (b[2] << 8) + (b[3]))
}
