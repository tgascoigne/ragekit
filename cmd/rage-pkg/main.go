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

	"github.com/tgascoigne/ragekit/resource"
)

var (
	recursive = flag.Bool("recursive", false, "recursively extract nested RPF files")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	if flag.NArg() < 2 {
		log.Fatal("Usage: program [-recursive] <input_file> <output_directory>")
	}

	inFile := flag.Arg(0)
	outDir := flag.Arg(1)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outDir, 0777); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	doExport(inFile, outDir)
}

func uniquePath(dir, base, ext string) string {
	for i := 0; ; i++ {
		path := fmt.Sprintf("%v/%v_%v.%v", dir, base, i, ext)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
}

func doExport(inFile, outDir string) {
	var data []byte
	var err error

	log.Printf("Unpacking %v to %v\n", inFile, outDir)
	if data, err = ioutil.ReadFile(inFile); err != nil {
		log.Fatal(err)
	}

	/* Set the architecture */
	resource.SetArch(resource.ArchPC)

	/* Unpack the package */
	pkg := new(resource.Package)
	if err = pkg.Unpack(data, path.Base(inFile), uint32(len(data))); err != nil {
		log.Print(err)
		return
	}

	root := pkg.Root()
	unpack(pkg, root, []string{outDir})

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
	dirPath := filepath.Join(newPath...)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		log.Fatalf("Failed to create directory %s: %v", dirPath, err)
	}
	for _, child := range dir.Children(pkg) {
		unpack(pkg, child, newPath)
	}
}

func unpackFile(pkg *resource.Package, file resource.PackageFile, path []string) {
	newPathParts := append(path, file.Name(pkg))
	newPath := filepath.Join(newPathParts...)
	data := file.Data(pkg)
	
	if err := ioutil.WriteFile(newPath, data, 0777); err != nil {
		log.Fatalf("Failed to write file %s: %v", newPath, err)
	}

	// If recursive flag is set and this is an RPF file, extract it
	if *recursive && strings.HasSuffix(strings.ToLower(file.Name(pkg)), ".rpf") {
		// Create output directory for this RPF
		basePath := strings.TrimSuffix(newPath, filepath.Ext(newPath))
		rpfOutDir := basePath + "_rpf"
		log.Printf("Recursively extracting %s to %s\n", newPath, rpfOutDir)
		
		// Create a new goroutine to handle the recursive extraction
		doExport(newPath, rpfOutDir)
	}
}
