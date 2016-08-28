package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/tgascoigne/ragekit/cmd/rage-map-export/ymap"
	"github.com/tgascoigne/ragekit/cmd/rage-map-export/ytyp"
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
	case strings.Contains(in_file, ".xmap") || strings.Contains(in_file, "xtyp"):
		resource.SetArch(resource.Arch360)
	case strings.Contains(in_file, ".ymap") || strings.Contains(in_file, "ytyp"):
		resource.SetArch(resource.ArchPC)
	default:
		panic(fmt.Sprintf("unknown architecture, path: %v", in_file))
	}

	/* Unpack the container */
	res := new(resource.Container)
	if err = res.Unpack(data, path.Base(in_file), uint32(len(data))); err != nil {
		log.Fatal(err)
	}

	switch {
	case strings.Contains(in_file, "map"):
		/* Unpack the map at 0x10 */
		ymap := ymap.NewMap()

		if err = ymap.Unpack(res, os.Args[2]); err != nil {
			log.Fatal(err)
		}

	case strings.Contains(in_file, "typ"):
		/* Unpack the map at 0x10 */
		ytyp := ytyp.NewDefinition()

		if err = ytyp.Unpack(res, os.Args[2]); err != nil {
			log.Fatal(err)
		}
	}
}
