# go-cache-bench

A neutral, reproducible comparison of Go in-process cache libraries: hit
rate and per-op cost, across synthetic workloads and real traces, with every
published number carrying the provenance to regenerate it.

**Who this is for:** a Go team spending ten minutes choosing a cache
dependency, and library authors who want a fair external scoreboard.

**Neutrality rule:** no library author owns this comparison. Every library —
including `as-cache`, whose repository this harness was extracted from — is
one row among peers, built from its documented default configuration. The
`as-cache` adapter lives in one droppable file and gets no special placement.

## Matrix

| Library | Eviction | Adapter notes |
| --- | --- | --- |
| otter v2 | S3-FIFO/W-TinyLFU lineage | forced maintenance so capacity holds (see METHODOLOGY) |
| theine | W-TinyLFU | defaults |
| ristretto v2 | sampled LFU | waits out async Set (see METHODOLOGY) |
| sturdyc | LRU + stampede protection | defaults |
| as-cache | adaptive (bandit over 9 arms incl. S3-FIFO, SIEVE) | README-recommended settings; FIFO arms via upstream wrapper — see note |

Workloads: `zipf`, `uniform`, `loop`, `scan`, `phase-shift` (deterministic,
seeded), plus trace replay from a local `CACHE_BENCH_TRACES` directory:
ARC traces, MSR Cambridge block traces (512-byte sectoring matching the
Caffeine simulator, reads-only by default, per-record expansion bounded),
and Meta CacheLib kvcache traces. Loaders are extracted from as-cache
@69c66fd, where they ship with 450+ lines of format tests.

**Note on the as-cache FIFO arms:** upstream currently adapts S3-FIFO and
SIEVE through `scalalang2/golang-fifo`, whose Resize rebuilds the cache and
discards hand/visited/ghost state. Numbers for those two arms inside the
adaptive row therefore reflect the wrapper, not the pure algorithms; the
pure algorithms will join the matrix as standalone rows once as-cache ships
native leaf modules (its STAGE M).

## Run it

```bash
go run ./cmd/cachebench -capacities 500,2000 -out results/results.json
```

`results.json` embeds module versions, the harness VCS revision, seeds, and
trace checksums. `make smoke` runs a small CI-sized matrix.

## Disputing a number

Open an issue naming the row. We rerun the pinned configuration and publish
the fresh results file either way. See METHODOLOGY.md.

## License

MPL-2.0, matching the repository this was extracted from.
