package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/fatih/color"

	"github.com/shivam95/bundlespy/internal/baseline"
	"github.com/shivam95/bundlespy/internal/parser"
)

// SprintFunc returns a function that wraps a string in ANSI codes.
// We call SprintFunc() once and reuse the returned function — cheaper than
// creating a new Color each time we print a line.
var (
	bold   = color.New(color.Bold).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	dim    = color.New(color.Faint).SprintFunc()
)

// AssetDiff holds a single asset's current size and its diff against baseline.
type AssetDiff struct {
	Name          string
	Size          int64
	Delta         int64
	DeltaPct      float64
	HasBaseline   bool
	ExceedsBudget bool
}

// Build computes diffs between current stats and an optional baseline.
// budget is the max allowed % increase (e.g. 5.0 for 5%). Pass 0 to skip enforcement.
func Build(stats *parser.BuildStats, base *baseline.Baseline, budget float64) []AssetDiff {
	baseMap := make(map[string]int64)
	if base != nil {
		for _, a := range base.Assets {
			baseMap[a.Name] = a.Size
		}
	}

	diffs := make([]AssetDiff, 0, len(stats.Assets))
	for _, a := range stats.Assets {
		d := AssetDiff{Name: a.Name, Size: a.Size}

		if baseSize, ok := baseMap[a.Name]; ok {
			d.HasBaseline = true
			d.Delta = a.Size - baseSize
			d.DeltaPct = float64(d.Delta) / float64(baseSize) * 100
			d.ExceedsBudget = budget > 0 && d.DeltaPct > budget
		}

		diffs = append(diffs, d)
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Size > diffs[j].Size
	})

	return diffs
}

// gzipEst returns a rough gzip size estimate using the 0.3 heuristic.
// Minified JS/CSS typically compresses to ~25-35% of its original size.
func gzipEst(size int64) int64 {
	return int64(float64(size) * 0.3)
}

// totalSize sums the Size of all diffs.
func totalSize(diffs []AssetDiff) int64 {
	var total int64
	for _, d := range diffs {
		total += d.Size
	}
	return total
}

// Print writes the formatted report to w. Returns true if any asset exceeded budget.
func Print(w io.Writer, stats *parser.BuildStats, diffs []AssetDiff, budget float64) bool {
	failed := false

	total := totalSize(diffs)

	fmt.Fprintln(w, bold("bundlespy — Frontend Build Report"))
	fmt.Fprintln(w, "──────────────────────────────────")
	fmt.Fprintf(w, "%s %s\n", dim("Tool:      "), stats.Tool)
	fmt.Fprintf(w, "%s %.1fs\n", dim("BuildTime:"), float64(stats.BuildTime)/1000)
	fmt.Fprintf(w, "%s %d\n", dim("Assets:   "), len(stats.Assets))
	fmt.Fprintf(w, "%s %s  %s\n", dim("Total:    "), formatBytes(total), dim("(gzip est: "+formatBytes(gzipEst(total))+")"))
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, bold("Asset Breakdown:"))

	for _, d := range diffs {
		// Pad raw strings BEFORE applying color — ANSI codes inflate byte
		// length so fmt width verbs break alignment if you colorize first.
		paddedSize := fmt.Sprintf("%8s", formatBytes(d.Size))
		paddedDiff := fmt.Sprintf("%-28s", rawDiff(d))

		var coloredDiff string
		switch {
		case !d.HasBaseline:
			coloredDiff = dim(paddedDiff)
		case d.ExceedsBudget:
			coloredDiff = red(paddedDiff)
		case d.Delta > 0:
			coloredDiff = yellow(paddedDiff)
		default:
			coloredDiff = green(paddedDiff)
		}

		status := green("✓")
		if d.ExceedsBudget {
			status = red("✗")
			failed = true
		}

		fmt.Fprintf(w, "  %-20s %s  [%s]  %s\n",
			d.Name, cyan(paddedSize), coloredDiff, status)
	}

	fmt.Fprintln(w, "")
	if failed {
		fmt.Fprintln(w, red(fmt.Sprintf("FAIL — asset exceeded size budget (--budget %.0f%%). Exit 1.", budget)))
	} else {
		fmt.Fprintln(w, green("OK — all assets within budget."))
	}

	return failed
}

// jsonReport is the shape of the JSON output — struct tags define the key names.
type jsonReport struct {
	Tool        string      `json:"tool"`
	BuildTimeMs int64       `json:"build_time_ms"`
	TotalAssets int         `json:"total_assets"`
	TotalSize   int64       `json:"total_size"`
	GzipEst     int64       `json:"gzip_est"`
	Assets      []jsonAsset `json:"assets"`
	Passed      bool        `json:"passed"`
}

type jsonAsset struct {
	Name          string  `json:"name"`
	Size          int64   `json:"size"`
	Delta         int64   `json:"delta,omitempty"`
	DeltaPct      float64 `json:"delta_pct,omitempty"`
	HasBaseline   bool    `json:"has_baseline"`
	ExceedsBudget bool    `json:"exceeds_budget"`
}

// PrintJSON writes a machine-readable JSON report to w.
func PrintJSON(w io.Writer, stats *parser.BuildStats, diffs []AssetDiff) bool {
	failed := false
	assets := make([]jsonAsset, len(diffs))
	for i, d := range diffs {
		assets[i] = jsonAsset{
			Name:          d.Name,
			Size:          d.Size,
			Delta:         d.Delta,
			DeltaPct:      d.DeltaPct,
			HasBaseline:   d.HasBaseline,
			ExceedsBudget: d.ExceedsBudget,
		}
		if d.ExceedsBudget {
			failed = true
		}
	}

	total := totalSize(diffs)
	out := jsonReport{
		Tool:        stats.Tool,
		BuildTimeMs: stats.BuildTime,
		TotalAssets: len(stats.Assets),
		TotalSize:   total,
		GzipEst:     gzipEst(total),
		Assets:      assets,
		Passed:      !failed,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(out)
	return failed
}

// rawDiff returns the uncolored diff string for an asset.
func rawDiff(d AssetDiff) string {
	if !d.HasBaseline {
		return "no baseline"
	}
	sign := "+"
	if d.Delta < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%s  %s%.1f%%", sign, formatBytes(d.Delta), arrowFor(d.Delta), d.DeltaPct)
}

func formatBytes(b int64) string {
	const kb = 1024
	const mb = 1024 * kb
	switch {
	case b < 0:
		return "-" + formatBytes(-b)
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/mb)
	case b >= kb:
		return fmt.Sprintf("%d KB", b/kb)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func arrowFor(delta int64) string {
	switch {
	case delta > 0:
		return "▲ "
	case delta < 0:
		return "▼ "
	default:
		return "— "
	}
}
