package dae

import (
	"fmt"
	"strings"

	xmlx "github.com/gwitmond/go-pkg-xmlx"
	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export"
)

func createImage(ctx *Context, path string) string {
	images := ctx.libImages

	if image, ok := ctx.imageIds[path]; ok {
		return image
	}

	imageId := strings.Replace(path, ".", "_", -1)
	ctx.imageIds[path] = imageId
	image := addChild(images, "image", Attribs{"id": imageId}, "")
	_ = addChild(image, "init_from", nil, path)
	return createImage(ctx, path)
}

func ExportMaterials(ctx *Context, model *export.Model) error {
	libMaterials := ctx.libMaterials
	libEffects := ctx.libEffects

	nodeId := func(n *xmlx.Node) string {
		return n.As("", "id")
	}

	ref := func(n *xmlx.Node) string {
		return fmt.Sprintf("#%v", nodeId(n))
	}

	subType := func(base, suffix string) string {
		return fmt.Sprintf("%v_%v", base, suffix)
	}

	materialIds := make([]string, 0)

	for _, material := range model.Materials {
		materialName := ctx.Unique(model.Name)
		diffusePath := material.DiffBitmap
		image := createImage(ctx, diffusePath)

		// <effect>
		effectName := subType(materialName, "effect")
		effect := addChild(libEffects, "effect", Attribs{"id": effectName}, "")
		profile := addChild(effect, "profile_COMMON", nil, "")
		technique := addChild(profile, "technique", Attribs{"sid": "standard"}, "")

		// <newparam>
		surfaceParamName := subType(materialName, "surface")
		param := addChild(technique, "newparam", Attribs{"sid": surfaceParamName}, "")
		surface := addChild(param, "surface", Attribs{"type": "2D"}, "")
		_ = addChild(surface, "init_from", nil, image)

		// <newparam>
		texParamName := subType(materialName, "texture")
		param = addChild(technique, "newparam", Attribs{"sid": texParamName}, "")
		sampler := addChild(param, "sampler2D", nil, "")
		_ = addChild(sampler, "source", nil, surfaceParamName)

		// <technique>
		lambert := addChild(technique, "lambert", nil, "")
		diffuse := addChild(lambert, "diffuse", nil, "")
		_ = addChild(diffuse, "texture", Attribs{"texture": texParamName, "texcoord": "UV0"}, "")

		// <material>
		material := addChild(libMaterials, "material", Attribs{"id": materialName}, "")
		_ = addChild(material, "instance_effect", Attribs{"url": ref(effect)}, "")

		materialIds = append(materialIds, materialName)
	}

	model.Extra = materialIds
	return nil
}
