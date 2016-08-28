package ymap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

const (
	SectionINST        = 0xce501483
	SectionINSTSize    = 0x80
	SectionLOD         = 0xd3593fa6
	SectionLODSize     = 0x70
	SectionSTRINGS     = 0x10
	SectionUNKNOWN     = 0x674f9350
	SectionUNKNOWNSize = 0x50
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

	Unk7Ptr types.Ptr32 /* an array of 0x20 structs.. 6 of them*/
	_       uint32
	Unk8    types.Ptr32
	_       uint32

	SectionListPtr types.Ptr32 /* talkol's section list */
	_              uint32
	_              uint32
	_              uint32

	Unk10       types.Ptr32
	_           uint32
	Unk7Count   uint16
	UnkCount2   uint16
	NumSections uint16
	UnkCount4   uint16
}

type Unk7Struct struct {
	Unk1    uint64
	Unk2    uint32 /* type of some kind */
	Nil1    uint32
	Unk3Ptr types.Ptr32
	Nil2    uint32
	Unk4    uint32
	Unk5    uint32
}

type Unk3Struct struct {
	SomeHash uint32
}

type SectionDef struct {
	SectionType uint32 /* type? */
	SizeBytes   uint32
	SectionPtr  types.Ptr32
	Unk         uint32
}

type InstSection struct {
	Nil1      uint64
	ModelHash uint32
	Unk1      uint32

	Unk2 [4]uint32

	Position types.Vec4f

	Unk3 types.Vec4f

	Rotation types.Vec4f

	Unk4     uint32
	UnkCount uint32
	Unk5     [2]uint32

	Unk6 types.Vec4f

	Unk7 types.Vec4f
}

type LODSection struct {
	Nil1         uint64
	ModelHash    uint32
	LODModelHash uint32
	Unk1         [4]uint32
	Positions    [4]types.Vec4f
	Unk4         [8]uint16
}

type UnknownSection struct {
	Nil1        uint64
	UnkConstant uint32
	Nil2        uint32
	Positions   [2]types.Vec4f
	Unk1        [4]uint32
}

type Map struct {
	Header          MapHeader `json:"-"`
	Somethings      []Unk7Struct
	SectionList     []SectionDef
	InstSections    []InstSection
	LODSections     []LODSection
	UnknownSections []UnknownSection
	StringTable     []byte
}

func NewMap() *Map {
	return &Map{}
}

func (ymap *Map) Unpack(res *resource.Container, outpath string) error {
	res.Parse(&ymap.Header)

	fmt.Printf("Header: %#v\n", ymap.Header)

	/* parse the section table */
	err := res.Detour(ymap.Header.SectionListPtr, func() error {
		count := ymap.Header.NumSections
		ymap.SectionList = make([]SectionDef, count)
		ymap.InstSections = make([]InstSection, 0)
		ymap.UnknownSections = make([]UnknownSection, 0)
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

			case SectionLOD:
				res.Detour(ymap.SectionList[i].SectionPtr, func() error {
					for j := 0; j < int(ymap.SectionList[i].SizeBytes); j += SectionLODSize {
						section := new(LODSection)
						res.Parse(section)
						ymap.LODSections = append(ymap.LODSections, *section)
						fmt.Printf("LOD %#v\n", section)
					}
					return nil
				})

			case SectionUNKNOWN:
				res.Detour(ymap.SectionList[i].SectionPtr, func() error {
					for j := 0; j < int(ymap.SectionList[i].SizeBytes); j += SectionUNKNOWNSize {
						section := new(UnknownSection)
						res.Parse(section)
						ymap.UnknownSections = append(ymap.UnknownSections, *section)
						fmt.Printf("Unknown %#v\n", section)
					}
					return nil
				})

			case SectionSTRINGS:
				res.Detour(ymap.SectionList[i].SectionPtr, func() error {
					fmt.Printf("String table: %x\n", res.Tell())
					length := int64(ymap.SectionList[i].SizeBytes)
					ymap.StringTable = make([]byte, length)
					copy(ymap.StringTable, res.Data[res.Tell():res.Tell()+length])
					return nil
				})

			default:
				fmt.Printf("Unknown section type: %x\n", ymap.SectionList[i].SectionType)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	/* parse the unknown table */
	err = res.Detour(ymap.Header.Unk7Ptr, func() error {
		count := ymap.Header.Unk7Count
		ymap.Somethings = make([]Unk7Struct, count)
		for i := 0; i < int(count); i++ {
			res.Parse(&ymap.Somethings[i])
			fmt.Printf("Something %#v\n", ymap.Somethings[i])
			res.Detour(ymap.Somethings[i].Unk3Ptr, func() error {
				something2 := new(Unk3Struct)
				res.Parse(something2)
				fmt.Printf("Something2 %#v\n", something2)
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(ymap, "", "\t")
	if err != nil {
		return err
	}

	fmt.Printf("Writing %v\n", outpath)
	err = ioutil.WriteFile(outpath, data, 0744)
	if err != nil {
		return err
	}

	return nil
}
