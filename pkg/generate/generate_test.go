package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKeyData(t *testing.T) {
	result, _ := generateKeyData("foo", 1)
	assert.Equal(t, result, []byte("foo"))

	_, err := generateKeyData("", 0)
	assert.Equal(t, err.Error(), "key length must be an greater than zero")

	resultOne, _ := generateKeyData("", 6)
	resultTwo, _ := generateKeyData("", 6)
	assert.Equal(t, len(resultOne), 6)
	assert.NotEqual(t, resultOne, resultTwo)
}
