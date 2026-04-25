# bundlespy

CLI tool that analyzes a build output directory, diffs against a saved baseline, and enforces size budgets as a CI gate. Works with any build tool — webpack, Vite, Next.js, Parcel, or anything that produces a `dist/` folder.

## Install

```bash
go install github.com/shivam95/bundlespy/cmd/bundlespy@latest
```

Or build from source:

```bash
git clone https://github.com/shivam95/bundlespy
cd bundlespy
go build -o bundlespy ./cmd/bundlespy/
```

## Usage

### Analyze a build output directory

```bash
bundlespy analyze --dir dist/
```

### Save a baseline

Run this once after a clean build and commit `.bundlespy-baseline.json` to git.

```bash
bundlespy baseline --dir dist/ --out .bundlespy-baseline.json
```

### Compare against baseline (CI mode)

Exits 1 if any asset grows more than 5% vs baseline.

```bash
bundlespy analyze --dir dist/ --baseline .bundlespy-baseline.json --budget 5
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--dir` | *(required)* | Path to build output directory |
| `--baseline` | — | Path to baseline file |
| `--budget` | `0` | Max allowed % size increase per asset |
| `--top` | `0` | Show only the N largest assets |
| `--format` | `text` | Output format: `text` or `json` |

## Sample Output

```
bundlespy — Frontend Build Report
──────────────────────────────────
Tool:       dir
BuildTime:  0.0s
Assets:     34
Total:      2.41 MB  (gzip est: 741 KB)

Asset Breakdown:
  static/chunks/vendor.js    1.1 MB  [+62 KB  ▲ 5.9%              ]  ✗
  static/chunks/main.js      412 KB  [+18 KB  ▲ 4.6%              ]  ✓
  static/chunks/dashboard.js  88 KB  [-3 KB   ▼ -3.3%             ]  ✓
  static/css/styles.css       21 KB  [no baseline                 ]  ✓

FAIL — asset exceeded size budget (--budget 5%). Exit 1.
```

## Framework-specific paths

| Framework | Build command | Directory to scan |
|---|---|---|
| Next.js | `next build` | `.next/static/` |
| Vite | `vite build` | `dist/` |
| Create React App | `npm run build` | `build/static/` |
| webpack | `webpack` | `dist/` |

Source maps (`.map` files) are automatically excluded.

## GitHub Actions

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
