package bounds

import (
	"github.com/tgascoigne/ragekit/cmd/rage-modelexport/export"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type NodesHeader struct {
	_            uint32
	_            types.Ptr32
	_            uint32
	_            uint32
	WorldCoords  [4]types.Vec4
	BoundingBox  types.Vec4
	BoundsTable  types.Ptr32
	VolumeMatrix types.Ptr32
	_            types.Ptr32
	VolumeInfo   types.Ptr32
	_            types.Ptr32
	_            types.Ptr32
	Count        uint16
	Capacity     uint16
	_            types.Ptr32
}

type Nodes struct {
	NodesHeader
	Volumes []*Volume
	Model   *export.Model
}

func (nodes *Nodes) Unpack(res *resource.Container) error {
	res.Parse(&nodes.NodesHeader)

	var err error

	nodes.Volumes = make([]*Volume, nodes.Capacity)
	volCollection := resource.PointerCollection{
		Addr:     nodes.BoundsTable,
		Count:    nodes.Count,
		Capacity: nodes.Capacity,
	}

	volInfoCollection := resource.Collection{
		Addr:     nodes.VolumeInfo,
		Count:    nodes.Count,
		Capacity: nodes.Capacity,
	}

	err = volInfoCollection.For(res, func(i int) error {
		nodes.Volumes[i] = new(Volume)
		if err := nodes.Volumes[i].VolumeInfo.Unpack(res); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = volCollection.For(res, func(i int) error {
		return nodes.Volumes[i].Unpack(res)
	})
	if err != nil {
		return err
	}

	nodes.Model = export.NewModel()
	for _, vol := range nodes.Volumes {
		nodes.Model.AddMesh(vol.Mesh)
	}

	return nil
}
