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

	SectionListPtr2 types.Ptr32 /* an array of 0x20 structs.. 6 of them*/
	_               uint32
	Unk8            types.Ptr32
	_               uint32

	SectionListPtr types.Ptr32
	_              uint32
	_              uint32
	_              uint32

	Unk10               types.Ptr32
	_                   uint32
	NumSections2        uint16
	NumUnknown2Sections uint16
	NumSections         uint16
	UnkCount4           uint16
}

type Unk3Struct struct {
	SomeHash uint32
}

type Map struct {
	Header       MapHeader
	FileName     string
	SectionList  []*SectionDef1
	SectionList2 []*SectionDef2
	Sections     Sections
	Sections2    Sections
	StringTable  []byte
}

func NewMap(filename string) *Map {
	return &Map{
		FileName:     filename,
		Sections:     make(Sections),
		Sections2:    make(Sections),
		SectionList:  make([]*SectionDef1, 0),
		SectionList2: make([]*SectionDef2, 0),
	}
}

func (ymap *Map) UnpackSection(res *resource.Container, sectionDef SectionDef, out Sections) (err error) {
	switch sectionDef.GetType() {
	case SectionINST:
		res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionINST] {
				section := new(InstSection)
				res.Parse(section)
				out.Add(SectionINST, section)
			}
			return nil
		})

	case SectionLOD:
		res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionLOD] {
				section := new(LODSection)
				res.Parse(section)
				out.Add(SectionLOD, section)
			}
			return nil
		})

	case SectionUNKNOWN1:
		res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionUNKNOWN1] {
				section := new(Unknown1Section)
				res.Parse(section)
				out.Add(SectionUNKNOWN1, section)
			}
			return nil
		})

	case SectionUNKNOWN2:
		res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionUNKNOWN2] {
				section := new(Unknown2Section)
				res.Parse(section)
				out.Add(SectionUNKNOWN2, section)
			}
			return nil
		})

	case SectionUNKNOWN3:
		res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionUNKNOWN3] {
				section := new(Unknown3Section)
				res.Parse(section)
				out.Add(SectionUNKNOWN3, section)
			}
			return nil
		})

	case SectionUNKNOWN4:
		res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionUNKNOWN4] {
				section := new(Unknown4Section)
				res.Parse(section)
				out.Add(SectionUNKNOWN4, section)
			}
			return nil
		})

	case SectionDefinitions:
		res.Detour(sectionDef.GetPtr(), func() error {
			for j := 0; j < int(sectionDef.GetSize()); j += SectionSize[SectionDefinitions] {
				section := new(DefinitionsSection)
				res.Parse(section)
				out.Add(SectionDefinitions, section)
			}
			return nil
		})

	case SectionSTRINGS:
		res.Detour(sectionDef.GetPtr(), func() error {
			length := int64(sectionDef.GetSize())
			if ymap.StringTable != nil {
				panic("multiple string tables!")
			}
			ymap.StringTable = make(StringTable, length)
			copy(ymap.StringTable, res.Data[res.Tell():res.Tell()+length])
			return nil
		})

	default:
		fmt.Printf("Unknown section type: %v\n", sectionDef.GetType())
	}
	return err
}

func (ymap *Map) Unpack(res *resource.Container, outpath string) error {
	res.Parse(&ymap.Header)

	/* parse the section table */
	err := res.Detour(ymap.Header.SectionListPtr, func() error {
		for i := 0; i < int(ymap.Header.NumSections); i++ {
			sectionDef := new(SectionDef1)
			res.Parse(sectionDef)
			err := ymap.UnpackSection(res, sectionDef, ymap.Sections)
			if err != nil {
				return err
			}

			ymap.SectionList = append(ymap.SectionList, sectionDef)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = res.Detour(ymap.Header.SectionListPtr2, func() error {
		for i := 0; i < int(ymap.Header.NumSections2); i++ {
			sectionDef := new(SectionDef2)
			res.Parse(sectionDef)
			err := ymap.UnpackSection(res, sectionDef, ymap.Sections2)
			if err != nil {
				return err
			}

			ymap.SectionList2 = append(ymap.SectionList2, sectionDef)
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
