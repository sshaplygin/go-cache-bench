// Package cachebench replays deterministic workloads and real traces against
// Go cache libraries, so a team choosing a cache can read one table instead
// of four READMEs.
//
// Neutrality is the product: no library author owns this comparison, every
// library is one row among peers, and every published number embeds enough
// provenance (module versions, trace checksums, commit) to be regenerated.
// Disputed numbers are resolved by reruns, not arguments; see METHODOLOGY.md.
package cachebench

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Result is one library's performance on one workload.
type Result struct {
	Library  string
	Workload string
	Capacity int
	Hits     int64
	Misses   int64
	Duration time.Duration
}

// HitRate returns the fraction of requests served from cache.
func (r Result) HitRate() float64 {
	total := r.Hits + r.Misses
	if total == 0 {
		return 0
	}

	return float64(r.Hits) / float64(total)
}

// NsPerOp returns the average time per request.
func (r Result) NsPerOp() float64 {
	total := r.Hits + r.Misses
	if total == 0 {
		return 0
	}

	return float64(r.Duration.Nanoseconds()) / float64(total)
}

// Cache is the subset of cache behaviour the harness replays against. Every
// library adapter implements exactly this, so every subject is measured
// identically regardless of its native API.
type Cache interface {
	Get(key string) (int, bool)
	Add(key string, value int) bool
}

// Replay runs a workload against a cache, filling on every miss exactly as a
// read-through cache would, and reports what it served.
//
// Hits and misses are counted here rather than read from the cache's own
// statistics so that every subject is measured identically, including ones
// whose internal accounting is sampled.
func Replay(name string, c Cache, w Workload) Result {
	result := Result{Library: name, Workload: w.Name}

	start := time.Now()
	for i, key := range w.Keys {
		if _, ok := c.Get(key); ok {
			result.Hits++

			continue
		}
		result.Misses++
		c.Add(key, i)
	}
	result.Duration = time.Since(start)

	return result
}

// Table renders results as a markdown table, best hit rate first.
func Table(results []Result) string {
	sorted := make([]Result, len(results))
	copy(sorted, results)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].HitRate() > sorted[j].HitRate()
	})

	var b strings.Builder
	b.WriteString("| Library | Hit rate | ns/op |\n| --- | --- | --- |\n")
	for _, r := range sorted {
		fmt.Fprintf(&b, "| %s | %.2f%% | %.0f |\n", r.Library, r.HitRate()*100, r.NsPerOp())
	}

	return b.String()
}

// Ranking returns library names ordered best hit rate first.
func Ranking(results []Result) []string {
	sorted := make([]Result, len(results))
	copy(sorted, results)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].HitRate() > sorted[j].HitRate()
	})

	names := make([]string, len(sorted))
	for i, r := range sorted {
		names[i] = r.Library
	}

	return names
}
