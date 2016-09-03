package item

import (
	"encoding/json"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type SectionType uint32

//go:generate stringer -type=SectionType

func (s SectionType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + s.String() + "\""), nil
}

const (
	SectionINST    SectionType = 0xce501483
	SectionLOD     SectionType = 0xd3593fa6
	SectionOBJ     SectionType = 0x82d6fc83
	SectionTOBJ    SectionType = 0x76B0C56C
	SectionSTRINGS SectionType = 0x10

	SectionUNKNOWN1  SectionType = 0x674f9350
	SectionUNKNOWN2  SectionType = 0x7
	SectionUNKNOWN3  SectionType = 0xe2cbcfd4
	SectionUNKNOWN4  SectionType = 0x15
	SectionUNKNOWN5  SectionType = 0x15deda27
	SectionUNKNOWN6  SectionType = 0x2604683b
	SectionUNKNOWN7  SectionType = 0x3a26e5e1
	SectionUNKNOWN8  SectionType = 0xd98bb561
	SectionUNKNOWN9  SectionType = 0x821d5421
	SectionUNKNOWN10 SectionType = 0xCC76A96C
	SectionUNKNOWN11 SectionType = 0x2CB3D4E3
	SectionUNKNOWN12 SectionType = 0xC4B2F638
	SectionUNKNOWN13 SectionType = 0x33
	// Lists hashes of item definition files to import (?)
	SectionTypeRef SectionType = 0x4a
)

type SectionPtr struct {
	Type SectionType
	Size uint32
	Ptr  types.Ptr32
	Unk  uint32
}

type Sections map[SectionType][]SectionEntry

func (s Sections) MarshalJSON() ([]byte, error) {
	// The key of maps aren't serialized using MarshalJSON, so we need to convert it to a map[String]
	m := make(map[string][]SectionEntry)
	for k, v := range s {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

func (s Sections) Add(typ SectionType, section SectionEntry) {
	if _, ok := s[typ]; !ok {
		s[typ] = make([]SectionEntry, 0)
	}

	s[typ] = append(s[typ], section)
}

type SectionEntry map[FieldName]FieldValue

func (s SectionEntry) MarshalJSON() ([]byte, error) {
	// The key of maps aren't serialized using MarshalJSON, so we need to convert it to a map[String]
	m := make(map[string]FieldValue)
	for k, v := range s {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

func (s SectionEntry) UnpackFromMap(res *resource.Container, baseAddr types.Ptr32, sectionMap []SectionMapField) error {
	for _, field := range sectionMap {
		value, err := field.UnpackField(res, baseAddr)
		if err != nil {
			return err
		}

		s[field.FieldName] = value
	}

	return nil
}
