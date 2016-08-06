package shader

import (
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type ParameterHeader struct {
	Type   uint32
	Offset types.Ptr32
}

type Parameter struct {
	ParameterHeader
	parameter interface{}
}

const (
	ParamDiffuseBitmap = 0x00000000
	ParamNormalBitmap  = 0x00020000
)

func (param *Parameter) Unpack(res *resource.Container) error {
	res.Parse(&param.ParameterHeader)

	if !param.Offset.Valid() {
		return nil
	}

	switch param.Type {
	case ParamDiffuseBitmap, ParamNormalBitmap:
		bitmap := new(BitmapParameter)
		if err := res.Detour(param.Offset, func() error {
			res.Parse(bitmap)
			param.parameter = bitmap
			return nil
		}); err != nil {
			return err
		}

	default:
		/* unsupported parameter. don't bother giving an error for now */
		param.parameter = nil
	}

	return nil
}

type BitmapParameter struct {
	_    uint32 /* vtable */
	_    uint32
	_    uint32
	_    uint32
	_    uint32
	_    uint32
	_    uint32
	_    uint32
	Path types.Ptr32
	_    uint32
	_    uint32
	_    uint32
	_    uint32
	_    types.Ptr32
	_    uint32
	_    uint32
}

func (bmp *BitmapParameter) Get(res *resource.Container) (string, error) {
	if !bmp.Path.Valid() {
		return "", nil
	}

	var path string
	if err := res.Detour(bmp.Path, func() error {
		res.Parse(&path)
		return nil
	}); err != nil {
		return "", err
	}

	return path, nil
}
