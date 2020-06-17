package t3xf

type t3xfError string

func (e t3xfError) Error() string {
	return string(e)
}

const (
	// ErrUnkownFormat means that the t3xf format or version could not be detected.
	ErrUnknownFormat = t3xfError("unknown format")

	// ErrDeprecatedWIDEN means that a t3xf code contains a deprecated
	// WIDEN instruction.
	ErrDeprecatedWIDEN = t3xfError("deprecated WIDEN instruction")

	// ErrDeprecatedWIDEN means that a t3xf code contains a deprecated
	// WIDEN instruction.
	//
	// ESWAP was introduced to support running little-endian t3xf files
	// on big-endian machines. We don't support that anymore.
	ErrDeprecatedESWAP = t3xfError("deprecated ESWAP instruction")

	// ErrUnmatchedBLOCk means a SCAN instruction had no matching BLOCK
	// instruction.
	ErrUnmatchedBLOCK = t3xfError("unmatched BLOCK instruction.")

	// ErrUnmatchedSCAN means a BLOCK instruction had no matching SCAN
	// instruction.
	ErrUnmatchedSCAN = t3xfError("unmatched SCAN instruction.")
)
