package drawable

import (
	"fmt"
	"log"
	"strings"

	"github.com/Jragonmiris/mathgl"

	"github.com/tgascoigne/ragekit/cmd/rage-modelexport/export"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/drawable/shader"
	"github.com/tgascoigne/ragekit/resource/types"
)

var NextUnnamedIndex int = 0

type DrawableCollection struct {
	resource.PointerCollection
	Drawables []*Drawable
}

func (col *DrawableCollection) Unpack(res *resource.Container) error {
	col.Drawables = make([]*Drawable, col.Count)
	for i := range col.Drawables {
		col.Drawables[i] = new(Drawable)
	}

	/* Read our model headers */
	for i, drawable := range col.Drawables {
		if err := col.Detour(res, i, func() error {
			if err := drawable.Unpack(res); err != nil {
				log.Printf("Error reading drawable")
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

type DrawableHeader struct {
	_               uint32
	BlockMap        types.Ptr32
	ShaderTable     types.Ptr32
	SkeletonData    types.Ptr32
	Center          types.Vec4
	BoundsMin       types.Vec4
	BoundsMax       types.Vec4
	ModelCollection types.Ptr32
	LodCollections  [3]types.Ptr32
	PointMax        types.Vec4
	_               [6]uint32
	_               types.Ptr32
	Title           types.Ptr32
}

type Drawable struct {
	Header  DrawableHeader
	Shaders shader.Group
	Models  ModelCollection
	Title   string
	Model   *export.Model
}

func (drawable *Drawable) Unpack(res *resource.Container) error {
	res.Parse(&drawable.Header)

	drawable.Model = export.NewModel()

	/* unpack */
	if drawable.Header.ShaderTable.Valid() {
		if err := res.Detour(drawable.Header.ShaderTable, func() error {
			return drawable.Shaders.Unpack(res)
		}); err != nil {
			return err
		}
	}

	if err := res.Detour(drawable.Header.ModelCollection, func() error {
		return drawable.Models.Unpack(res)
	}); err != nil {
		return err
	}

	if drawable.Header.Title.Valid() {
		if err := res.Detour(drawable.Header.Title, func() error {
			res.Parse(&drawable.Title)
			drawable.Title = drawable.Title[:strings.LastIndex(drawable.Title, ".")]
			return nil
		}); err != nil {
			return err
		}
	} else {
		drawable.Title = fmt.Sprintf("unnamed_%v", NextUnnamedIndex)
		NextUnnamedIndex++
	}

	/* Load everything into our exportable */
	drawable.Model.Name = drawable.Title

	for _, shader := range drawable.Shaders.Shaders {
		material := export.NewMaterial()
		if shader.DiffusePath != "" {
			material.DiffBitmap = fmt.Sprintf("%v.dds", shader.DiffusePath)
		}
		drawable.Model.AddMaterial(material)
	}

	for _, model := range drawable.Models.Models {
		for _, geom := range model.Geometry {
			mesh := export.NewMesh()
			mesh.Material = int(geom.Shader)

			for _, vert := range geom.Vertices.Vertex {
				/* Even if a feature isn't supported, the nil value should be fine */
				newVert := export.Vertex{
					Pos: mathgl.Vec4f{
						vert.WorldCoord[0],
						vert.WorldCoord[1],
						vert.WorldCoord[2],
						1.0,
					},
					UV: mathgl.Vec2f{
						vert.UV0.U.Value(),
						(-vert.UV0.V.Value()) + 1,
					},
					Colour: vert.Colour,
				}

				mesh.AddVert(newVert)
			}
			mesh.Format = geom.Vertices.Format

			for _, face := range geom.Indices.Index {
				mesh.AddFace(*face)
			}
			drawable.Model.AddMesh(mesh)
		}
	}

	return nil
}
