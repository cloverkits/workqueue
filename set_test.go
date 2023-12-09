package workqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	s := make(set)
	assert.Equal(t, 0, s.len())
	assert.False(t, s.has("foo"))
	s.insert("foo")
	assert.Equal(t, 1, s.len())
	assert.True(t, s.has("foo"))
	s.delete("foo")
	assert.Equal(t, 0, s.len())
	assert.False(t, s.has("foo"))
}
