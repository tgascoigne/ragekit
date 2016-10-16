package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/tgascoigne/ragekit/encyclopedia"
	"github.com/tgascoigne/ragekit/jenkins"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/item"
)

var recurse = flag.Bool("recursive", false, "record all files in a directory")

var assetHandlers = map[string]func(path string){
	".ymap": handlePlacement,
	".ytyp": handlePlacement,
}

func main() {
	flag.Parse()

	jenkins.ReadIndexFromEnv()

	path := flag.Arg(0)

	encyclopedia.ConnectDb("bolt://neo4j:jetpack@mimas:7687")

	wg := new(sync.WaitGroup)

	if *recurse {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			fmt.Printf("recording %v\n", path)

			wg.Add(1)
			go func() {
				defer wg.Done()

				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("error recording %v: %v", path, err)
					}
				}()

				record(path)
			}()

			return nil
		})
	} else {
		record(path)
	}

	wg.Wait()
}

func record(path string) {
	extension := filepath.Ext(path)
	if handler, ok := assetHandlers[extension]; ok {
		handler(path)
		return
	}

	panic("todo: generic handler")
}

func handlePlacement(path string) {
	var data []byte
	var err error

	if data, err = ioutil.ReadFile(path); err != nil {
		log.Fatal(err)
	}

	/* Set the architecture */
	resource.SetArch(resource.ArchPC)

	/* Unpack the container */
	res := new(resource.Container)
	if err = res.Unpack(data, filepath.Base(path), uint32(len(data))); err != nil {
		log.Print(err)
		return
	}

	/* Unpack the map at 0x10 */
	placement := item.NewDefinition(path)

	if err = placement.Unpack(res); err != nil {
		log.Print(err)
		return
	}

	nodes := make([]encyclopedia.Node, 0)
	for typ, entries := range placement.Sections {
		for _, entry := range entries {
			nodes = append(nodes, encyclopedia.PlacementRecord{typ, entry})
		}
	}

	conn := encyclopedia.NewConn()
	defer conn.Close()
	conn.Graph(nodes)
}
