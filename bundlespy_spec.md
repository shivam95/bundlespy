# bundlespy — Project Spec

**Type:** Go CLI tool  
**Purpose:** Scan a build output directory, report bundle health, diff against a saved baseline, and enforce size budgets as a CI gate.  
**Build time:** 1–2 days  
**Resume target:** Frontend Platform Engineer roles (DevEx, CI/CD, build tooling)

---

## CLI Interface

```bash
# Analyze a build output directory
bundlespy analyze --dir ./dist

# Save current build as baseline
bundlespy baseline --dir ./dist --out .bundlespy-baseline.json

# Analyze + compare to baseline (CI mode)
bundlespy analyze --dir ./dist --baseline .bundlespy-baseline.json --budget 5

# Show only the N largest assets
bundlespy analyze --dir ./dist --top 10

# Output as JSON (for piping into other tools)
bundlespy analyze --dir ./dist --format json
```

---

## Sample Output

```
bundlespy — Frontend Build Report
──────────────────────────────────
Tool:       dir
BuildTime:  0.0s
Assets:     34
Total:      2.41 MB  (gzip est: 741 KB)

Asset Breakdown:
  static/chunks/main.js      412 KB  [+18 KB  ▲ 4.6%              ]  ✓
  static/chunks/vendor.js    1.1 MB  [+62 KB  ▲ 5.9%              ]  ✗
  static/chunks/dashboard.js  88 KB  [-3 KB   ▼ -3.3%             ]  ✓
  static/css/styles.css       21 KB  [no baseline                 ]  ✓

FAIL — asset exceeded size budget (--budget 5%). Exit 1.
```

---

## Project Structure

```
bundlespy/
├── cmd/
│   └── bundlespy/
│       └── main.go                  ← CLI entrypoint: cobra commands, flag wiring
├── internal/
│   ├── parser/
│   │   ├── parser.go                ← Asset and BuildStats types
│   │   └── scandir.go               ← filesystem scanner (filepath.WalkDir)
│   ├── report/
│   │   └── report.go                ← diff logic, color output, JSON output
│   └── baseline/
│       └── baseline.go              ← read/write .bundlespy-baseline.json
├── testdata/
│   └── dist/                        ← dummy build output for testing
│       ├── static/chunks/
│       │   ├── main.js
│       │   ├── vendor.js
│       │   └── dashboard.js
│       └── static/css/
│           └── styles.css
├── go.mod
└── README.md
```

---

## How It Works

bundlespy uses `filepath.WalkDir` to traverse the build output directory and reads file sizes directly from the filesystem via `fs.DirEntry.Info().Size()`. No build tool plugins or stats file configuration required — just point it at `dist/`, `.next/static/`, or any build output folder.

Source maps (`.map` files) are automatically excluded since they are not served to users.

---

## Framework-Specific Paths

| Framework | Build command | Directory to scan |
|---|---|---|
| Next.js | `next build` | `.next/static/` |
| Vite | `vite build` | `dist/` |
| Create React App | `npm run build` | `build/static/` |
| webpack | `webpack` | `dist/` |

---

## Core Data Structures

```go
// internal/parser/parser.go

type Asset struct {
    Name    string
    Size    int64  // bytes
    IsChunk bool   // true for .js and .css files
}

type BuildStats struct {
    Tool      string  // always "dir"
    BuildTime int64   // always 0 (not available from filesystem)
    Hash      string  // always "" (not available from filesystem)
    Assets    []Asset
}
```

```go
// internal/parser/scandir.go

func ScanDir(dir string) (*BuildStats, error)
```

---

## Baseline Format

Saved as `.bundlespy-baseline.json` in project root. Commit this to git after a clean build.

```json
{
  "tool": "dir",
  "hash": "",
  "recorded_at": "2026-04-25T14:00:00Z",
  "assets": [
    { "name": "static/chunks/main.js",   "size": 421888, "is_chunk": true },
    { "name": "static/chunks/vendor.js", "size": 1153024, "is_chunk": true }
  ]
}
```

---

## Flags

### `analyze`

| Flag | Default | Description |
|---|---|---|
| `--dir` | *(required)* | Path to build output directory |
| `--baseline` | — | Path to baseline file |
| `--budget` | `0` | Max allowed % size increase per asset |
| `--top` | `0` | Show only the N largest assets (0 = all) |
| `--format` | `text` | Output format: `text` or `json` |

### `baseline`

| Flag | Default | Description |
|---|---|---|
| `--dir` | *(required)* | Path to build output directory |
| `--out` | `.bundlespy-baseline.json` | Output path for baseline file |

---

## What Was Built

| Feature | Status |
|---|---|
| `analyze` command — scan dir, print asset table | ✅ |
| `baseline` command — save snapshot | ✅ |
| Diff vs baseline — delta + % change per asset | ✅ |
| `--budget` flag — exit 1 if any asset exceeds threshold | ✅ |
| Cobra CLI + color output | ✅ |
| `--top N` flag | ✅ |
| `--format json` — machine-readable output | ✅ |
| Gzip size estimation (0.3 heuristic) | ✅ |
| GitHub Actions example in README | ✅ |
| Tool-agnostic filesystem scanner | ✅ |

---

## GitHub Actions Integration

```yaml
# .github/workflows/bundle-check.yml
name: Bundle Size Check
on: [pull_request]

jobs:
  bundle-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Install bundlespy
        run: go install github.com/shivam95/bundlespy/cmd/bundlespy@latest

      - name: Check bundle budgets
        run: bundlespy analyze --dir dist/ --baseline .bundlespy-baseline.json --budget 5
```

---

## Resume Bullets

```
• Built bundlespy, a Go CLI that scans frontend build output directories and generates
  per-asset size reports with baseline diffing — enabling automated enforcement of
  bundle size budgets as a CI gate in GitHub Actions pipelines across any build tool
  (webpack, Vite, Next.js, CRA).

• Implemented tool-agnostic bundle analysis using Go's filepath.WalkDir, eliminating
  the need for build tool plugins; added configurable size regression detection with
  exit-code-based CI enforcement and JSON output for pipeline integration.
```

---

## Key Links

- Go filepath.WalkDir: https://pkg.go.dev/path/filepath#WalkDir
- Go Cobra CLI: https://github.com/spf13/cobra
- fatih/color: https://github.com/fatih/color
