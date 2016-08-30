package types

import (
	"fmt"
	"math"
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

type Vec4f [4]float32

func (v Vec4f) MarshalJSON() ([]byte, error) {
	var out [4]string
	for i, f := range v {
		if math.IsNaN(float64(f)) {
			out[i] = fmt.Sprintf("null")
		} else {
			out[i] = fmt.Sprintf("%v", f)
		}
	}
	result := fmt.Sprintf("[%v, %v, %v, %v]", out[0], out[1], out[2], out[3])
	return []byte(result), nil
}
