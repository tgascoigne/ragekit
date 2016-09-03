package item

import (
	"github.com/tgascoigne/ragekit/jenkins"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type SectionMapPtr struct {
	Type      SectionType
	Hash      jenkins.Jenkins32
	Unk1      types.Unknown32
	Nil1      types.Unknown32
	Ptr       types.Ptr32
	Nil2      types.Unknown32
	EntrySize uint32
	Nil3      uint16
	NumFields uint16
}

func (s SectionMapPtr) Unpack(res *resource.Container) ([]SectionMapField, error) {
	fields := make([]SectionMapField, s.NumFields)
	err := res.Detour(s.Ptr, func() error {
		for i := 0; i < int(s.NumFields); i++ {
			res.Parse(&fields[i])
		}

		return nil
	})
	return fields, err
}

type FieldType uint16

//go:generate stringer -type=FieldType

func (f FieldType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + f.String() + "\""), nil
}

const (
	FieldJenkins  FieldType = 0x4a
	FieldVec4f    FieldType = 0x33
	FieldFloat32  FieldType = 0x21
	FieldFlags32  FieldType = 0x15
	FieldUint32   FieldType = 0x14
	FieldUnknown1 FieldType = 0x62
	FieldUnknown2 FieldType = 0x52
)

type FieldName jenkins.Jenkins32

func (name FieldName) String() string {
	return jenkins.Jenkins32(name).String()
}

type FieldValue interface{}

type SectionMapField struct {
	FieldName FieldName
	Offset    uint32
	FieldType FieldType
	Unk1      uint16
	Unk2      types.Unknown32
}

func (f SectionMapField) UnpackField(res *resource.Container, baseAddr types.Ptr32) (result FieldValue, err error) {
	err = res.Detour(baseAddr+types.Ptr32(f.Offset), func() error {
		switch f.FieldType {
		case FieldJenkins:
			var value jenkins.Jenkins32
			res.Parse(&value)
			result = value

		case FieldVec4f:
			var value types.Vec4f
			res.Parse(&value)
			result = value

		case FieldFloat32:
			var value types.Float32
			res.Parse(&value)
			result = value

		case FieldFlags32:
			fallthrough

		case FieldUint32:
			var value uint32
			res.Parse(&value)
			result = value

		case FieldUnknown1:
			fallthrough

		case FieldUnknown2:
			var value types.Unknown32
			res.Parse(&value)
			result = value

		}
		return nil
	})
	return result, err
}
