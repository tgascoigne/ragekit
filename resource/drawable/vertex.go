package drawable

import (
	"bytes"
	"encoding/binary"

	"github.com/tgascoigne/ragekit/cmd/rage-model-export/export"
	"github.com/tgascoigne/ragekit/resource"
	"github.com/tgascoigne/ragekit/resource/types"
)

type Vertex struct {
	/* This is the only information we can make use of for now */
	types.WorldCoord          /* Position: Vertex.[X,Y,Z,W] */
	UV0, UV1         types.UV /* UV: Vertex.[U,V] */
	Colour           uint32
}

func (vert *Vertex) Unpack(res *resource.Container, buf *VertexBuffer) error {
	buffer := make([]byte, buf.Stride)
	reader := bytes.NewReader(buffer)

	/* Read the vertex into our local buffer */
	if size, err := res.Read(buffer); uint16(size) != buf.Stride || err != nil {
		return err
	}

	offset := 0

	/* Parse out the info we can */
	if buf.Format.Has(export.VertXYZ) {
		if err := binary.Read(reader, binary.BigEndian, &vert.WorldCoord); err != nil {
			return err
		}
		offset += (4 * 3)
	}

	if buf.Format.Has(export.VertUnkA) {
		junk := make([]byte, 4)
		reader.Read(junk)
		offset += 4
	}

	if buf.Format.Has(export.VertUnkB) {
		junk := make([]byte, 4)
		reader.Read(junk)
		offset += 4
	}

	if buf.Format.Has(export.VertUnkC) {
		junk := make([]byte, 4)
		reader.Read(junk)
		offset += 4
	}

	if buf.Format.Has(export.VertColour) {
		if err := binary.Read(reader, binary.BigEndian, &vert.Colour); err != nil {
			return err
		}
		offset += 4
	}

	if buf.Format.Has(export.VertUnkD) {
		junk := make([]byte, 4)
		reader.Read(junk)
		offset += 4
	}

	if buf.Format.Has(export.VertUV0) {
		if err := binary.Read(reader, binary.BigEndian, &vert.UV0); err != nil {
			return err
		}
		offset += 4
	}

	if buf.Format.Has(export.VertUV1) {
		if err := binary.Read(reader, binary.BigEndian, &vert.UV1); err != nil {
			return err
		}
		offset += 4
	}

	if buf.Format.Has(export.VertUnkX) {
		junk := make([]byte, 4)
		reader.Read(junk)
		offset += 4
	}
	return nil
}
