package texture

import (
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type TextureHeader struct {
	_       uint32
	_       types.Ptr32
	_       uint32
	_       uint32
	_       resource.Collection
	Bitmaps resource.PointerCollection
}

type Texture struct {
	Header  TextureHeader
	Bitmaps Bitmaps
}

func (texture *Texture) Unpack(res *resource.Container) error {
	res.Parse(&texture.Header)

	texture.Bitmaps = make(Bitmaps, texture.Header.Bitmaps.Capacity)
	if err := texture.Bitmaps.Unpack(res, &texture.Header.Bitmaps); err != nil {
		return err
	}

	return nil
}
