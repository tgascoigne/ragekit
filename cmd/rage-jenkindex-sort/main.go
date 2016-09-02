package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/tgascoigne/ragekit/jenkins"
)

func main() {
	flag.Parse()

	jenkins.ReadIndex(os.Stdin)

	sort.Sort(jenkins.IndexByHash(jenkins.Index))

	for _, s := range jenkins.Index {
		fmt.Println(s)
	}
}
