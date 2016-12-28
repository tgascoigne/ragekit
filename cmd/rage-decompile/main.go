package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/decompiler"
	"github.com/tgascoigne/ragekit/resource/script"
)

func main() {

	var data []byte
	var err error

	log.SetFlags(0)

	/* Read the file */
	in_file := os.Args[1]
	log.Printf("Decompiling %v\n", os.Args[1])

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

	code := make([]script.Instruction, 0)
	emitFunc := func(istr script.Instruction) {
		code = append(code, istr)
	}

	if err = outScript.Unpack(res, emitFunc); err != nil {
		log.Fatal(err)
	}

	decompileMachine := decompiler.NewMachine(outScript, code)
	file := decompileMachine.Decompile()
	fmt.Println(file.CString())

	/*
			in1Var := decompiler.Variable{
				Identifier: "a",
				Type:       decompiler.Uint32Type,
			}

			in2Var := decompiler.Variable{
				Identifier: "b",
				Type:       decompiler.Uint32Type,
			}

			resultVar := decompiler.Variable{
				Identifier: "result",
				Type:       decompiler.Uint32Type,
			}

			fn := decompiler.Function{
				Identifier: "add",
				In: []decompiler.Variable{
					in1Var, in2Var,
				},
				Out: resultVar,
				Statements: []decompiler.Node{
					resultVar.Declaration(),
					decompiler.AssignStmt{resultVar, decompiler.BinaryExpr{in1Var, decompiler.AddToken, in2Var}},
					decompiler.ReturnStmt{resultVar},
				},
			}
		fmt.Println(fn.CString())
	*/
}
