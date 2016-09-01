package item

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
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

	SectionListPtr types.Ptr32
	_              uint32
	_              uint32
	_              uint32

	Unk10               types.Ptr32
	_                   uint32
	Unk7Count           uint16
	NumUnknown2Sections uint16
	NumSections         uint16
	UnkCount4           uint16
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

type Map struct {
	Header      MapHeader
	FileName    string
	SectionList []SectionDef
	Sections    Sections
	Unknown2    []Unk7Struct
	StringTable []byte
}

func NewMap(filename string) *Map {
	return &Map{
		FileName: filename,
		Sections: make(Sections),
	}
}

func (ymap *Map) Unpack(res *resource.Container, outpath string) error {
	res.Parse(&ymap.Header)

	/* parse the section table */
	err := res.Detour(ymap.Header.SectionListPtr, func() error {
		count := ymap.Header.NumSections
		ymap.SectionList = make([]SectionDef, count)
		for i := 0; i < int(count); i++ {
			res.Parse(&ymap.SectionList[i])

			switch ymap.SectionList[i].Type {
			case SectionINST:
				res.Detour(ymap.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ymap.SectionList[i].Size); j += SectionSize[SectionINST] {
						section := new(InstSection)
						res.Parse(section)
						ymap.Sections.Add(ymap.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionLOD:
				res.Detour(ymap.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ymap.SectionList[i].Size); j += SectionSize[SectionLOD] {
						section := new(LODSection)
						res.Parse(section)
						ymap.Sections.Add(ymap.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionUNKNOWN1:
				res.Detour(ymap.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ymap.SectionList[i].Size); j += SectionSize[SectionUNKNOWN1] {
						section := new(Unknown1Section)
						res.Parse(section)
						ymap.Sections.Add(ymap.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionUNKNOWN2:
				res.Detour(ymap.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ymap.SectionList[i].Size); j += SectionSize[SectionUNKNOWN2] {
						section := new(Unknown2Section)
						res.Parse(section)
						ymap.Sections.Add(ymap.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionUNKNOWN3:
				res.Detour(ymap.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ymap.SectionList[i].Size); j += SectionSize[SectionUNKNOWN3] {
						section := new(Unknown3Section)
						res.Parse(section)
						ymap.Sections.Add(ymap.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionUNKNOWN4:
				res.Detour(ymap.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ymap.SectionList[i].Size); j += SectionSize[SectionUNKNOWN4] {
						section := new(Unknown4Section)
						res.Parse(section)
						ymap.Sections.Add(ymap.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionDefinitions:
				res.Detour(ymap.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ymap.SectionList[i].Size); j += SectionSize[SectionDefinitions] {
						section := new(DefinitionsSection)
						res.Parse(section)
						ymap.Sections.Add(ymap.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionSTRINGS:
				res.Detour(ymap.SectionList[i].Ptr, func() error {
					length := int64(ymap.SectionList[i].Size)
					ymap.StringTable = make([]byte, length)
					copy(ymap.StringTable, res.Data[res.Tell():res.Tell()+length])
					return nil
				})

			default:
				fmt.Printf("Unknown section type: %v\n", ymap.SectionList[i].Type)
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
		ymap.Unknown2 = make([]Unk7Struct, count)
		for i := 0; i < int(count); i++ {
			res.Parse(&ymap.Unknown2[i])
			res.Detour(ymap.Unknown2[i].Unk3Ptr, func() error {
				something2 := new(Unk3Struct)
				res.Parse(something2)
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
