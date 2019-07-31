package syntax

func WalkModuleDefs(fun func(def Node) bool, nodes ...Node) {
	for _, n := range nodes {
		switch x := n.(type) {
		case *Module:
			walkModuleDefs(fun, x.Defs...)
		case *GroupDecl:
			walkModuleDefs(fun, x.Defs...)
		case *ImportDecl,
			*FriendDecl,
			*SubTypeDecl,
			*PortTypeDecl,
			*ComponentTypeDecl,
			*StructTypeDecl,
			*EnumTypeDecl,
			*BehaviourTypeDecl,
			*TemplateDecl,
			*ModuleParameterGroup,
			*ValueDecl,
			*SignatureDecl,
			*FuncDecl,
			*ControlPart:
			if !fun(x) {
				return
			}
		}
	}
}

func walkModuleDefs(fun func(def Node) bool, defs ...*ModuleDef) {
	for _, d := range defs {
		WalkModuleDefs(fun, d)
	}
}
