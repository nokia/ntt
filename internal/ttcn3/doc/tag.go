package doc

import (
	"bufio"
	"regexp"
	"strings"
)

var (
	tagRegex = regexp.MustCompile(`^[/*\s]*(@[A-Za-z0-9_]+)\s*:?\s*(.*?)[/*\r\n\s]*$`)
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

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		if s := FindTag(scanner.Text()); s != nil {
			res = append(res, s)
		}
	}

	return res
}
