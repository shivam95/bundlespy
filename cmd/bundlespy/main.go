package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/shivam95/bundlespy/internal/baseline"
	"github.com/shivam95/bundlespy/internal/parser"
	"github.com/shivam95/bundlespy/internal/report"
)

func main() {
	root := &cobra.Command{
		Use:   "bundlespy",
		Short: "Frontend bundle size analyzer and CI gate",
		// Suppress cobra's default error/usage printing so we control the format.
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newAnalyzeCmd())
	root.AddCommand(newBaselineCmd())

	if err := root.Execute(); err != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func newAnalyzeCmd() *cobra.Command {
	var dir, baselinePath, format string
	var budget float64
	var top int

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze a build output directory and report asset sizes",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := parser.ScanDir(dir)
			if err != nil {
				return err
			}

			var base *baseline.Baseline
			if baselinePath != "" {
				base, err = baseline.Load(baselinePath)
				if err != nil {
					return err
				}
			}

			diffs := report.Build(stats, base, budget)
			if top > 0 && top < len(diffs) {
				diffs = diffs[:top]
			}

			var failed bool
			if format == "json" {
				failed = report.PrintJSON(os.Stdout, stats, diffs)
			} else {
				failed = report.Print(os.Stdout, stats, diffs, budget)
			}
			if failed {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "path to build output directory (required)")
	cmd.Flags().StringVar(&baselinePath, "baseline", "", "path to baseline file")
	cmd.Flags().Float64Var(&budget, "budget", 0, "max allowed % size increase per asset (e.g. 5)")
	cmd.Flags().IntVar(&top, "top", 0, "show only the N largest assets")
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	cmd.MarkFlagRequired("dir")

	return cmd
}

func newBaselineCmd() *cobra.Command {
	var dir, outPath string

	cmd := &cobra.Command{
		Use:   "baseline",
		Short: "Save current build output as a baseline snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := parser.ScanDir(dir)
			if err != nil {
				return err
			}

			if err := baseline.Save(stats, outPath); err != nil {
				return err
			}

			color.New(color.FgGreen).Fprintf(os.Stdout, "baseline saved to %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "path to build output directory (required)")
	cmd.Flags().StringVar(&outPath, "out", ".bundlespy-baseline.json", "output path for baseline file")
	cmd.MarkFlagRequired("dir")

	return cmd
}
