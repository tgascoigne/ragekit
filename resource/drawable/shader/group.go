package shader

import (
	"log"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/texture"
	"github.com/tgascoigne/ragekit/resource/types"
)

type GroupHeader struct {
	_          uint32 /* vtable */
	TexturePtr types.Ptr32
	resource.PointerCollection
}

type Group struct {
	GroupHeader
	Shaders []*Shader
	Texture *texture.Texture
}

func (group *Group) Unpack(res *resource.Container) error {
	res.Parse(&group.GroupHeader)

	/* Read any texture dictionary */
	if group.TexturePtr.Valid() {
		if err := res.Detour(group.TexturePtr, func() error {
			group.Texture = new(texture.Texture)
			return group.Texture.Unpack(res)
		}); err != nil {
			return err
		}
	}

	/* Read our shader headers */
	group.Shaders = make([]*Shader, group.Count)
	for i := range group.Shaders {
		group.Shaders[i] = new(Shader)
	}

	/* Read the shaders */
	for i, shader := range group.Shaders {
		if err := group.Detour(res, i, func() error {
			if err := shader.Unpack(res); err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Printf("Error reading shader %v\n", i)
			return err
		}
	}

	return nil
}
