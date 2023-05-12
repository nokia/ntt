package lnav

type Spec struct {
	Schema  string            `json:"$schema"`
	Formats map[string]Format `json:"inline"`
}

type Format struct {
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	URL             []string          `json:"url"`
	Regex           map[string]Regex  `json:"regex"`
	TimestampFormat []string          `json:"timestamp_format"`
	OrderedByTime   bool              `json:"ordered_by_time"`
	OPidField       string            `json:"opid_field"`
	LevelField      string            `json:"level_field"`
	Level           map[string]string `json:"level"`
	Value           map[string]Value  `json:"value"`
}

type Regex struct {
	Pattern string `json:"pattern"`
}

type Value struct {
	Kind        string `json:"kind"`
	Identifier  bool   `json:"identifier"`
	Hidden      bool   `json:"hidden,omitempty"`
	Description string `json:"description,omitempty"`
}
