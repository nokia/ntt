package runtime

type Module struct {
	name string
	file string
	tags [][]string

	Imports []*Import
}

func NewModule(name, file string, tags [][]string) *Module {
	return &Module{name: name, file: file, tags: tags}
}

func (m *Module) Tags() [][]string {
	return m.tags
}

func (m *Module) Name() string {
	return m.name
}

func (m *Module) File() string {
	return m.file
}

func (m *Module) AddImport(imp *Import) {
	m.Imports = append(m.Imports, imp)
}
