package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFile(t *testing.T) {
	assert.Equal(t, "../fs/filesystem.go", FindFile("..", "fs", "filesystem.go"))
	assert.Equal(t, "", FindFile("..", "fs", "nothing to see here"))
}

func TestFindFiles(t *testing.T) {
	lst, err := FindFiles(".", func(n string) bool { return n == "filesystem.go" })
	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"filesystem.go"}, lst)

	lst, err = FindFiles(".", func(n string) bool { return false })
	assert.Equal(t, nil, err)
	assert.Equal(t, []string{}, lst)

	lst, err = FindFiles("${HOME}", func(n string) bool { return true })
	assert.Equal(t, true, err != nil)
	assert.Equal(t, []string(nil), lst)
}

func TestBasename(t *testing.T) {
	assert.Equal(t, "fs", Basename(``))
	assert.Equal(t, "fs", Basename(`.`))
	assert.Equal(t, "fs", Basename(`../../internal/fs`))
	assert.Equal(t, "internal", Basename(`../.`))
	assert.Equal(t, "internal", Basename(`..`))
	assert.Equal(t, "bar", Basename(`bar`))
	assert.Equal(t, "Bar", Basename(`Bar`))
	assert.Equal(t, "${HOME}", Basename(`${HOME}`))
	assert.Equal(t, "bar", Basename(`../../foo/bar`))
	assert.Equal(t, "/", Basename(`/`))
	assert.Equal(t, "filesystem.go", Basename(`../fs/filesystem.go`))
	assert.Equal(t, "filesystem.go", Basename(`../fs/filesystem.go/`))
}
