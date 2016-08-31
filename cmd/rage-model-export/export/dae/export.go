package dae

import (
	"fmt"
	"os"

	xmlx "github.com/jteeuwen/go-pkg-xmlx"

	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export"
)

type Attribs map[string]interface{}

type Context struct {
	DaeFile       *os.File
	NextID        int
	Object        *xmlx.Document
	imageIds      map[string]string
	libImages     *xmlx.Node
	libMaterials  *xmlx.Node
	libEffects    *xmlx.Node
	libGeometries *xmlx.Node
	libScenes     *xmlx.Node
	scene         *xmlx.Node
}

func addChild(node *xmlx.Node, name string, attribs Attribs, contents string) *xmlx.Node {
	child := xmlx.NewNode(xmlx.NT_ELEMENT)
	child.Name.Local = name
	child.Value = contents

	for key, val := range attribs {
		child.SetAttr(key, fmt.Sprintf("%v", val))
	}

	node.AddChild(child)

	return child
}

func Export(object export.Exportable) error {
	outFile := fmt.Sprintf("%v.dae", object.GetName())
	ctx := Context{
		Object:   xmlx.New(),
		imageIds: make(map[string]string),
	}

	ctx.Object.Root = xmlx.NewNode(xmlx.NT_ROOT)
	root := ctx.Object.Root

	collada := addChild(root, "COLLADA",
		Attribs{"xmlns": "http://www.collada.org/2005/11/COLLADASchema", "version": "1.4.1"}, "")

	asset := addChild(collada, "asset", nil, "")
	_ = addChild(asset, "unit", Attribs{"meter": "1", "name": "meter"}, "")
	_ = addChild(asset, "up_axis", nil, "Y_UP")

	ctx.libImages = addChild(collada, "library_images", nil, "")
	ctx.libMaterials = addChild(collada, "library_materials", nil, "")
	ctx.libEffects = addChild(collada, "library_effects", nil, "")
	ctx.libGeometries = addChild(collada, "library_geometries", nil, "")

	ctx.libScenes = addChild(collada, "library_visual_scenes", nil, "")
	ctx.scene = addChild(ctx.libScenes, "visual_scene", Attribs{"id": "Scene",
		"name": object.GetName()}, "")

	scenes := addChild(collada, "scene", nil, "")
	_ = addChild(scenes, "instance_visual_scene", Attribs{"url": "#Scene"}, "")

	for _, model := range object.GetModels() {
		var err error
		if err = ExportMaterials(&ctx, model); err != nil {
			return err
		}
		if err = ExportGeometries(&ctx, model); err != nil {
			return err
		}
	}

	if err := ctx.Object.SaveFile(outFile); err != nil {
		return err
	}

	return nil
}

func (ctx *Context) Unique(name string) string {
	ctx.NextID++
	return fmt.Sprintf("%v_%.4d", name, ctx.NextID-1)
}
