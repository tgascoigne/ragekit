package item

import (
	"github.com/tgascoigne/ragekit/jenkins"
	"github.com/tgascoigne/ragekit/resource/types"
)

type SectionType int

const (
	SectionINST        SectionType = 0xce501483
	SectionINSTSize                = 0x80
	SectionLOD         SectionType = 0xd3593fa6
	SectionLODSize                 = 0x70
	SectionSTRINGS     SectionType = 0x10
	SectionUNKNOWN     SectionType = 0x674f9350
	SectionUNKNOWNSize             = 0x50
)

type SectionDef struct {
	Type SectionType
	Size uint32
	Ptr  types.Ptr32
	Unk  uint32
}

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

type UnknownSection struct {
	Nil1        uint64
	UnkConstant uint32
	Nil2        uint32
	Positions   [2]types.Vec4f
	Unk1        [4]uint32
}
