package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/tgascoigne/ragekit/jenkins"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/item"
)

var batch = flag.Bool("batch", false, "batch conversion")

func main() {
	flag.Parse()

	log.SetFlags(0)

	jenkins.ReadIndexFromEnv()

	in_file := flag.Arg(0)
	out_file := flag.Arg(1)

	if *batch {
		filepath.Walk(in_file, func(path string, f os.FileInfo, err error) error {
			log.Printf("looking at %v\n", path)
			if !strings.HasSuffix(path, "ymap") && !strings.HasSuffix(path, "ytyp") {
				return nil
			}

			basename := filepath.Base(path)
			outpath := uniquePath(out_file, basename, "json")
			doExport(path, outpath)
			return nil
		})
	} else {
		doExport(in_file, out_file)
	}
}

func uniquePath(dir, base, ext string) string {
	for i := 0; ; i++ {
		path := fmt.Sprintf("%v/%v_%v.%v", dir, base, i, ext)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
}

func doExport(in_file, out_file string) {
	var data []byte
	var err error

	if *batch {
		defer func() {
			if err := recover(); err != nil {
				log.Print(err)
			}
		}()
	}

	log.Printf("Exporting %v\n", in_file)

	if data, err = ioutil.ReadFile(in_file); err != nil {
		log.Fatal(err)
	}

	/* Set the architecture */
	resource.SetArch(resource.ArchPC)

	/* Unpack the container */
	res := new(resource.Container)
	if err = res.Unpack(data, path.Base(in_file), uint32(len(data))); err != nil {
		log.Print(err)
		return
	}

	/* Unpack the map at 0x10 */
	ytyp := item.NewDefinition(in_file)

	if err = ytyp.Unpack(res); err != nil {
		log.Print(err)
		return
	}

	if err = ytyp.Dump(out_file); err != nil {
		log.Print(err)
		return
	}
}
