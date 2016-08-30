package jenkins

import (
	"fmt"
	"strings"
)

/* stolen from https://gist.github.com/Chase-san/5556547 */

type Jenkins32 uint32

func (j Jenkins32) String() string {
	if Index != nil {
		results := Lookup(j)
		switch {
		case len(results) == 0:
			return fmt.Sprintf("%v", uint32(j))
		case len(results) == 1:
			return results[0]
		default:
			return fmt.Sprintf("[%v]", strings.Join(results, ", "))
		}
	}
	return fmt.Sprintf("%v", uint32(j))
}

func (j Jenkins32) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", j.String())), nil
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

func (h *Jenkins) Reset() {
	h.hash = 0
}
