package data_structure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBloom_Exist(t *testing.T) {
	b := CreateBloomFilter(10, 0.01)
	b.Add("a")
	b.Add("b")
	assert.EqualValues(t, 10, b.NumberOfEntries)
	assert.EqualValues(t, 0.01, b.ErrorRate)
	assert.True(t, b.Exist("a"))
	assert.True(t, b.Exist("b"))
	assert.False(t, b.Exist("c"))
	assert.False(t, b.Exist("d"))
}

func TestBloom_CalcHash(t *testing.T) {
	b := CreateBloomFilter(10, 0.01)
	x := b.GenerateHashValue("abcdef")
	y := b.GenerateHashValue("abcdef")
	assert.EqualValues(t, x.a, y.a)
	assert.EqualValues(t, x.b, y.b)
}
