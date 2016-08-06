package dictionary

import (
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/drawable"
	"github.com/tgascoigne/ragekit/resource/types"
)

type DictionaryHeader struct {
	_                  uint32
	BlockMap           types.Ptr32
	_                  uint32
	_                  uint32
	_                  resource.Collection
	DrawableCollection resource.PointerCollection
}

type Dictionary struct {
	Header DictionaryHeader
	drawable.DrawableCollection
}

func (dict *Dictionary) Unpack(res *resource.Container) error {
	res.Parse(&dict.Header)

	dict.DrawableCollection.PointerCollection = dict.Header.DrawableCollection
	if err := dict.DrawableCollection.Unpack(res); err != nil {
		return err
	}

	return nil
}
