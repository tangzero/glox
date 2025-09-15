package glox

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner_ScanTokens(t *testing.T) {
	source := `var language = "lox";`
	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()

	require.NoError(t, err)
	assert.Len(t, tokens, 5)
}
