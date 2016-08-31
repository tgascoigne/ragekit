package dae

import (
	"bytes"
	"fmt"

	xmlx "github.com/jteeuwen/go-pkg-xmlx"
	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export"
)

func ExportGeometries(ctx *Context, model *export.Model) error {
	root := ctx.libGeometries
	modelName := ctx.Unique(model.Name)
	materialIds := model.Extra.([]string)
	sceneNode := addChild(ctx.scene, "node", Attribs{
		"id":   modelName,
		"name": modelName,
	}, "")

	nodeId := func(n *xmlx.Node) string {
		return n.As("", "id")
	}

	ref := func(n *xmlx.Node) string {
		return fmt.Sprintf("#%v", nodeId(n))
	}

	subType := func(base, suffix string) string {
		return fmt.Sprintf("%v_%v", base, suffix)
	}

	for _, objMesh := range model.Meshes {
		var posBuf, uvBuf, faceBuf, colourBuf bytes.Buffer
		meshName := ctx.Unique(model.Name)
		// <geometry>
		geometry := addChild(root, "geometry", Attribs{
			"id": meshName,
		}, "")
		// <mesh>
		mesh := addChild(geometry, "mesh", nil, "")

		/* Generate the vertex and face buffers */
		for _, vert := range objMesh.Vertices {
			if !objMesh.Format.Has(export.VertXYZ) {
				panic("Vert has no XYZ?")
			}
			x, y, z := vert.Pos[0], vert.Pos[1], vert.Pos[2]
			posBuf.WriteString(fmt.Sprintf("%v %v %v ", x, y, z))

			u, v := vert.UV[0], vert.UV[1]
			if export.FlipYZ {
				y, z = z, y
			}
			uvBuf.WriteString(fmt.Sprintf("%v %v ", u, v))

			colour := vert.Colour
			a := float32((colour&0xFF000000)>>24) / 255
			r := float32((colour&0x00FF0000)>>16) / 255
			g := float32((colour&0x0000FF00)>>8) / 255
			b := float32(colour&0x000000FF) / 255
			colourBuf.WriteString(fmt.Sprintf("%v %v %v %v ", r, g, b, a))
		}

		for _, face := range objMesh.Faces {
			a, b, c := int(face.A), int(face.B), int(face.C)
			faceBuf.WriteString(fmt.Sprintf("%v %v %v %v %v %v %v %v %v ", a, a, a, b, b, b, c, c, c))
		}

		// <source>
		posSource := addChild(mesh, "source", Attribs{"id": subType(meshName, "positions"),
			"name": "position"}, "")
		posArray := addChild(posSource, "float_array", Attribs{"id": subType(nodeId(posSource), "array"),
			"count": len(objMesh.Vertices) * 3}, posBuf.String())

		// <technique_common>
		format := addChild(posSource, "technique_common", nil, "")
		accessor := addChild(format, "accessor", Attribs{"count": len(objMesh.Vertices),
			"offset": 0, "source": ref(posArray), "stride": "3"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "X", "type": "float"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "Y", "type": "float"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "Z", "type": "float"}, "")

		// <source>
		uvSource := addChild(mesh, "source", Attribs{"id": subType(meshName, "uv"), "name": "uv0"}, "")
		uvArray := addChild(uvSource, "float_array", Attribs{"id": subType(nodeId(uvSource), "array"),
			"count": len(objMesh.Vertices) * 2}, uvBuf.String())

		// <technique_common>
		format = addChild(uvSource, "technique_common", nil, "")
		accessor = addChild(format, "accessor", Attribs{"count": len(objMesh.Vertices),
			"offset": 0, "source": ref(uvArray), "stride": "2"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "S", "type": "float"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "T", "type": "float"}, "")

		// <source>
		colourSource := addChild(mesh, "source", Attribs{"id": subType(meshName, "colours"),
			"name": "colour"}, "")
		colourArray := addChild(colourSource, "float_array", Attribs{"id": subType(nodeId(colourSource), "array"),
			"count": len(objMesh.Vertices) * 4}, colourBuf.String())

		// <technique_common>
		format = addChild(colourSource, "technique_common", nil, "")
		accessor = addChild(format, "accessor", Attribs{"count": len(objMesh.Vertices),
			"offset": 0, "source": ref(colourArray), "stride": "4"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "R", "type": "float"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "G", "type": "float"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "B", "type": "float"}, "")
		_ = addChild(accessor, "param", Attribs{"name": "A", "type": "float"}, "")

		// <vertices>
		vertices := addChild(mesh, "vertices", Attribs{"id": subType(meshName, "vertices")}, "")
		_ = addChild(vertices, "input", Attribs{"semantic": "POSITION", "source": ref(posSource)}, "")

		// <triangles>
		materialId := materialIds[objMesh.Material]
		materialInstId := subType(meshName, "material")
		triangles := addChild(mesh, "triangles", Attribs{"count": len(objMesh.Faces), "material": materialInstId}, "")
		_ = addChild(triangles, "input", Attribs{"offset": 0, "semantic": "VERTEX",
			"source": ref(vertices)}, "")
		_ = addChild(triangles, "input", Attribs{"offset": 1, "semantic": "TEXCOORD",
			"source": ref(uvSource)}, "")
		_ = addChild(triangles, "input", Attribs{"offset": 2, "semantic": "COLOR",
			"source": ref(colourSource)}, "")
		_ = addChild(triangles, "p", nil, faceBuf.String())

		// <instance_geometry>
		geomInst := addChild(sceneNode, "instance_geometry", Attribs{"url": ref(geometry)}, "")
		bindMaterial := addChild(geomInst, "bind_material", nil, "")
		technique := addChild(bindMaterial, "technique_common", nil, "")
		materialInst := addChild(technique, "instance_material", Attribs{"symbol": materialInstId, "target": fmt.Sprintf("#%v", materialId)}, "")
		_ = addChild(materialInst, "bind_vertex_input", Attribs{"semantic": "UV0", "input_semantic": "TEXCOORD"}, "")
	}

	return nil
}
