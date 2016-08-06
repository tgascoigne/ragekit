package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/tgascoigne/ragekit/resource"
)

func main() {

	var data []byte
	var err error

	log.SetFlags(0)

	/* Read the file */
	in_file := os.Args[1]

	if data, err = ioutil.ReadFile(in_file); err != nil {
		log.Fatal(err)
	}

	/* Set the architecture */
	resource.SetArch(resource.ArchPC)

	/* Unpack the container */
	res := new(resource.Container)
	if err = res.Unpack(data, path.Base(in_file), uint32(len(data))); err != nil {
		log.Fatal(err)
	}

	/* Write it out */
	os.Stdout.Write(res.Data)
}
