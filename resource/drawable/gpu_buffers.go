package drawable

import (
	"bytes"
	"encoding/binary"

	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type VertexHeader struct {
	_      uint32 /* vtable */
	Stride uint16
	_      uint16
	Buffer types.Ptr32
	Count  uint32
	_      types.Ptr32 /* Buffer */
	_      uint32
	Info   types.Ptr32
	_      types.Ptr32
}

type VertexInfo struct {
	Format export.VertexFormat
	_      uint16 /* Stride */
	_      uint16 /* correlated with Stride */
	_      uint32 /* usually 0xAA111111 */
	_      uint32 /* usually 0x1199a996 */
}

type VertexBuffer struct {
	VertexHeader
	VertexInfo
	Vertex []*Vertex
}

func (buf *VertexBuffer) Unpack(res *resource.Container) error {
	res.Parse(&buf.VertexHeader)

	if err := res.Detour(buf.Info, func() error {
		res.Parse(&buf.VertexInfo)
		return nil
	}); err != nil {
		return err
	}

	buf.Vertex = make([]*Vertex, buf.Count)
	for i := range buf.Vertex {
		buf.Vertex[i] = new(Vertex)
	}

	if err := res.Detour(buf.Buffer, func() error {
		for _, vert := range buf.Vertex {
			if err := vert.Unpack(res, buf); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

type IndexHeader struct {
	_      uint32 /* vtable */
	Count  uint32
	Buffer types.Ptr32
	Info   types.Ptr32
}

type IndexBuffer struct {
	IndexHeader
	Index  []*types.Tri
	Stride int /* todo: is this referenced in the geom? */
}

func (buf *IndexBuffer) Unpack(res *resource.Container) error {
	buf.Stride = 3 * 2 // 3*uint16 /* is this stored anywhere? */
	res.Parse(&buf.IndexHeader)

	buf.Index = make([]*types.Tri, buf.Count/3)
	for i := range buf.Index {
		buf.Index[i] = new(types.Tri)
	}

	if err := res.Detour(buf.Buffer, func() error {
		buffer := make([]byte, buf.Stride)
		reader := bytes.NewReader(buffer)

		for _, idx := range buf.Index {
			/* Read the index into our local buffer */
			if size, err := res.Read(buffer); size != buf.Stride || err != nil {
				return err
			}

			/* Parse out the info we can */
			if err := binary.Read(reader, binary.BigEndian, idx); err != nil {
				return err
			}
			reader.Seek(0, 0)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
