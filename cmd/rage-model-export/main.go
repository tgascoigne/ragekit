package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export"
	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export/dae"
	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export/obj"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/bounds"
	"github.com/tgascoigne/ragekit/resource/dictionary"
	"github.com/tgascoigne/ragekit/resource/drawable"
	"github.com/tgascoigne/ragekit/resource/frag"
)

var (
	SupportedExtensions = []string{".xdr", ".xdd", ".xft", ".xbn"}
)

func main() {
	var mergeFile = flag.String("merge", "", "The basename of a file to merge all output to")
	flag.BoolVar(&export.FlipYZ, "flip", false, "Flip the Z and Y axes")
	var outputObj = flag.Bool("obj", false, "Output to OBJ instead of DAE")
	flag.Parse()

	log.SetFlags(0)

	inputFiles := make([]string, 0)
	converted := 0
	defer func() {
		log.Printf("Converted %v models\n", converted)
	}()

	if flag.NArg() > 0 {
		for _, s := range flag.Args() {
			inputFiles = append(inputFiles, s)
		}
	} else {
		inputFiles = locateInputFiles("./")
	}

	/* Convert all supported models in directory */

	var object *export.ModelGroup

	var exportFunc func(export.Exportable) error
	if *outputObj {
		exportFunc = obj.Export
	} else {
		exportFunc = dae.Export
	}

	if *mergeFile != "" {
		object = export.NewModelGroup()
		object.Name = *mergeFile
	}

	doExport := func(o export.Exportable) {
		if err := exportFunc(o); err != nil {
			log.Printf(err.Error())
		}
	}

	for _, filePath := range inputFiles {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Unable to convert %v\n", filePath)
					panic(err)
				}
			}()

			log.Printf("Converting %v.. ", filePath)

			if *mergeFile == "" {
				object = export.NewModelGroup()
			}

			processModel(filePath, object)
			converted++

			if *mergeFile == "" {
				doExport(object)
			}

			log.Printf("done\n")
		}()
	}

	if *mergeFile != "" {
		doExport(object)
	}

}

func locateInputFiles(path string) []string {
	inFiles := make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		for _, ext := range SupportedExtensions {
			if strings.Contains(f.Name(), ext) {
				filePath := fmt.Sprintf("./%v", f.Name())
				inFiles = append(inFiles, filePath)
			}
		}
	}
	return inFiles
}

func processModel(inFile string, object *export.ModelGroup) {
	var data []byte
	var err error

	baseName := filepath.Base(inFile)

	/* Read the file */
	if data, err = ioutil.ReadFile(inFile); err != nil {
		panic(err)
	}

	/* Unpack the container */
	res := new(resource.Container)
	if err = res.Unpack(data, baseName, uint32(len(data))); err != nil {
		panic(err)
	}

	baseName = baseName[:strings.LastIndex(baseName, ".")]

	switch {
	case strings.Contains(filepath.Ext(inFile), "dr"):
		object.Merge(unpackDrawable(res))
	case strings.Contains(filepath.Ext(inFile), "dd"):
		object.Merge(unpackDrawableDictionary(res, baseName))
	case strings.Contains(filepath.Ext(inFile), "ft"):
		object.Merge(unpackFragType(res, baseName))
	case strings.Contains(filepath.Ext(inFile), "bn"):
		object.Merge(unpackBoundsNodes(res, baseName))
	}
}

func unpackDrawable(res *resource.Container) export.Exportable {
	drawable := new(drawable.Drawable)
	if err := drawable.Unpack(res); err != nil {
		panic(err)
	}

	return drawable.Model
}

func unpackDrawableDictionary(res *resource.Container, title string) export.Exportable {
	/* Unpack the dictionary */
	dictionary := new(dictionary.Dictionary)
	if err := dictionary.Unpack(res); err != nil {
		panic(err)

	}

	group := export.NewModelGroup()
	group.Name = title
	for _, drawable := range dictionary.Drawables {
		idx := strings.LastIndex(drawable.Title, ".")
		if idx != -1 {
			drawable.Title = drawable.Title[:strings.LastIndex(drawable.Title, ".")]
		}

		drawable.Model.Name = drawable.Title
		group.Add(drawable.Model)
	}
	return group
}

func unpackFragType(res *resource.Container, title string) export.Exportable {
	/* Unpack the frag type */
	frag := new(frag.FragType)
	if err := frag.Unpack(res); err != nil {
		panic(err)
	}

	/* Drawables inside frag files dont seem to be named properly. */
	frag.Drawable.Model.Name = title

	return frag.Drawable.Model
}

func unpackBoundsNodes(res *resource.Container, title string) export.Exportable {
	/* Unpack the nodes */
	nodes := new(bounds.Nodes)

	if err := nodes.Unpack(res); err != nil {
		panic(err)
	}

	nodes.Model.Name = title

	return nodes.Model
}
