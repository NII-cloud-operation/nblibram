# nblibram

The notebook Swiss Army knife—query, edit, and sanitize Jupyter files from the shell.
Treats notebooks as structured documents so agents and humans can slice, mutate, and clean them without a GUI.
Ships as a single Go binary with no runtime dependencies.

## Commands (stdin → stdout)

All commands read `.ipynb` JSON from stdin (or `--file`) and write to stdout.

### Query

- `nblibram toc` – heading structure with preview (`--words`, `--format md|json`).
- `nblibram section` – cells under a heading, including nested subsections (`--sets`, `--format md|json|py`).
- `nblibram cells` – consecutive cells from a matched position (`--count N`), or Markdown+code pairs (`--sets N`).
- `nblibram outputs` – cell outputs (`--format text|json|raw`, `--mime`).

### Mutate

- `nblibram insert` – insert a cell (`--query`, `--position before|after`, `--type code|markdown`, `--source`).
- `nblibram update` – replace cell content (`--query`, `--source`, `--hash` required).
- `nblibram delete` – remove a cell (`--query`, `--hash` required).

Mutation commands write the modified notebook JSON to stdout. Use `-i` for in-place file update.
`--hash` enforces optimistic locking—obtain it via `nblibram hash`.

### Filter

- `nblibram filter` – sanitize sensitive information (`--config`, `--gitleaks`, `-i`).
- `nblibram init-config` – create default `~/.nbfilterrc.toml`.

Supports [gitleaks](https://github.com/gitleaks/gitleaks) rule files via `--gitleaks` for extended secret detection.

### Utility

- `nblibram hash` – compute djb2 hashes for cells (`--query` or `--all`).
- `nblibram pkl` – read pickled kernel output logs (`--file`, `--format json|text`).

## Queries

Use `--query TYPE:VALUE` to locate cells. Multiple `--query` flags are ANDed.

- `start:37` – absolute cell index.
- `match:"pattern"` – regex against cell content.
- `contains:"text"` – substring match.
- `id:abc123` – Jupyter cell_id.
- `meme:UUID` – nblineage meme ID. Trailing `*` for prefix match (e.g. `meme:642f96e0*` matches branched cells).

## Examples

```bash
# Table of contents
nblibram toc --format json < notebook.ipynb

# Extract a section
nblibram section --query match:"## Setup" --sets 2 < notebook.ipynb

# Get 5 consecutive cells starting from index 10
nblibram cells --query start:10 --count 5 --format py < notebook.ipynb

# Extract output as PNG
nblibram outputs --query id:plot-cell --format raw --mime image/png < notebook.ipynb > plot.png

# Insert a cell, then sanitize
nblibram insert --query start:0 --source 'x = 1' < notebook.ipynb | nblibram filter > clean.ipynb

# Delete with optimistic locking
HASH=$(nblibram hash --query start:3 < notebook.ipynb | jq -r '.[0]._hash')
nblibram delete --query start:3 --hash "$HASH" -i notebook.ipynb

# Sanitize with gitleaks rules
nblibram filter --gitleaks .gitleaks.toml < notebook.ipynb > sanitized.ipynb

# Read a pickled kernel log
nblibram pkl --file output.pkl --format text
```

## Build

```bash
go build ./cmd/nblibram/
```

## Configuration

Filter patterns live in `~/.nbfilterrc.toml`:

```toml
[[filters]]
pattern = '\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b'
label = "[IPv4_#]"

[[filters]]
pattern = '[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.(com|org|net|jp|io|dev|local|internal)'
label = "[DOMAIN_#]"
```

`#` in the label is replaced with a sequential number per unique match, preserving equivalence (e.g. the same IP always maps to `[IPv4_1]`).
