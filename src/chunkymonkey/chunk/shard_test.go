package chunk

import (
	"testing"

	. "chunkymonkey/types"
)

func TestShardCoordFromChunkCoord(t *testing.T) {
	type Test struct {
		input    ChunkCoord
		expected shardCoord
	}

	tests := []Test{
		{-2*ShardSize - 1, -3},
		{-2 * ShardSize, -2},
		{-ShardSize - 1, -2},
		{-ShardSize, -1},
		{-1, -1},
		{0, 0},
		{ShardSize - 1, 0},
		{ShardSize, 1},
		{2*ShardSize - 1, 1},
		{2 * ShardSize, 2},
	}

	for _, test := range tests {
		result := shardCoordFromChunkCoord(test.input)
		if test.expected != result {
			t.Errorf(
				"ChunkCoord(%d) expected %d, but got %d",
				test.input, test.expected, result,
			)
		}
	}
}
