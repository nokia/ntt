package runtime

// ResetBuiltins sets the internal builtin map to nil.
func ResetBuiltins() {
	builtins = nil
}
