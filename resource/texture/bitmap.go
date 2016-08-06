package texture

import (
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type BitmapHeader struct {
	_     uint32
	_     uint32
	_     uint32
	_     uint32
	_     uint32
	_     uint32
	_     uint32
	_     uint32
	Title types.Ptr32
}

type Bitmaps []*Bitmap

type Bitmap struct {
	Header BitmapHeader
	Title  string
}

func (b Bitmaps) Unpack(res *resource.Container, col *resource.PointerCollection) error {
	for i := 0; i < cap(b); i++ {
		if err := col.Detour(res, i, func() error {
			b[i] = new(Bitmap)
			return b[i].Unpack(res)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bitmap) Unpack(res *resource.Container) error {
	res.Parse(&b.Header)

	if b.Header.Title.Valid() {
		res.Detour(b.Header.Title, func() error {
			res.Parse(&b.Title)
			return nil
		})
	}
	return nil
}
