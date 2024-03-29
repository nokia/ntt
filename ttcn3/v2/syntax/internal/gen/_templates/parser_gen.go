// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at {{now}}

package syntax

{{ $data := . }}

// this is a minimal set of keywords required to parse TTCN-3.
// Note, that the standard specifies more keywords.
var keywords = map[string]Kind{
	{{range $k := .KeywordTokens}}"{{index $data.TokenMap $k}}": {{$k}},
	{{end}}
}

{{/* Generate the parser for the given productions. */}}
{{- range .Productions -}}
	{{if not (index $data.ImplementedProductions .Name.String)}}
		{{- text . -}}
		func (p *parser) parse{{.Name.String}}() bool {
			p.Push({{.Name.String}});
			{{- template "generate" .Expr -}}
			p.Pop();
			return true
		}

	{{  end -}}
{{- end -}}

{{/* Expect a token. */}}
{{- define "token" -}}
	{{- $tok := .String -}}
	{{- if eq $tok ";" -}}
	p.expectSemicolon();
	{{- else if eq $tok "," -}}
	p.expectComma();
	{{- else -}}
	p.expect({{kind .String}});
	{{- end -}}
{{- end -}}

{{/* Reference a named rule or token. */}}
{{- define "name" -}}
	{{if lexeme .String}}
		{{template "token" .}}
	{{else}}
		p.parse{{.String}}();
	{{end}}
{{- end -}}

{{/* Parse a sequence of rules */}}
{{- define "sequence" -}}
	{{range .}}{{template "generate" .}}{{end}}
{{- end -}}

{{/* Inline groupings. */}}
{{- define "group" -}}
	{{template "generate" .Body}}
{{- end -}}

{{/* Parse an optional rule */}}
{{- define "option" -}}
	if p.accept({{join (first .) ","}}) {
		{{- template "generate" .Body -}}
	};
{{- end -}}

{{- define "alternative" -}}
	switch p.next {
	{{- range . -}}
	case {{join (first .) ","}}:
		{{- template "generate" . -}}
	{{- end}}
	};
{{- end -}}

{{- define "repetition" -}}
	for p.accept({{join (first .) ","}}) {
		{{- template "generate" .Body -}}
	};
{{- end -}}

{{/* The generate template dispatches to the appropriate rule */}}
{{- define "generate" -}}
{{- $t := type . -}}
{{-      if eq $t "token"       -}}{{- template "token" .       -}}
{{- else if eq $t "name"        -}}{{- template "name" .        -}}
{{- else if eq $t "option"      -}}{{- template "option" .      -}}
{{- else if eq $t "sequence"    -}}{{- template "sequence" .    -}}
{{- else if eq $t "alternative" -}}{{- template "alternative" . -}}
{{- else if eq $t "repetition"  -}}{{- template "repetition" .  -}}
{{- else if eq $t "group"       -}}{{- template "group" .       -}}
{{- else -}}
    panic("missing template for {{$t}}");
{{- end -}}
{{- end -}}
