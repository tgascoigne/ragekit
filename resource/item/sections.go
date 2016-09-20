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
	CEntityDef                          SectionType = 0xce501483
	CMapData                            SectionType = 0xd3593fa6
	CBaseArchetypeDef                   SectionType = 0x82d6fc83
	CTimeArchetypeDef                   SectionType = 0x76B0C56C
	CTimeCycleModifier                  SectionType = 0x674f9350
	CExtensionDefAudioCollisionSettings SectionType = 0x15deda27
	CExtensionDefAudioEmitter           SectionType = 0x2604683b
	CExtensionDefParticleEffect         SectionType = 0x3a26e5e1
	CMapTypes                           SectionType = 0xd98bb561
	CExtensionDefLadder                 SectionType = 0x821d5421
	CExtensionDefBuoyancy               SectionType = 0x2CB3D4E3
	CExtensionDefSpawnPoint             SectionType = 0xC4B2F638
	CCarGen                             SectionType = 1860713439
	CExtensionDefExplosionEffect        SectionType = 104349545
	CMloInstanceDef                     SectionType = 164374718
	CMloRoomDef                         SectionType = 186126833
	CExtensionDefDoor                   SectionType = 1965932561
	CExtensionDefProcObject             SectionType = 2565191912
	CMloPortalDef                       SectionType = 2572186314
	CExtensionDefSpawnPointOverride     SectionType = 2716862120
	CExtensionDefLightShaft             SectionType = 2718997053
	CMloArchetypeDef                    SectionType = 273704021
	CMloEntitySet                       SectionType = 3601308153
	CExtensionDefExpression             SectionType = 3870521079
	CLightAttrDef                       SectionType = 4115341947
	CExtensionDefWindDisturbance        SectionType = 569228403
	CExtensionDefLightEffect            SectionType = 663891011
	CMloTimeCycleModifier               SectionType = 807246248
	PhVerletClothCustomBounds           SectionType = 847348117

	SectionUNKNOWN1 SectionType = 1701774085
	SectionUNKNOWN2 SectionType = 1185771007
	SectionUNKNOWN3 SectionType = 1980345114
	SectionUNKNOWN4 SectionType = 2085051229
	SectionUNKNOWN5 SectionType = 2741784237
	SectionUNKNOWN6 SectionType = 3985044770
	SectionUNKNOWN7 SectionType = 975711773
	SectionUNKNOWN8 SectionType = 0xCC76A96C
	SectionUNKNOWN9 SectionType = 0xe2cbcfd4

	// Lists hashes of item definition files to import (?)
	SectionTypeRef   SectionType = 0x4a
	SectionSTRINGS   SectionType = 0x10
	SectionUNKNOWN10 SectionType = 0x7
	SectionUNKNOWN11 SectionType = 0x15
	SectionUNKNOWN12 SectionType = 0x33
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
