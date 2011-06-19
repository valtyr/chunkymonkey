package generation

import (
	"testing"

	"chunkymonkey/chunkstore"
	. "chunkymonkey/types"
)

func Benchmark_TestGenerator_generate(b *testing.B) {
	gen := NewTestGenerator(0)
	var loc ChunkXz
	resultChan := make(chan chunkstore.ChunkResult, 1)

	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// Generate a good sweep of different chunks, but don't go off forever.
		loc.X = ChunkCoord(i & 0xffff)
		gen.generate(loc, resultChan)
		// Don't block on next loop.
		_ = <-resultChan
	}
}
