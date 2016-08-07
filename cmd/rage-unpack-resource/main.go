package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/tgascoigne/ragekit/resource"
)

var force = flag.Bool("force", false, "force will output the input file even if it's not a valid resource")

func main() {

	var data []byte
	var err error

	log.SetFlags(0)

	flag.Parse()

	/* Read the file */
	in_file := flag.Arg(0)

	if data, err = ioutil.ReadFile(in_file); err != nil {
		log.Fatal(err)
	}

	/* Set the architecture */
	resource.SetArch(resource.ArchPC)

	/* Unpack the container */
	res := new(resource.Container)
	if err = res.Unpack(data, path.Base(in_file), uint32(len(data))); err != nil {
		if !*force {
			log.Fatal(err)
		}
	}

	/* Write it out */
	os.Stdout.Write(res.Data)
}
