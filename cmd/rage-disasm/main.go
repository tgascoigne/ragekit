package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/script"
)

func main() {

	var data []byte
	var err error

	log.SetFlags(0)

	/* Read the file */
	in_file := os.Args[1]
	log.Printf("Disassembling %v\n", os.Args[1])

	if data, err = ioutil.ReadFile(in_file); err != nil {
		log.Fatal(err)
	}

	/* Set the architecture */
	switch {
	case strings.Contains(in_file, "xsc"):
		resource.SetArch(resource.Arch360)
	case strings.Contains(in_file, "ysc"):
		resource.SetArch(resource.ArchPC)
	default:
		panic(fmt.Sprintf("unknown architecture, path: %v", in_file))
	}

	/* Unpack the container */
	res := new(resource.Container)
	if err = res.Unpack(data, path.Base(in_file), uint32(len(data))); err != nil {
		log.Fatal(err)
	}

	/* Unpack the script at 0x10 */
	outScript := script.NewScript(path.Base(in_file), uint32(len(data)))
	if err = outScript.LoadNativeDB("./natives.json", "./native_translation.dat"); err != nil {
		log.Printf("Unable to load hash dictionary (%v). Lookups will be unavailable\n", err)
	}

	var emitFunc script.EmitFunc

	if len(os.Args) > 2 {
		outFile, err := os.Create(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}

		defer outFile.Close()

		emitFunc = func(istr script.Instruction) {
			outFile.WriteString(fmt.Sprintf("%v\n", istr.String()))
		}
	} else {
		emitFunc = func(istr script.Instruction) {
			fmt.Println(istr.String())
		}
	}

	if err = outScript.Unpack(res, emitFunc); err != nil {
		log.Fatal(err)
	}
}
