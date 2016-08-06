package export

type Material struct {
	DiffBitmap string
}

func NewMaterial() *Material {
	return &Material{}
}
