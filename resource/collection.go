package resource

import (
	"github.com/tgascoigne/ragekit/resource/types"
)

type Collection struct {
	Addr     types.Ptr32
	Count    uint16
	Capacity uint16
}

func (col *Collection) Detour(res *Container, callback func() error) error {
	return res.Detour(col.Addr, callback)
}

func (col *Collection) For(res *Container, callback func(i int) error) error {
	return col.Detour(res, func() error {
		for i := 0; i < int(col.Count); i++ {
			if err := callback(i); err != nil {
				return err
			}
		}
		return nil
	})
}
