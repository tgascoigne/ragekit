package shader

import (
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type Header struct {
	ParameterList  types.Ptr32
	_              uint32
	ParameterCount uint8
	_              uint8 /* what are all of these? */
	_              uint16
	_              uint32
	_              uint32
	_              uint32
	_              uint32
	_              uint32
}

type Shader struct {
	Header
	Parameters  []*Parameter
	DiffusePath string
	NormalPath  string
}

func (shader *Shader) Unpack(res *resource.Container) error {
	res.Parse(&shader.Header)

	shader.Parameters = make([]*Parameter, shader.ParameterCount)
	for i := range shader.Parameters {
		shader.Parameters[i] = new(Parameter)
	}

	if err := res.Detour(shader.ParameterList, func() error {
		for _, param := range shader.Parameters {
			if err := param.Unpack(res); err != nil {
				return err
			}

			if param.parameter != nil {
				var err error
				switch param.Type {
				case ParamDiffuseBitmap:
					shader.DiffusePath, err = param.parameter.(*BitmapParameter).Get(res)
				case ParamNormalBitmap:
					shader.NormalPath, err = param.parameter.(*BitmapParameter).Get(res)
				}
				if err != nil {
					return err
				}
			}

		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
