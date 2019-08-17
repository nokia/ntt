package runtime

type Test struct {
	mod  *Module
	tags [][]string
	name string
}

func NewTest(mod *Module, name string, tags [][]string) *Test {
	return &Test{mod, tags, name}
}

func (t *Test) Module() *Module {
	return t.mod
}

func (t *Test) Tags() [][]string {
	return t.tags
}

func (t *Test) Name() string {
	return t.name
}

func (t *Test) FullName() string {
	return t.Module().Name() + "." + t.name
}
