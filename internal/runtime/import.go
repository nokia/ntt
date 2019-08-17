package runtime

type Import struct {
	mod      *Module
	imported string
	tags     [][]string
}

func NewImport(mod *Module, imported string, tags [][]string) *Import {
	return &Import{
		mod:      mod,
		imported: imported,
		tags:     tags,
	}
}

func (imp *Import) Module() *Module {
	return imp.mod
}

func (imp *Import) Tags() [][]string {
	return imp.tags
}

func (imp *Import) ImportedModule() string {
	return imp.imported
}
