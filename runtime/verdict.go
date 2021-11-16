package runtime

type Verdict string

func (v Verdict) Type() ObjectType { return VERDICT }
func (v Verdict) Inspect() string  { return string(v) }

const (
	NoneVerdict   Verdict = "none"
	PassVerdict   Verdict = "pass"
	InconcVerdict Verdict = "inconc"
	FailVerdict   Verdict = "fail"
	ErrorVerdict  Verdict = "error"
)
