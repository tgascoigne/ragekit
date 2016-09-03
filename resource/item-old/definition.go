package item

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tgascoigne/ragekit/jenkins"
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

	SectionListPtr2 types.Ptr32
	Unk9            uint32
	Unk10Ptr        types.Ptr32
	Unk11           uint32

	SectionListPtr types.Ptr32
	Unk13          uint32
	Unk14          uint32
	Unk15          uint32

	Unk16        types.Ptr32
	Unk17        uint32
	NumSections2 uint16
	UnkCount19   uint16
	NumSections  uint16
	Unk21        uint16
}

type Definition struct {
	Header         DefinitionHeader
	FileName       string
	SectionList    []*SectionDef1
	SectionList2   []*SectionDef2
	Unknown3       DefUnk3Struct
	DefUnk10Struct DefUnk10Struct
	DefUnk10Entry  []DefUnk10Entry
	Sections       Sections
	Sections2      Sections
	StringTable    StringTable
}

type DefUnk3Struct struct {
	Unk1 [4]uint32
}

type DefUnk10Struct struct {
	Hash  [2]jenkins.Jenkins32
	Ptr   types.Ptr32
	Nil1  uint32
	Count uint32
	Nil2  uint32
}

type DefUnk10Entry struct {
	Hash  jenkins.Jenkins32
	Index uint32
}

func NewDefinition(filename string) *Definition {
	return &Definition{
		FileName:     filename,
		Sections:     make(Sections),
		Sections2:    make(Sections),
		SectionList:  make([]*SectionDef1, 0),
		SectionList2: make([]*SectionDef2, 0),
	}
}

func (ytyp *Definition) UnpackSection(res *resource.Container, sectionDef SectionDef, out Sections) (err error) {
	switch sectionDef.GetType() {

	case SectionUNKNOWN6:
		err = res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionUNKNOWN6] {
				section := new(Unknown6Section)
				res.Parse(section)
				out.Add(SectionUNKNOWN6, section)
			}
			return nil
		})

	case SectionUNKNOWN7:
		err = res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionUNKNOWN7] {
				section := new(Unknown7Section)
				res.Parse(section)
				out.Add(SectionUNKNOWN7, section)
			}
			return nil
		})

	case SectionUNKNOWN8:
		err = res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionUNKNOWN8] {
				section := new(Unknown8Section)
				res.Parse(section)
				out.Add(SectionUNKNOWN8, section)
			}
			return nil
		})

	case SectionUNKNOWN10:
		err = res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionUNKNOWN10] {
				section := new(Unknown10Section)
				res.Parse(section)
				out.Add(SectionUNKNOWN10, section)
			}
			return nil
		})

	case SectionTOBJ:
		err = res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionTOBJ] {
				section := new(TOBJSection)
				res.Parse(section)
				out.Add(SectionTOBJ, section)
			}
			return nil
		})

	case SectionOBJ:
		err = res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionOBJ] {
				section := new(OBJSection)
				res.Parse(section)
				out.Add(SectionOBJ, section)
			}
			return nil
		})

	case SectionSTRINGS:
		err = res.Detour(sectionDef.GetPtr(), func() error {
			length := int64(sectionDef.GetSize())
			if ytyp.StringTable != nil {
				panic("multiple string tables!")
			}
			ytyp.StringTable = make(StringTable, length)
			copy(ytyp.StringTable, res.Data[res.Tell():res.Tell()+length])
			return nil
		})

	default:
		fmt.Printf("Unknown section type: %v\n", sectionDef.GetType())
	}
	return err
}

func (ytyp *Definition) Unpack(res *resource.Container, outpath string) error {
	res.Parse(&ytyp.Header)

	/* parse the section table */
	err := res.Detour(ytyp.Header.SectionListPtr, func() error {
		for i := 0; i < int(ytyp.Header.NumSections); i++ {
			sectionDef := new(SectionDef1)
			res.Parse(sectionDef)
			err := ytyp.UnpackSection(res, sectionDef, ytyp.Sections)
			if err != nil {
				return err
			}

			ytyp.SectionList = append(ytyp.SectionList, sectionDef)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = res.Detour(ytyp.Header.SectionListPtr2, func() error {
		for i := 0; i < int(ytyp.Header.NumSections2); i++ {
			sectionDef := new(SectionDef2)
			res.Parse(sectionDef)
			err := ytyp.UnpackSection(res, sectionDef, ytyp.Sections2)
			if err != nil {
				return err
			}

			ytyp.SectionList2 = append(ytyp.SectionList2, sectionDef)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = res.Detour(ytyp.Header.Unk3Ptr, func() error {
		res.Parse(&ytyp.Unknown3)
		return nil
	})
	if err != nil {
		return err
	}

	err = res.Detour(ytyp.Header.Unk10Ptr, func() error {
		res.Parse(&ytyp.DefUnk10Struct)
		count := ytyp.DefUnk10Struct.Count
		ytyp.DefUnk10Entry = make([]DefUnk10Entry, count)

		err = res.Detour(ytyp.DefUnk10Struct.Ptr, func() error {
			for i := 0; i < int(count); i++ {
				res.Parse(&ytyp.DefUnk10Entry[i])
			}
			return nil
		})

		return err
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
