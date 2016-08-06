package resource

import (
	"encoding/binary"

	"github.com/tgascoigne/ragekit/resource/types"
)

type Arch int

const (
	ArchPC Arch = iota
	Arch360
)

var nativeEndian binary.ByteOrder

func SetArch(a Arch) {
	switch a {
	case ArchPC:
		nativeEndian = binary.LittleEndian
	case Arch360:
		nativeEndian = binary.BigEndian
	default:
		panic("unknown architecture")
	}
}

func parseStruct(res *Container, data interface{}) error {
	return binary.Read(res, binary.BigEndian, data)
}

/* Borrowed/Adapted from encoding/binary/binary.go */
func intDataSize(data interface{}) int {
	switch data := data.(type) {
	case int8, *int8, *uint8:
		return 1
	case []int8:
		return len(data)
	case []uint8:
		return len(data)
	case int16, *int16, *uint16:
		return 2
	case []int16:
		return 2 * len(data)
	case []uint16:
		return 2 * len(data)
	case int32, *int32, *uint32:
		return 4
	case []int32:
		return 4 * len(data)
	case []uint32:
		return 4 * len(data)
	case int64, *int64, *uint64:
		return 8
	case []int64:
		return 8 * len(data)
	case []uint64:
		return 8 * len(data)
	case types.Ptr32, *types.Ptr32:
		return 4
	}
	return 0
}
