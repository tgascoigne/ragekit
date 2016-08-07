package jenkins

/* stolen from https://gist.github.com/Chase-san/5556547 */

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
