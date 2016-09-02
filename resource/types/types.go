package types

import (
	"encoding/json"
	"fmt"
	"math"
	"unsafe"

	"github.com/tgascoigne/ragekit/jenkins"
)

type Ptr32 uint32

func (p Ptr32) Valid() bool {
	return p != 0
}

func (p Ptr32) Partition() uint32 {
	return (uint32(p) >> 24) & 0xFF
}

func (p Ptr32) PartitionOffset() uint32 {
	return (uint32(p) & 0xFFFFFF)
}

func (p Ptr32) String() string {
	return fmt.Sprintf("%x (partition %x)", p.PartitionOffset(), p.Partition())
}

func (p Ptr32) MarshalJSON() ([]byte, error) {
	result := fmt.Sprintf("\"%v\"", p.String())
	return []byte(result), nil
}

type FixedString [64]byte

func (s FixedString) String() string {
	result := ""
	for _, b := range s {
		if b == 0 {
			break
		}
		result = result + string(b)
	}
	return result
}

func (s FixedString) MarshalJSON() ([]byte, error) {
	result := fmt.Sprintf("\"%v\"", s.String())
	return []byte(result), nil
}

type Float16 uint16

func (i Float16) Value() float32 {
	/* Lovingly adapted from http://stackoverflow.com/a/15118210 */
	t1 := uint32(i & 0x7fff)
	t2 := uint32(i & 0x8000)
	t3 := uint32(i & 0x7c00)
	t1 <<= 13
	t2 <<= 16
	t1 += 0x38000000
	if t3 == 0 {
		t1 = 0
	}
	t1 |= t2
	return math.Float32frombits(t1)
}

func (i Float16) String() string {
	return fmt.Sprintf("%v", i.Value())
}

func (i Float16) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("\"%v\"", i.String())
	return []byte(str), nil
}

type UV struct {
	U Float16
	V Float16
}

type Tri struct {
	A uint16
	B uint16
	C uint16
}

type Vec3i [3]int16
type Vec3h [3]Float16
type WorldCoord Vec3

type WorldCoordh Vec3h

type Float32 float32

func (f Float32) MarshalJSON() ([]byte, error) {
	var out string
	if math.IsNaN(float64(f)) {
		out = fmt.Sprintf("null")
	} else {
		out = fmt.Sprintf("%v", f)
	}
	return []byte(out), nil
}

type Vec4f [4]Float32

type Unknown32 uint32

func (u Unknown32) MarshalJSON() ([]byte, error) {
	if u == 0 {
		return json.Marshal("0")
	}

	result := map[string]interface{}{
		"hash":    jenkins.Jenkins32(u),
		"float32": *(*Float32)(unsafe.Pointer(&u)),
		"float16": []Float16{Float16((u >> 16) & 0xFFFF), Float16(u & 0xFFFF)},
		"uint16":  []uint16{uint16((u >> 16) & 0xFFFF), uint16(u & 0xFFFF)},
		"bytes":   []uint32{uint32((u >> 24) & 0xFF), uint32((u >> 16) & 0xFF), uint32((u >> 8) & 0xFF), uint32(u & 0xFF)},
		"hex":     fmt.Sprintf("0x%x", uint32(u)),
	}

	return json.Marshal(result)
}
