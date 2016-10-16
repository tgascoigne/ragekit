package jenkins

import (
	"encoding/json"
	"fmt"
	"regexp"
	"unsafe"
)

/* stolen from https://gist.github.com/Chase-san/5556547 */

var propNameRegexp = regexp.MustCompile("[a-zA-Z0-9_-]+")

type Jenkins32 uint32

func (j Jenkins32) String() string {
	if Index != nil {
		result := Lookup(j)
		if result != "" {
			_, value := splitEntry(result)
			return value
		}
	}
	return fmt.Sprintf("jenkins(%v)", uint32(j))
}

func (j Jenkins32) AsPropertyName() string {
	if Index != nil {
		result := Lookup(j)
		if result != "" {
			_, value := splitEntry(result)
			if propNameRegexp.MatchString(value) {
				// name must be alphanumeric, otherwise just return unk
				return value
			}
		}
	}
	return fmt.Sprintf("unk_%v", uint32(j))
}

func (j Jenkins32) Uint32() uint32 {
	return uint32(j)
}

func (j Jenkins32) Int32() int32 {
	return *(*int32)(unsafe.Pointer(&j))
}

func (j Jenkins32) Hex() string {
	return fmt.Sprintf("0x%x", uint32(j))
}

func (j Jenkins32) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.String())
}

type Jenkins struct {
	hash uint32
}

func New() *Jenkins {
	return &Jenkins{}
}

func (h *Jenkins) Update(b uint8) {
	h.hash += uint32(b)
	h.hash += (h.hash << 10)
	h.hash ^= (h.hash >> 6)
}

func (h *Jenkins) UpdateArray(b []uint8) {
	for _, e := range b {
		h.hash += uint32(e)
		h.hash += (h.hash << 10)
		h.hash ^= (h.hash >> 6)
	}
}

func (h *Jenkins) Hash() uint32 {
	hout := h.hash
	hout += hout << 3
	hout ^= hout >> 11
	hout += hout << 15
	return hout
}

func (h *Jenkins) HashJenkins32() Jenkins32 {
	return Jenkins32(h.Hash())
}

func (h *Jenkins) Reset() {
	h.hash = 0
}
