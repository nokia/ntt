package session

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tempStorage() (*storage, func()) {
	dir, err := ioutil.TempDir("", "storage.test")
	if err != nil {
		panic(err.Error())
	}
	return New(dir), func() {
		os.RemoveAll(dir)
	}
}

func TestEmptyStorage(t *testing.T) {
	s := New(".")
	list, err := s.Sessions()
	assert.Equal(t, 0, len(list))
	assert.Nil(t, err)

}

func TestDeepStorage(t *testing.T) {
	s := New("/tmp/ntt/sessions/does-not-exist")
	list, err := s.Sessions()
	assert.Equal(t, 0, len(list))
	assert.Nil(t, err)
}

func TestInvalidStorage(t *testing.T) {
	s := New("storage_test.go")
	list, err := s.Sessions()
	assert.Nil(t, list)
	assert.NotNil(t, err)
}
