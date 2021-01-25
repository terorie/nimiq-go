package tree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNibbles_String(t *testing.T) {
	assert.Equal(t, "", Nibbles{}.String())
	assert.Equal(t, "0", Nibbles{0x00}.String())
	assert.Equal(t, "4f302", Nibbles{0x04, 0x0F, 0x03, 0x00, 0x02}.String())
}

func TestKeyToNibbles(t *testing.T) {
	assert.Equal(t, KeyToNibbles(&[20]byte{}).String(),
		"0000000000000000000000000000000000000000")
	assert.Equal(t, KeyToNibbles(&[20]byte{0x0c}).String(),
		"0c00000000000000000000000000000000000000")
	assert.Equal(t, KeyToNibbles(&[20]byte{0x12, 0x03, 0x30}).String(),
		"1203300000000000000000000000000000000000")
}
