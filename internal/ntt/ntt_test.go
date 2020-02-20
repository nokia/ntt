package ntt_test

import (
	"testing"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

func TestId(t *testing.T) {
	suite := &ntt.Suite{}
	id, err := suite.Id()
	assert.Nil(t, err)
	assert.NotZero(t, id)
}
