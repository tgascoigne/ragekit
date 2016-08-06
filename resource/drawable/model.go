package drawable

import (
	"log"

	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type ModelCollection struct {
	resource.PointerCollection
	Models []*Model
}

type ModelHeader struct {
	_                  uint32 /* vtable */
	GeometryCollection resource.PointerCollection
	_                  types.Ptr32 /* Ptr to vectors */
	ShaderMappings     types.Ptr32
}

type Model struct {
	Header   ModelHeader
	Geometry []*Geometry
}

func (col *ModelCollection) Unpack(res *resource.Container) error {
	res.Parse(&col.PointerCollection)

	col.Models = make([]*Model, col.Count)
	for i := range col.Models {
		col.Models[i] = new(Model)
	}

	/* Read our model headers */
	for i, model := range col.Models {
		if err := col.Detour(res, i, func() error {
			if err := model.Unpack(res); err != nil {
				log.Printf("Error reading model")
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (model *Model) Unpack(res *resource.Container) error {
	res.Parse(&model.Header)

	geomCollection := &model.Header.GeometryCollection

	model.Geometry = make([]*Geometry, geomCollection.Count)
	for i := range model.Geometry {
		model.Geometry[i] = new(Geometry)
	}

	err := geomCollection.For(res, func(i int) error {
		geometry := model.Geometry[i]
		if err := geometry.Unpack(res); err != nil {
			return err
		}

		if model.Header.ShaderMappings.Valid() {
			if err := res.PeekElem(model.Header.ShaderMappings, i, &geometry.Shader); err != nil {
				return err
			}
		} else {
			geometry.Shader = ShaderNone
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
