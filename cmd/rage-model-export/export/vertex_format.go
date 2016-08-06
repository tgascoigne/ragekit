package export

import (
	"fmt"
)

type VertexFormat uint32

const (
	VertXYZ    = (1 << 0)  /* + 4*3 */
	VertUnkA   = (1 << 1)  /* + 4 */
	VertUnkB   = (1 << 2)  /* + 4 */
	VertUnkC   = (1 << 3)  /* + 4 */
	VertColour = (1 << 4)  /* + 4 */
	VertUnkD   = (1 << 5)  /* + 4 */
	VertUV0    = (1 << 6)  /* + 4 */
	VertUV1    = (1 << 7)  /* + 4 */
	VertUnkX   = (1 << 14) /* + 4 */
)

func (f VertexFormat) Has(field int) bool {
	return (int(f) & field) != 0
}

func (f VertexFormat) String() string {
	return fmt.Sprintf("0x%x", int(f)) /* todo: better representation. */
}
