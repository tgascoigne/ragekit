package item

import "github.com/tgascoigne/ragekit/resource/types"

type Header struct {
	Unk1    uint32
	Unk2    uint32
	Unk3Ptr types.Ptr32
	Unk4    uint32

	Unk4Ptr types.Ptr32
	Unk5    uint32
	Unk6    uint32
	Unk7    uint32

	SectionDefPtr types.Ptr32
	Unk9          uint32
	Unk10Ptr      types.Ptr32
	Unk11         uint32

	SectionsPtr types.Ptr32
	Unk13       uint32
	Unk14       uint32
	Unk15       uint32

	Unk16          types.Ptr32
	Unk17          uint32
	NumSectionDefs uint16
	UnkCount19     uint16
	NumSections    uint16
	Unk21          uint16
}
