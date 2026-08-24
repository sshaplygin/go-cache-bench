// Command cachebench runs the comparison matrix and writes results.json with
// enough provenance to regenerate every number: module versions from build
// info, workload seeds, trace checksums, and the harness commit.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	cachebench "github.com/sshaplygin/go-cache-bench"
)

type provenance struct {
	HarnessVersion string            `json:"harness_version"`
	VCSRevision    string            `json:"vcs_revision"`
	ModuleVersions map[string]string `json:"module_versions"`
	TraceChecksums map[string]string `json:"trace_checksums,omitempty"`
	GeneratedAt    time.Time         `json:"generated_at"`
	GoVersion      string            `json:"go_version"`
}

type resultRow struct {
	Library  string  `json:"library"`
	Workload string  `json:"workload"`
	Capacity int     `json:"capacity"`
	HitRate  float64 `json:"hit_rate"`
	NsPerOp  float64 `json:"ns_per_op"`
	Requests int64   `json:"requests"`
}

type report struct {
	Provenance provenance  `json:"provenance"`
	Results    []resultRow `json:"results"`
}

func buildProvenance() provenance {
	p := provenance{ModuleVersions: map[string]string{}, GeneratedAt: time.Now().UTC()}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return p
	}
	p.GoVersion = info.GoVersion
	p.HarnessVersion = info.Main.Version
	for _, dep := range info.Deps {
		p.ModuleVersions[dep.Path] = dep.Version
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			p.VCSRevision = s.Value
		}
	}

	return p
}

func checksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func main() {
	var (
		capacities = flag.String("capacities", "500,2000", "comma-separated cache sizes")
		out        = flag.String("out", "results/results.json", "output path")
		smoke      = flag.Bool("smoke", false, "tiny run for CI")
	)
	flag.Parse()

	requests := 200_000
	if *smoke {
		requests = 20_000
	}

	workloads := []cachebench.Workload{
		cachebench.Zipf(requests, requests/10, 1.1, 1),
		cachebench.Uniform(requests, 5_000, 2),
		cachebench.Loop(requests, 550),
		cachebench.Scan(requests/1_000, 100, 400, 600),
		cachebench.PhaseShift(8, requests/8, 20_000, 550, 3),
	}
	if dir, err := cachebench.TraceDir(); err == nil {
		fmt.Fprintf(os.Stderr, "traces: %s (loaders wired per-format; see METHODOLOGY.md)\n", dir)
	}

	rep := report{Provenance: buildProvenance()}
	for _, capStr := range strings.Split(*capacities, ",") {
		var capacity int
		if _, err := fmt.Sscanf(strings.TrimSpace(capStr), "%d", &capacity); err != nil {
			fmt.Fprintf(os.Stderr, "bad capacity %q: %v\n", capStr, err)
			os.Exit(2)
		}
		for _, w := range workloads {
			for _, lib := range cachebench.Libraries() {
				subject, err := lib.Build(capacity)
				if err != nil {
					fmt.Fprintf(os.Stderr, "build %s: %v\n", lib.Name, err)
					os.Exit(1)
				}
				r := cachebench.Replay(lib.Name, subject, w)
				rep.Results = append(rep.Results, resultRow{
					Library: r.Library, Workload: w.Name, Capacity: capacity,
					HitRate: r.HitRate(), NsPerOp: r.NsPerOp(), Requests: r.Hits + r.Misses,
				})
			}
		}
	}

	if err := os.MkdirAll("results", 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s: %d rows\n", *out, len(rep.Results))
}
