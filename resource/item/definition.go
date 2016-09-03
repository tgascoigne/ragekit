package item

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type ItemDefinition struct {
	Header         Header
	FileName       string
	StringTable    StringTable
	Sections       Sections
	SectionMapPtrs map[SectionType]SectionMapPtr
	SectionPtrs    []SectionPtr
	SectionMaps    map[SectionType][]SectionMapField
}

func NewDefinition(filename string) *ItemDefinition {
	return &ItemDefinition{
		FileName:       filename,
		Sections:       make(Sections),
		SectionMapPtrs: make(map[SectionType]SectionMapPtr),
		SectionPtrs:    make([]SectionPtr, 0),
		SectionMaps:    make(map[SectionType][]SectionMapField),
	}
}

func (typ *ItemDefinition) Unpack(res *resource.Container, outpath string) error {
	res.Parse(&typ.Header)

	err := res.Detour(typ.Header.SectionDefPtr, func() error {
		for i := 0; i < int(typ.Header.NumSectionDefs); i++ {
			sectionMapPtr := new(SectionMapPtr)
			res.Parse(sectionMapPtr)
			typ.SectionMapPtrs[sectionMapPtr.Type] = *sectionMapPtr

			fields, err := sectionMapPtr.Unpack(res)
			if err != nil {
				return err
			}

			typ.SectionMaps[sectionMapPtr.Type] = fields
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = res.Detour(typ.Header.SectionsPtr, func() error {
		for i := 0; i < int(typ.Header.NumSections); i++ {
			sectionPtr := new(SectionPtr)
			res.Parse(sectionPtr)
			typ.SectionPtrs = append(typ.SectionPtrs, *sectionPtr)
		}
		return nil
	})

	if err != nil {
		return err
	}

	for _, section := range typ.SectionPtrs {
		sectionMap, ok := typ.SectionMaps[section.Type]
		if !ok {
			fmt.Printf("missing section map for section %v\n", section.Type)
			continue
		}

		entrySize := typ.SectionMapPtrs[section.Type].EntrySize
		sectionSize := section.Size
		fmt.Printf("%v %v %v\n", section.Type, sectionSize, entrySize)
		numEntries := sectionSize / entrySize

		for i := uint32(0); i < numEntries; i++ {
			baseAddr := section.Ptr + types.Ptr32(i*entrySize)

			entry := make(SectionEntry)
			err := entry.UnpackFromMap(res, baseAddr, sectionMap)
			if err != nil {
				return err
			}
			typ.Sections.Add(section.Type, entry)
		}
	}

	data, err := json.MarshalIndent(typ, "", "\t")
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
