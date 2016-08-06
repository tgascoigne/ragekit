package export

type Exportable interface {
	GetName() string
	GetModels() []*Model
}

type ModelGroup struct {
	Name   string
	Models []*Model
}

func NewModelGroup() *ModelGroup {
	return &ModelGroup{
		Name:   "",
		Models: make([]*Model, 0),
	}
}

func (group *ModelGroup) GetName() string {
	return group.Name
}

func (group *ModelGroup) GetModels() []*Model {
	return group.Models
}

func (group *ModelGroup) Add(model *Model) {
	group.Models = append(group.Models, model)
}

func (group *ModelGroup) Merge(other Exportable) {
	group.Models = append(group.Models, other.GetModels()...)
	if group.Name == "" && other.GetName() != "" {
		group.Name = other.GetName()
	}
}

type Model struct {
	Name      string
	Meshes    []*Mesh
	Materials []*Material
	Extra     interface{}
}

func NewModel() *Model {
	return &Model{
		Meshes: make([]*Mesh, 0),
	}
}

func (model *Model) AddMesh(mesh *Mesh) {
	model.Meshes = append(model.Meshes, mesh)
}

func (model *Model) AddMaterial(material *Material) {
	model.Materials = append(model.Materials, material)
}

func (model *Model) GetName() string {
	return model.Name
}

func (model *Model) GetModels() []*Model {
	return []*Model{model}
}
