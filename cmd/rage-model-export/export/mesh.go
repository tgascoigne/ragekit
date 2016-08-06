package export

import (
	"fmt"

	"github.com/Jragonmiris/mathgl"

	"github.com/tgascoigne/ragekit/resource/types"
)

type Vertex struct {
	Pos    mathgl.Vec4f
	UV     mathgl.Vec2f
	Colour uint32
}

type Mesh struct {
	Format   VertexFormat
	Vertices []Vertex
	Faces    []types.Tri
	Material int
}

func NewMesh() *Mesh {
	return &Mesh{
		Vertices: make([]Vertex, 0),
		Faces:    make([]types.Tri, 0),
		Material: -1,
	}
}

func (mesh *Mesh) AddVert(pos Vertex) {
	mesh.Vertices = append(mesh.Vertices, pos)
}

func (mesh *Mesh) AddVert4f(pos mathgl.Vec4f) {
	v := Vertex{
		Pos: pos,
	}
	mesh.AddVert(v)
}

func (mesh *Mesh) Rel(idx int) uint16 {
	numVerts := len(mesh.Vertices)
	return uint16(numVerts + idx)
}

func (mesh *Mesh) AddFace(face types.Tri) {
	max := func(i ...uint16) int {
		max := i[0]
		for _, x := range i {
			if x > max {
				max = x
			}
		}
		return int(max)
	}
	if max(face.A, face.B, face.C) > len(mesh.Vertices) {
		panic(fmt.Sprintf("invalid vert reference: %v", face))
	}
	mesh.Faces = append(mesh.Faces, face)
}
