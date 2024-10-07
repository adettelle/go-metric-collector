package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRangeChunks(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6, 7}

	res := RangeChunks(3, data)
	expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}

	require.Equal(t, expected, res)
}
