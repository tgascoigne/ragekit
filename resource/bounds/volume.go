package bounds

import (
	"fmt"

	"github.com/Jragonmiris/mathgl"

	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type VolumeHeader struct {
	_             uint32
	_             uint32
	_             uint32
	_             float32
	_             [4]types.Vec4 /* looks like world coords also */
	_             types.Vec4
	_             uint32
	_             uint32
	_             uint16
	_             uint16 /* some count */
	IndicesAddr   types.Ptr32
	ScaleFactor   types.Vec4
	Offset        types.Vec4
	VerticesAddr  types.Ptr32
	_             uint32
	_             uint32
	_             uint16
	VertexCount   uint16
	_             uint16
	IndexCount    uint16
	_             uint32
	_             uint32
	_             uint32
	_             types.Ptr32
	_             uint32
	_             uint32
	_             uint32
	_             uint32
	MaterialMap   types.Ptr32
	_             uint32 /* count? */
	_             uint32
	BoundBoxesPtr types.Ptr32
	_             uint32
	_             uint32
	_             uint32
}

type Volume struct {
	VolumeHeader
	VolumeInfo
	*export.Mesh
}

func (vol *Volume) Unpack(res *resource.Container) error {
	res.Parse(&vol.VolumeHeader)

	vol.Mesh = export.NewMesh()

	res.Detour(vol.VerticesAddr, func() error {
		return vol.unpackVertices(res)
	})

	res.Detour(vol.IndicesAddr, func() error {
		return vol.unpackFaces(res)
	})

	return nil
}

func (vol *Volume) unpackVertices(res *resource.Container) error {
	for i := 0; i < int(vol.VertexCount); i++ {
		iVec := new(types.Vec3i)
		res.Parse(iVec)
		x := (float32(iVec[0]) * vol.ScaleFactor[0]) + vol.Offset[0]
		y := (float32(iVec[1]) * vol.ScaleFactor[1]) + vol.Offset[1]
		z := (float32(iVec[2]) * vol.ScaleFactor[2]) + vol.Offset[2]
		v := mathgl.Vec4f{x, y, z, 1.0}
		vol.AddVert4f(v)
	}

	return nil
}

func (vol *Volume) unpackFaces(res *resource.Container) error {
	fixIndex := func(c uint16) uint16 {
		c &= 0x7FFF
		return c
	}

	for i := 0; i < int(vol.IndexCount); i++ {
		var polygonType, junk uint16
		res.Parse(&junk)
		res.Parse(&polygonType)
		idxValues := make([]uint16, 6)
		res.Parse(idxValues)

		debug := func(junk, polyType uint16, idxValues []uint16) {
			fmt.Printf("%.4x ", junk)
			fmt.Printf("%.4x ", polyType)
			for _, i := range idxValues {
				fmt.Printf("%.4x ", i)
			}
			fmt.Printf("\n")
		}

		polygonType &= 0xF
		if polygonType == 0x4 {
			//			debug(junk, polygonType, idxValues)
		} else if polygonType == 0x3 || polygonType == 0xB {
			/* cube, 4 points specified */
			a := fixIndex(idxValues[0])
			b := fixIndex(idxValues[1])
			c := fixIndex(idxValues[2])
			d := fixIndex(idxValues[3])
			buildCube(vol.Mesh, a, b, c, d)
		} else if polygonType == 0x2 {
			/* seems to be two points */
			a := fixIndex(junk)
			b := fixIndex(idxValues[2])
			traceVerts(vol.Mesh, a, b)
		} else if polygonType == 0x1 {
			//			debug(junk, polygonType, idxValues)
		} else if polygonType == 0x0 || polygonType == 0x8 { /* it's probably a triangle */
			a := fixIndex(idxValues[0])
			b := fixIndex(idxValues[1])
			c := fixIndex(idxValues[2])
			vol.AddFace(types.Tri{
				A: a,
				B: b,
				C: c,
			})
		} else {
			debug(junk, polygonType, idxValues)
		}
	}
	return nil
}

func traceVerts(mesh *export.Mesh, verts ...uint16) {
	if len(verts) < 2 {
		panic("not enough verts to trace")
	}

	for i := 1; i < len(verts); i++ {
		mesh.AddVert(mesh.Vertices[verts[i]])
		mesh.AddVert(mesh.Vertices[verts[i]])
		mesh.AddVert(mesh.Vertices[verts[i-1]])
		mesh.AddFace(types.Tri{mesh.Rel(-1), mesh.Rel(-2), mesh.Rel(-3)})
	}
}
