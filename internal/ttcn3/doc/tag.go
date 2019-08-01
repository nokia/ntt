package doc

import (
	"regexp"
)

var (
	tagRegex = regexp.MustCompile(`(?m)^[/*\s]*(@[A-Za-z0-9_]+)\s*:?\s*(.*?)[/*\s\r\n]*$`)
)

// Finds first tag in s. A Return value of nil indicates not match.
func FindTag(s string) []string {
	res := tagRegex.FindStringSubmatch(s)
	if len(res) == 3 {
		return res[1:]
	}
	return nil
}

// Finds all tags in s.
func FindAllTags(s string) [][]string {
	var res [][]string
	for _, v := range tagRegex.FindAllStringSubmatch(s, -1) {
		if len(v) == 3 {
			res = append(res, v[1:])
		}
	}
	return res
}
