package item

import (
	"github.com/tgascoigne/ragekit/jenkins"
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
	SectionTOBJ      SectionType = 0x76B0C56C
	// Lists hashes of item definition files to import (?)
	SectionDefinitions SectionType = 0x4a
)

var SectionSize = map[SectionType]int{
	SectionINST:        0x80,
	SectionLOD:         0x70,
	SectionTOBJ:        0xa0,
	SectionOBJ:         0x90,
	SectionDefinitions: 0x4,

	SectionUNKNOWN1:  0x50,
	SectionUNKNOWN2:  0x8,
	SectionUNKNOWN3:  0x12,
	SectionUNKNOWN4:  0x4,
	SectionUNKNOWN6:  0x40,
	SectionUNKNOWN7:  0x60,
	SectionUNKNOWN8:  0x50,
	SectionUNKNOWN10: 0xa0,
}

type SectionDef struct {
	Type SectionType
	Size uint32
	Ptr  types.Ptr32
	Unk  uint32
}

type Section interface{}

type InstSection struct {
	Nil1      uint64
	ModelHash jenkins.Jenkins32
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
	ModelHash    jenkins.Jenkins32
	LODModelHash jenkins.Jenkins32
	Unk1         [4]uint32
	Positions    [4]types.Vec4f
	Unk4         [8]uint16
}

type Unknown1Section struct {
	Nil1        uint64
	UnkConstant uint32
	Nil2        uint32
	Positions   [2]types.Vec4f
	Unk1        [4]uint32
}

type Unknown2Section struct {
	Bytes [0x8]uint8
}

type Unknown3Section struct {
	Value [6]uint16
}

type Unknown4Section struct {
	Value [4]uint8
}

type Unknown6Section struct {
	Nil1      uint64
	ModelHash jenkins.Jenkins32
	Unk1      uint32

	Position types.Vec4f

	Unk3 types.Vec4f

	Hash2 jenkins.Jenkins32
	Nil2  [3]uint32
}

type Unknown7Section struct {
	Unk1 [0x18]types.Unknown32
}

type Unknown8Section struct {
	Nil1   [6]uint32
	Counts [8]uint16
	Hash   jenkins.Jenkins32
	Nil2   [9]uint32
}

type Unknown10Section struct {
	Nil1   [4]uint32
	Unk1   types.Vec4f
	Unk2   types.Vec4f
	Unk3   [2]uint32
	Hash   jenkins.Jenkins32
	Nil2   byte
	String types.FixedString
	Nil3   [3]byte
	Unk4   [4]jenkins.Jenkins32
	Unk5   types.Vec4f
}

type OBJSection struct {
	Unk1 [4]uint32
	Nil1 [4]uint32

	BoundsMin types.Vec4f
	BoundsMax types.Vec4f

	Rotation types.Vec4f

	Radius      float32
	Unk4        uint32
	ModelHash   jenkins.Jenkins32
	TextureHash jenkins.Jenkins32
	Nil2        [2]uint32
	UnkHash     jenkins.Jenkins32
	Unk5        uint32
	Model2Hash  jenkins.Jenkins32
	Nil3        [3]uint32
	Nil4        [4]jenkins.Jenkins32
}

type TOBJSection struct {
	Unk1 [4]uint32
	Nil1 [4]uint32

	BoundsMin types.Vec4f
	BoundsMax types.Vec4f

	Rotation types.Vec4f

	Radius      float32
	Unk4        uint32
	ModelHash   jenkins.Jenkins32
	TextureHash jenkins.Jenkins32
	Nil2        [3]uint32
	Unk5        uint32
	UnkHash     jenkins.Jenkins32
	Unk6        [3]uint32
	Nil4        [4]uint32
	Unk7        uint32
	Nil5        [3]uint32
}

type DefinitionsSection struct {
	Hash jenkins.Jenkins32
}
