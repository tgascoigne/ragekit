package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/tgascoigne/ragekit/resource"
)

func main() {
	flag.Parse()

	log.SetFlags(0)

	in_file := flag.Arg(0)
	out_file := flag.Arg(1)

	doExport(in_file, out_file)
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

	log.Printf("Unpacking %v\n", in_file)

	if data, err = ioutil.ReadFile(in_file); err != nil {
		log.Fatal(err)
	}

	/* Set the architecture */
	resource.SetArch(resource.ArchPC)

	/* Unpack the package */
	pkg := new(resource.Package)
	if err = pkg.Unpack(data, path.Base(in_file), uint32(len(data))); err != nil {
		log.Print(err)
		return
	}

	root := pkg.Root()
	unpack(pkg, root, []string{})

	if unvisited := pkg.UnvisitedEntries(); len(unvisited) != 0 {
		fmt.Printf("Unvisited entries!\n")
		for _, e := range unvisited {
			fmt.Printf("%#v\n", e)
		}
	}
}

func unpack(pkg *resource.Package, node resource.PackageNode, path []string) {
	switch node := node.(type) {
	case resource.PackageDirectory:
		unpackDir(pkg, node, path)
	case resource.PackageFile:
		unpackFile(pkg, node, path)
	default:
		panic(fmt.Sprintf("unknown node type: %T", node))
	}
}

func unpackDir(pkg *resource.Package, dir resource.PackageDirectory, path []string) {
	newPath := append(path, dir.Name(pkg))
	os.MkdirAll(filepath.Join(newPath...), 0777)
	for _, child := range dir.Children(pkg) {
		unpack(pkg, child, newPath)
	}
}

func unpackFile(pkg *resource.Package, file resource.PackageFile, path []string) {
	newPathParts := append(path, file.Name(pkg))
	newPath := filepath.Join(newPathParts...)
	data := file.Data(pkg)
	err := ioutil.WriteFile(filepath.Join(newPath), data, 0777)
	if err != nil {
		panic(err)
	}
}
