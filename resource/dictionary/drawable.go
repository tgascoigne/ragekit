package dictionary

import (
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/drawable"
)

type DrawableCollection struct {
	resource.Collection
	Drawables []*drawable.Drawable
}
