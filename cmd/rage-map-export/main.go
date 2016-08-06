package main 

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/tgascoigne/ragekit/cmd/rage-mapexport/ymap"
	"github.com/tgascoigne/ragekit/resource"
)

func main() {

	var data []byte
	var err error

	log.SetFlags(0)

	/* Read the file */
	in_file := os.Args[1]
	log.Printf("Exporting %v\n", os.Args[1])

	if data, err = ioutil.ReadFile(in_file); err != nil {
		log.Fatal(err)
	}

	/* Set the architecture */
	switch {
	case strings.Contains(in_file, "xmap"):
		resource.SetArch(resource.Arch360)
	case strings.Contains(in_file, "ymap"):
		resource.SetArch(resource.ArchPC)
	default:
		panic(fmt.Sprintf("unknown architecture, path: %v", in_file))
	}

	/* Unpack the container */
	res := new(resource.Container)
	if err = res.Unpack(data); err != nil {
		log.Fatal(err)
	}

	/* Unpack the map at 0x10 */
	ymap := ymap.NewMap(path.Base(in_file), uint32(len(data)))

	if err = ymap.Unpack(res, os.Args[2]); err != nil {
		log.Fatal(err)
	}
}
