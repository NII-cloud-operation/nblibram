# nblibram

The notebook Swiss Army knife—query, edit, and sanitize Jupyter files from the shell.
Treats notebooks as structured documents so agents and humans can slice, mutate, and clean them without a GUI.
Secret detection is powered by [gitleaks](https://github.com/gitleaks/gitleaks) (222+ built-in rules for API keys, tokens, credentials, etc.).
Ships as a single Go binary with no runtime dependencies.

## Philosophy

Notebooks are computational narratives—code cells don't stand alone, they're explained by the Markdown cells around them.
nblibram's query operations slice Markdown/code cell pairs together so the surrounding story always travels with the code.
`section` slices by heading hierarchy, `cells` slices by cell order, both preserving the narrative context that makes code understandable.

## Commands (stdin → stdout)

All commands read `.ipynb` JSON from stdin (or `--file`) and write to stdout.
Query commands apply gitleaks-based privacy filters by default (`--no-filter` to disable).

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

### Filter & Audit

- `nblibram filter` – sanitize sensitive information in a notebook (`-i` for in-place).
- `nblibram audit` – check for leaked secrets without modifying the notebook. Exits 1 if leaks are found (`--format text|json`).

Both use gitleaks' built-in rules by default. Set `NBLIBRAM_GITLEAKS_CONFIG` to a `.gitleaks.toml` path to add custom rules (e.g. IP addresses, domain names, internal URLs).

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

# Audit for secrets (CI-friendly, exits 1 on findings)
nblibram audit < notebook.ipynb

# Audit with JSON output
nblibram audit --format json < notebook.ipynb

# Read a pickled kernel log
nblibram pkl --file output.pkl --format text
```

## Getting started

Download the latest binary from [Releases](https://github.com/NII-cloud-operation/nblibram/releases) and place it in your `$PATH`:

```bash
# macOS (Apple Silicon)
curl -sL https://github.com/NII-cloud-operation/nblibram/releases/latest/download/nblibram_darwin_arm64.tar.gz | tar xz
mv nblibram /usr/local/bin/

# macOS (Intel)
curl -sL https://github.com/NII-cloud-operation/nblibram/releases/latest/download/nblibram_darwin_amd64.tar.gz | tar xz
mv nblibram /usr/local/bin/

# Linux (amd64)
curl -sL https://github.com/NII-cloud-operation/nblibram/releases/latest/download/nblibram_linux_amd64.tar.gz | tar xz
sudo mv nblibram /usr/local/bin/
```

## Build from source

```bash
go build ./cmd/nblibram/
```

## Configuration

nblibram uses [gitleaks](https://github.com/gitleaks/gitleaks) for secret detection. By default, gitleaks' 222+ built-in rules are active.

To add custom rules (e.g. IP addresses, internal URLs), create a `.gitleaks.toml` and point to it:

```bash
export NBLIBRAM_GITLEAKS_CONFIG=/path/to/.gitleaks.toml
```

Example `.gitleaks.toml`:

```toml
[extend]
useDefault = true

[[rules]]
id = "ipv4-address"
description = "Detects IPv4 addresses"
regex = '''\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b'''
[rules.allowlist]
regexes = [
    '''192\.168\.\d{1,3}\.\d{1,3}''',
    '''127\.0\.0\.1''',
    '''0\.0\.0\.0''',
]

[[rules]]
id = "email-address"
description = "Detects email addresses"
regex = '''[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'''
[rules.allowlist]
regexes = [
    '''.*@example\.com''',
]
```

Detected secrets are replaced with `[rule-id_N]` labels, preserving equivalence (e.g. the same secret always maps to `[generic-api-key_1]`).
