package item

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type DefinitionHeader struct {
	Unk1    uint32
	Unk2    uint32
	Unk3Ptr types.Ptr32
	Unk4    uint32

	Unk4Ptr types.Ptr32
	Unk5    uint32
	Unk6    uint32
	Unk7    uint32

	Unk8Ptr  types.Ptr32
	Unk9     uint32
	Unk10Ptr types.Ptr32
	Unk11    uint32

	SectionListPtr types.Ptr32
	Unk13          uint32
	Unk14          uint32
	Unk15          uint32

	Unk16       types.Ptr32
	Unk17       uint32
	UnkCount18  uint16
	UnkCount19  uint16
	NumSections uint16
	Unk21       uint16
}

type Definition struct {
	Header      DefinitionHeader
	FileName    string
	SectionList []SectionDef
	Sections    Sections
	StringTable StringTable
}

func NewDefinition(filename string) *Definition {
	return &Definition{
		FileName: filename,
		Sections: make(Sections),
	}
}

func (ytyp *Definition) Unpack(res *resource.Container, outpath string) error {
	res.Parse(&ytyp.Header)

	/* parse the section table */
	err := res.Detour(ytyp.Header.SectionListPtr, func() error {
		count := ytyp.Header.NumSections
		ytyp.SectionList = make([]SectionDef, count)
		for i := 0; i < int(count); i++ {
			res.Parse(&ytyp.SectionList[i])

			switch ytyp.SectionList[i].Type {

			case SectionUNKNOWN6:
				res.Detour(ytyp.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ytyp.SectionList[i].Size); j += SectionSize[SectionUNKNOWN6] {
						section := new(Unknown6Section)
						res.Parse(section)
						ytyp.Sections.Add(ytyp.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionUNKNOWN7:
				res.Detour(ytyp.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ytyp.SectionList[i].Size); j += SectionSize[SectionUNKNOWN7] {
						section := new(Unknown7Section)
						res.Parse(section)
						ytyp.Sections.Add(ytyp.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionUNKNOWN8:
				res.Detour(ytyp.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ytyp.SectionList[i].Size); j += SectionSize[SectionUNKNOWN8] {
						section := new(Unknown8Section)
						res.Parse(section)
						ytyp.Sections.Add(ytyp.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionUNKNOWN10:
				res.Detour(ytyp.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ytyp.SectionList[i].Size); j += SectionSize[SectionUNKNOWN10] {
						section := new(Unknown10Section)
						res.Parse(section)
						ytyp.Sections.Add(ytyp.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionTOBJ:
				res.Detour(ytyp.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ytyp.SectionList[i].Size); j += SectionSize[SectionTOBJ] {
						section := new(TOBJSection)
						res.Parse(section)
						ytyp.Sections.Add(ytyp.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionOBJ:
				res.Detour(ytyp.SectionList[i].Ptr, func() error {
					for j := 0; j < int(ytyp.SectionList[i].Size); j += SectionSize[SectionOBJ] {
						section := new(OBJSection)
						res.Parse(section)
						ytyp.Sections.Add(ytyp.SectionList[i].Type, section)
					}
					return nil
				})

			case SectionSTRINGS:
				res.Detour(ytyp.SectionList[i].Ptr, func() error {
					length := int64(ytyp.SectionList[i].Size)
					ytyp.StringTable = make(StringTable, length)
					copy(ytyp.StringTable, res.Data[res.Tell():res.Tell()+length])
					return nil
				})

			default:
				fmt.Printf("Unknown section type: %v\n", ytyp.SectionList[i].Type)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(ytyp, "", "\t")
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
