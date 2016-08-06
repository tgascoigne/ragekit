package bounds

import (
	"github.com/Jragonmiris/mathgl"

	"github.com/tgascoigne/ragekit/cmd/rage-modelexport/export"
	"github.com/tgascoigne/ragekit/resource/types"
)

func buildCube(mesh *export.Mesh, a, b, c, d uint16) {
	A, B, C, D := mesh.Vertices[a].Pos, mesh.Vertices[b].Pos, mesh.Vertices[c].Pos, mesh.Vertices[d].Pos

	/* find the center point */
	center := A.Add(B).Add(C).Add(D).Mul(float32(0.25))

	findMidpoint := func(a, b mathgl.Vec4f) mathgl.Vec4f {
		return a.Add(b.Sub(a).Mul(0.5))
	}

	findMissing := func(a, b, c mathgl.Vec4f) mathgl.Vec4f {
		/* Find the midpoints */
		var midpoint [3]mathgl.Vec4f
		var vert mathgl.Vec4f
		midpoint[0] = findMidpoint(a, b)
		midpoint[1] = findMidpoint(b, c)
		midpoint[2] = findMidpoint(a, c)

		/* Find the vertex */
		for _, mid := range midpoint {
			vert = vert.Add(mid.Sub(center))
		}
		vert = vert.Add(center)
		return vert
	}

	/* create the missing vertices */
	E := findMissing(A, B, C)
	F := findMissing(A, D, B)
	G := findMissing(A, C, D)
	H := findMissing(B, D, C)

	/* Create the faces */
	mesh.AddVert4f(H)
	mesh.AddVert4f(G)
	mesh.AddVert4f(F)
	mesh.AddVert4f(E)
	mesh.AddVert4f(D)
	mesh.AddVert4f(C)
	mesh.AddVert4f(B)
	mesh.AddVert4f(A)

	a, b, c, d = mesh.Rel(-1), mesh.Rel(-2), mesh.Rel(-3), mesh.Rel(-4)
	e, f, g, h := mesh.Rel(-5), mesh.Rel(-6), mesh.Rel(-7), mesh.Rel(-8)

	//traceVerts(mesh, a, b, c, d, e, f, g, h)

	mesh.AddFace(types.Tri{a, f, e})
	mesh.AddFace(types.Tri{e, f, b})

	mesh.AddFace(types.Tri{f, d, b})
	mesh.AddFace(types.Tri{d, h, b})

	mesh.AddFace(types.Tri{d, g, h})
	mesh.AddFace(types.Tri{g, c, h})

	mesh.AddFace(types.Tri{g, a, c})
	mesh.AddFace(types.Tri{c, a, e})

	mesh.AddFace(types.Tri{b, e, c})
	mesh.AddFace(types.Tri{b, c, h})

	mesh.AddFace(types.Tri{a, f, d})
	mesh.AddFace(types.Tri{a, d, g})

}
