package obj

import (
	"fmt"

	"github.com/tgascoigne/ragekit/cmd/rage-modelexport/export"
)

func ExportModel(ctx *Context, model *export.Model) error {
	materialNames := make([]string, len(model.Materials))
	modelName := ctx.Unique(model.Name)

	for i, material := range model.Materials {
		materialNames[i] = ctx.Unique(model.Name)
		fmt.Fprintf(ctx.MtlFile, "newmtl %v\n", materialNames[i])
		if material.DiffBitmap != "" {
			fmt.Fprintf(ctx.MtlFile, "map_Kd %v\n", material.DiffBitmap)
		}
	}

	fmt.Fprintf(ctx.ObjFile, "o %v\n", modelName)
	for _, mesh := range model.Meshes {
		fmt.Fprintf(ctx.ObjFile, "g %v\n", ctx.Unique(model.Name))
		if mesh.Material != -1 {
			fmt.Fprintf(ctx.ObjFile, "usemtl %v\n", materialNames[mesh.Material])
		}

		for _, vert := range mesh.Vertices {
			x, y, z := vert.Pos[0], vert.Pos[1], vert.Pos[2]
			u, v := vert.UV[0], vert.UV[1]
			if export.FlipYZ {
				y, z = z, y
			}

			fmt.Fprintf(ctx.ObjFile, "v %v %v %v\n", x, y, z)
			fmt.Fprintf(ctx.ObjFile, "vt %v %v\n", u, v)
		}

		numVerts := len(mesh.Vertices)

		for _, face := range mesh.Faces {
			a, b, c := -int(numVerts-int(face.A)), -int(numVerts-int(face.B)), -int(numVerts-int(face.C))
			fmt.Fprintf(ctx.ObjFile, "f %v/%v %v/%v %v/%v\n", a, a, b, b, c, c)
		}
	}
	return nil
}
