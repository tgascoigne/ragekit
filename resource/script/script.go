package script

import (
	"fmt"
	"math"
	"runtime"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

var lineEnding string

func init() {
	if runtime.GOOS == "windows" { /* I wish Go had this in its standard lib :( */
		lineEnding = "\r\n"
	} else {
		lineEnding = "\n"
	}
}

type ScriptHeader struct {
	_        uint32
	BlockMap types.Ptr32
	_        types.Ptr32
	_        uint32

	CodeMapPtr types.Ptr32
	_          uint32
	_          uint32
	CodeLength uint32

	_           uint32
	StaticCount uint32
	_           uint32
	NativeCount uint32

	StaticTable types.Ptr32
	_           uint32
	_           uint32
	_           uint32

	NativeTable types.Ptr32
	_           types.Ptr32
	_           uint32
	_           uint32

	_ uint32
	_ uint32
	_ uint32
	_ uint32

	TitlePtr       types.Ptr32
	_              uint32
	StringTablePtr types.Ptr32
	_              uint32

	StringTableLen uint32
	_              uint32
	_              types.Ptr32
	_              types.Ptr32
}

type Script struct {
	FileName     string
	FileSize     uint32
	Header       ScriptHeader
	NativeTable  []Native64
	StaticValues []uint64
	StringTable  []byte
	Code         []*Instruction
	VM           VM
	HashTable    *NativeDB
}

func NewScript(filename string, filesize uint32) *Script {
	return &Script{
		FileName: filename,
		FileSize: filesize,
	}
}

func (script *Script) NativeLookup(hash Native64) ([]string, bool) {
	if script.HashTable == nil {
		return nil, false
	}

	entry := script.HashTable.LookupNative(hash)
	if entry == nil {
		return nil, false
	}

	return []string{entry.Name}, true
}

func (script *Script) LoadNativeDB(dictPath, xlatePath string) error {
	var err error
	script.HashTable, err = LoadNatives(dictPath)
	if err != nil {
		return err
	}

	err = script.HashTable.LoadTranslations(xlatePath)
	if err != nil {
		return err
	}

	return nil
}

type EmitFunc func(Instruction)

func (script *Script) Unpack(res *resource.Container, emitFn EmitFunc) (err error) {
	res.Parse(&script.Header)

	script.VM.Init(script)

	/* parse the static initializers */
	err = res.Detour(script.Header.StaticTable, func() error {
		count := script.Header.StaticCount
		script.StaticValues = make([]uint64, count)
		for i := 0; i < int(count); i++ {
			res.Parse(&script.StaticValues[i])
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Couldn't parse static initializers: %v\n", err)
	}

	/* parse the native table */
	err = res.Detour(script.Header.NativeTable, func() error {
		count := script.Header.NativeCount
		script.NativeTable = make([]Native64, count)
		for i := 0; i < int(count); i++ {
			var mangledNative Native64
			res.Parse(&mangledNative)
			script.NativeTable[i] = mangledNative.unmangle(script.Header.CodeLength, i)
			//fmt.Printf("native %v is %x (%x)\n", i, script.NativeTable[i], mangledNative)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Couldn't parse native table: %v\n", err)
	}

	/* parse the string table */
	err = res.Detour(script.Header.StringTablePtr, func() error {
		var blockAddr types.Ptr32
		toRead := int(script.Header.StringTableLen)
		script.StringTable = make([]byte, script.Header.StringTableLen)
		for i := 0; i < 4; i++ {
			/* get the next block */
			res.Parse(&blockAddr)
			if !blockAddr.Valid() {
				return nil
			}

			/* parse it */
			res.Detour(blockAddr, func() error {
				offset := int(script.Header.StringTableLen) - toRead
				length := int(math.Min(float64(0x4000), float64(toRead)))
				res.Parse(script.StringTable[offset : offset+length])
				toRead -= length
				return nil
			})
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Couldn't parse string table: %v\n", err)
	}

	/* disassemble */
	err = res.Detour(script.Header.CodeMapPtr, func() error {
		var blockAddr types.Ptr32
		toRead := script.Header.CodeLength
		for i := 0; toRead > 0; {
			/* get the next block */
			res.Parse(&blockAddr)
			//			fmt.Printf("blockaddr %x tell %x\n", blockAddr, res.Tell())
			if !blockAddr.Valid() {
				continue
			}

			/* disassemble it */
			if err := res.Detour(blockAddr, func() error {
				script.disassembleBlock(uint32(i*0x4000), res, emitFn, toRead)
				if toRead < 0x4000 {
					toRead = 0
				} else {
					toRead -= 0x4000
				}
				return nil
			}); err != nil {
				return err
			}

			i++
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Couldn't parse code block: %v\n", err)
	}

	return nil
}

func (script *Script) disassembleBlock(base uint32, res *resource.Container, emitFn EmitFunc, toRead uint32) {
	startAddrReal := uint32(res.Tell())
	//	fmt.Printf("disassembling block at %x, toread %v\n", startAddrReal, toRead)

	virtAddrOffset := startAddrReal - base
	numNops := 0
	for {
		curAddrReal := uint32(res.Tell())
		//		fmt.Printf("at %v\n", curAddrReal)
		curAddrVirt := curAddrReal - virtAddrOffset
		if curAddrReal-startAddrReal >= 0x4000 {
			/* max block = 0x4000 */
			return
		}

		if (curAddrReal - startAddrReal) > toRead {
			/* end of code */
			return
		}

		istr := &Instruction{Address: curAddrVirt}
		res.Parse(&istr.Opcode)
		istr.Operation = OpType[istr.Opcode]

		/* lame way to check for end of code section */
		if istr.Operation == OpNop {
			numNops++
		} else {
			numNops = 0
		}
		if numNops >= 2 {
			return
		}

		/* Unpack operands */
		if operandFunc, ok := OperandFunc[istr.Opcode]; ok {
			istr.Operands = operandFunc()
			istr.Operands.Unpack(istr, script, res)
		} else {
			istr.Operands = &NoOperands{}
		}

		script.VM.Execute(istr)

		emitFn(*istr)
	}
}

func (script *Script) StringTableEntry(offset int) string {
	stringEnd := offset
	for script.StringTable[stringEnd] != 0 {
		stringEnd++
	}
	return string(script.StringTable[offset:stringEnd])
}
