package resource

import (
	"log"

	"github.com/tgascoigne/ragekit/resource/types"
)

type PointerCollection struct {
	Addr     types.Ptr32
	Count    uint16
	Capacity uint16
}

func (col *PointerCollection) Detour(res *Container, i int, callback func() error) error {
	addr, err := col.GetPtr(res, i)
	if err != nil {
		return err
	}

	return res.Detour(addr, callback)
}

func (col *PointerCollection) For(res *Container, callback func(i int) error) error {
	for i := 0; i < int(col.Count); i++ {
		if err := col.Detour(res, i, func() error {
			return callback(i)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (col *PointerCollection) JumpTo(res *Container, i int) error {
	addr, err := col.GetPtr(res, i)
	if err != nil {
		return err
	}

	if err = res.Jump(addr); err != nil {
		log.Printf("Error performing collection lookup")
		return err
	}

	return nil
}

func (col *PointerCollection) GetPtr(res *Container, i int) (types.Ptr32, error) {
	var addr types.Ptr32
	if err := res.PeekElem(col.Addr, i, &addr); err != nil {
		log.Printf("Error performing collection lookup")
		return 0, err
	}
	return addr, nil
}
