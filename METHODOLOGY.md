# Methodology

Every number this repository publishes is produced by `cmd/cachebench` and
carries provenance: module versions and VCS revision from Go build info,
workload seeds, and SHA-256 checksums of any trace files used. A number that
cannot be regenerated does not get published.

## Measurement rules

1. **Identical replay.** Every library is driven through the same two-method
   interface (`Get`, `Add`) by the same deterministic request sequence. Hits
   and misses are counted by the harness, never read from a library's own
   statistics — several libraries sample or approximate their internal
   accounting.
2. **Deterministic workloads.** Generators are seeded PCG; two libraries are
   always compared on the identical key sequence. Wall-clock plays no role in
   workload construction.
3. **Default shape, stated deviations.** Each library is configured to its
   documented default shape. Any deviation is stated in a comment on its
   builder in `libraries.go`. Where `github.com/maypok86/benchmarks` also
   benchmarks a library, configuration matches it so the two publications can
   be read side by side.
4. **Count-based capacity.** The harness measures object hit rate at an
   entry-count capacity. Byte-weighted miss ratio on variable-size objects is
   out of scope here (see Limitations).

## Known measurement hazards (and what this harness does about them)

These were found the hard way and are encoded as conformance tests:

- **otter admits ahead of eviction.** otter evicts on a maintenance pass, so
  a tight replay loop can hold far more than the configured capacity (5000
  writes into a 500-cap cache left 1916 retrievable before maintenance ran).
  The adapter forces cleanup so the measurement reflects the configured
  capacity. Without this, otter's hit rate is inflated by measuring a bigger
  cache than requested.
- **ristretto applies Sets asynchronously.** A Set may be dropped or applied
  late; the adapter waits where the API allows, and the conformance test
  asserts retrievability assumptions explicitly.
- **Wall-clock anything.** Any subject whose behaviour depends on elapsed
  time (TTL arms, clock-based epochs) makes results depend on machine load.
  Such configurations are excluded from headline tables and marked where
  they appear.

## Disputes

A disputed number is resolved by a rerun, not an argument. Open an issue with
the row you dispute; the maintainers rerun the pinned configuration in CI and
publish the fresh results file either way. If your library's adapter
misconfigures it, a PR fixing the builder plus the regenerated table is the
preferred format.

## Limitations

- Object-count capacity only; no byte-weighted miss ratios yet.
- Single-threaded replay measures policy quality and per-op cost, not
  contended throughput. For contended throughput see
  `github.com/maypok86/benchmarks`, whose scope this repository deliberately
  does not duplicate.
- Trace files are not redistributed (size and licensing); `CACHE_BENCH_TRACES`
  points at a local directory, and checksums in results files pin exactly
  which bytes were replayed.

## Prior art

`github.com/maypok86/benchmarks` (otter's author: throughput, memory,
hit-ratio simulator), ben-manes/caffeine's simulator, and
`github.com/1a1a11a/libCacheSim` (the reference trace simulator; this
harness's block-trace semantics follow it). This repository exists to add
what those do not aim at: a library-neutral, version-pinned, provenance-
carrying comparison maintained as a publication.
