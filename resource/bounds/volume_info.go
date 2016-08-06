package bounds

import (
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type VolumeInfoHeader struct {
	Low  types.Vec4
	High types.Vec4
}

type VolumeInfo struct {
	VolumeInfoHeader
}

func (info *VolumeInfo) Unpack(res *resource.Container) error {
	res.Parse(&info.VolumeInfoHeader)

	return nil
}
