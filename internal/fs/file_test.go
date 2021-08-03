package fs_test

import (
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/stretchr/testify/assert"
)

// When acting as a language server, open files must not come from disk but from
// language server client.
//
// This test checks if content of files and file names are handled correctly.
func TestFiles(t *testing.T) {
	f1 := fs.Open("../fs/file_test.go")
	f2 := fs.Open("../../internal/fs/file_test.go")
	assert.NotNil(t, f1)
	assert.NotNil(t, f2)
	assert.Equal(t, f1, f2)

	f1 = fs.Open("../fs/does_not_exist")
	f2 = fs.Open("../../internal/fs/does_not_exist")
	assert.NotNil(t, f1)
	assert.NotNil(t, f2)
	assert.Equal(t, f1, f2)

	f1 = fs.Open("../fs/file_test.go")
	f2 = fs.Open("../fs/does_not_exist")
	assert.NotNil(t, f1)
	assert.NotNil(t, f2)
	assert.NotEqual(t, f1, f2)
}

func TestFileContent(t *testing.T) {
	t.Run("DiskRead", func(t *testing.T) {
		f1 := fs.Open("../fs/file_test.go")
		b1, err := f1.Bytes()
		assert.Nil(t, err)
		if len(b1) >= 16 {
			assert.Equal(t, []byte("package fs_test"), b1[:15])
		} else {
			t.Errorf("length of b1 < 16")
		}
	})

	t.Run("DiskRead2", func(t *testing.T) {
		f1 := fs.Open("../fs/file_test.go")
		f1.SetBytes([]byte("fnord"))
		b1, err := f1.Bytes()
		assert.Nil(t, err)
		assert.Equal(t, []byte("fnord"), b1)
	})

	t.Run("DiskRead3", func(t *testing.T) {
		f1 := fs.Open("../fs/does_not_exist")
		_, err := f1.Bytes()
		assert.NotNil(t, err)
		f1.SetBytes([]byte("fnord"))
		b1, err := f1.Bytes()
		assert.Nil(t, err)
		assert.Equal(t, []byte("fnord"), b1)
	})

	t.Run("DiskReadError", func(t *testing.T) {
		f1 := fs.Open("../fs")
		_, err := f1.Bytes()
		assert.NotNil(t, err)
	})

	t.Run("DiskReadError2", func(t *testing.T) {
		f1 := fs.Open("../fs/does_not_exist")
		f1.Reset()
		_, err := f1.Bytes()
		assert.NotNil(t, err)
	})

}
