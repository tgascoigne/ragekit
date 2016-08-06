package ymap

import (
	"fmt"
	"io/ioutil"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

const (
	SectionINST     = 0xce501483
	SectionINSTSize = 0x80
	SectionUNK      = 0xd3593fa6
	SectionUNKSize  = 0x70
)

type MapHeader struct {
	Unk1 uint32
	Unk2 uint32
	Unk3 types.Ptr32 /* zeros, a random 1 value */
	_    uint32

	Unk4 uint32
	Unk5 uint32
	_    uint32
	Unk6 uint32

	Unk7 types.Ptr32 /* an array of 0x20 structs.. 6 of them*/
	_    uint32
	Unk8 types.Ptr32
	_    uint32

	SectionListPtr types.Ptr32 /* talkol's section list */
	_              uint32
	_              uint32
	_              uint32

	Unk10       types.Ptr32
	_           uint32
	UnkCount1   uint8
	UnkCount2   uint8
	NumSections uint8
	UnkCount4   uint8
}

type UnknownStruct struct {
	Unk1 uint32 /* identifier of some kind, mentioned twice in the file */
	Unk2 uint32 /* Another identifier */
	Unk3 uint32 /* type of some kind */
	Nil1 uint32

	Unk4   types.Ptr32
	Nil2   uint32
	Unk5   uint16 /* bitfield */
	Nil3   uint16
	Nil4   uint16
	Count1 uint16
}

type UnknownStruct2 struct {
	SomeHash uint32
}

type SectionDef struct {
	SectionType uint32 /* type? */
	SizeBytes   uint32
	SectionPtr  types.Ptr32
	Unk         uint32
}

type UnknownSection struct {
	Nil1         uint64
	ModelHash    uint32
	LODModelHash uint32
	Unk1         [4]uint32
	Positions    [4][4]float32
	Unk4         [8]uint16
}

type InstSection struct {
	Nil1      uint64
	ModelHash uint32
	Unk1      uint32

	Unk2 [4]uint32

	Position [4]float32

	Unk3 [4]float32

	Rotation [4]float32

	Unk4     uint32
	UnkCount uint32
	Unk5     [2]uint32

	Unk6 [4]float32

	Unk7 [4]float32
}

type Map struct {
	FileName     string
	FileSize     uint32
	Header       MapHeader
	Somethings   []UnknownStruct
	SectionList  []SectionDef
	InstSections []InstSection
	UnkSections  []UnknownSection
}

func NewMap(filename string, filesize uint32) *Map {
	return &Map{
		FileName: filename,
		FileSize: filesize,
	}
}

func (ymap *Map) Unpack(res *resource.Container, outpath string) error {
	err := res.Deflate()
	if err != nil {
		fmt.Printf("Deflate failed: %v\n", err)
	}

	ioutil.WriteFile("test.ymap.raq", res.Data, 0744)

	res.Parse(&ymap.Header)

	fmt.Printf("Header: %#v\n", ymap.Header)

	/* parse the section table */
	err = res.Detour(ymap.Header.SectionListPtr, func() error {
		count := ymap.Header.NumSections
		ymap.SectionList = make([]SectionDef, count)
		ymap.InstSections = make([]InstSection, 0)
		for i := 0; i < int(count); i++ {
			res.Parse(&ymap.SectionList[i])
			fmt.Printf("SectionDef %#v\n", ymap.SectionList[i])

			switch ymap.SectionList[i].SectionType {
			case SectionINST:
				res.Detour(ymap.SectionList[i].SectionPtr, func() error {
					for j := 0; j < int(ymap.SectionList[i].SizeBytes); j += SectionINSTSize {
						section := new(InstSection)
						res.Parse(section)
						ymap.InstSections = append(ymap.InstSections, *section)
						fmt.Printf("INST %#v\n", section)
					}
					return nil
				})

			case SectionUNK:
				res.Detour(ymap.SectionList[i].SectionPtr, func() error {
					for j := 0; j < int(ymap.SectionList[i].SizeBytes); j += SectionUNKSize {
						section := new(UnknownSection)
						res.Parse(section)
						ymap.UnkSections = append(ymap.UnkSections, *section)
						fmt.Printf("LOD? %#v\n", section)
					}
					return nil
				})

			default:
				fmt.Printf("Unknown section type: %x\n", ymap.SectionList[i].SectionType)
			}
		}
		return nil
	})

	/* parse the unknown table */
	err = res.Detour(ymap.Header.Unk7, func() error {
		count := ymap.Header.UnkCount1
		ymap.Somethings = make([]UnknownStruct, count)
		for i := 0; i < int(count); i++ {
			res.Parse(&ymap.Somethings[i])
			fmt.Printf("Something %#v\n", ymap.Somethings[i])
			res.Detour(ymap.Somethings[i].Unk4, func() error {
				something2 := new(UnknownStruct2)
				res.Parse(something2)
				fmt.Printf("Something2 %#v\n", something2)
				return nil
			})
		}
		return nil
	})

	return nil
}
