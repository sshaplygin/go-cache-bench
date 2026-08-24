package cachebench

import (
	ascache "github.com/sshaplygin/as-cache"
	"github.com/sshaplygin/as-cache/bandit"
	"github.com/sshaplygin/as-cache/policies"
	"github.com/sshaplygin/as-cache/policies/arc"
	"github.com/sshaplygin/as-cache/policies/fifo"
	"github.com/sshaplygin/as-cache/policies/tinylfu"
)

// as-cache appears in the matrix as one row among peers. This file is the
// only place that imports it, so dropping the row is deleting one file.
//
// Configuration is the one its own README recommends, not one that flatters
// it: request-counted epochs (wall-clock epochs made this table move 12
// points between runs on a loaded machine), warm migration, full arm set.

// newASCacheLibrary builds an adaptive as-cache with every policy it ships.
func newASCacheLibrary(size int) (Cache, error) {
	arms := []ascache.Policy[string, int]{}
	builders := []func(int) (ascache.Policy[string, int], error){
		func(n int) (ascache.Policy[string, int], error) { return policies.NewLRU[string, int](n) },
		func(n int) (ascache.Policy[string, int], error) { return policies.NewLFU[string, int](n) },
		func(n int) (ascache.Policy[string, int], error) { return policies.NewTwoQueue[string, int](n) },
		arc.NewPolicy[string, int],
		tinylfu.NewPolicy[string, int],
		fifo.NewS3FIFOPolicy[string, int],
		fifo.NewSievePolicy[string, int],
	}
	for _, build := range builders {
		p, err := build(size)
		if err != nil {
			return nil, err
		}
		arms = append(arms, p)
	}

	return ascache.NewAdaptiveCache(arms,
		bandit.NewThompson(0.7, 7),
		&ascache.Settings{
			EpochRequests:               2_000,
			EvictPartialCapacityFilling: true,
			MigrationStrategy:           ascache.MigrationWarm,
		})
}
