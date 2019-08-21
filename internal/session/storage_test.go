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
	s, _ := New(dir)
	return s, func() {
		os.RemoveAll(dir)
	}
}

func TestEmptyStorage(t *testing.T) {
	s, _ := New(".")
	list, err := s.Sessions()
	assert.Equal(t, 0, len(list))
	assert.Nil(t, err)

}

func TestDeepStorage(t *testing.T) {
	s, _ := New("/tmp/ntt/sessions/does-not-exist")
	list, err := s.Sessions()
	assert.Equal(t, 0, len(list))
	assert.Nil(t, err)
}

func TestInvalidStorage(t *testing.T) {
	_, err := New("storage_test.go")
	assert.NotNil(t, err)
}
