package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockReward(t *testing.T) {
	assert.Equal(t, uint64(440597534), BlockReward(1))
	assert.Equal(t, uint64(440597429), BlockReward(2))
	assert.Equal(t, uint64(440597324), BlockReward(3))
	assert.Equal(t, uint64(440492605), BlockReward(1000))
	assert.Equal(t, uint64(440072823), BlockReward(4999))
	assert.Equal(t, uint64(440072718), BlockReward(5000))
	assert.Equal(t, uint64(440072613), BlockReward(5001))
	assert.Equal(t, uint64(440072508), BlockReward(5002))
	assert.Equal(t, uint64(430217207), BlockReward(100000))
	assert.Equal(t, uint64(40607225), BlockReward(10000000))
	assert.Equal(t, uint64(4001), BlockReward(48692959))
	assert.Equal(t, uint64(4000), BlockReward(48692960))
	assert.Equal(t, uint64(4000), BlockReward(52888984))
	assert.Equal(t, uint64(0), BlockReward(52888985))
}
