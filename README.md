# IPTV M3U Enhancer (CLI)

A tiny Go CLI to parse an M3U playlist and write a filtered playlist containing only entries from a specified `group-title`.

## Build

```bash
go build -o iptv-m3u-enhancer .
```

## Usage

```bash
iptv-m3u-enhancer [--group-title "<name>"] [--out <path>] [--strict] <input.m3u>
```

- `--group-title "<name>"`: filter entries by `group-title` (case-insensitive). If omitted, all entries are included.
- `--out <path>`: output M3U path. If omitted, defaults to `<input>.<group>.m3u` (or `<input>.filtered.m3u` when no filter is given) in the same directory.
- `--strict`: fail on malformed lines and structural issues.

## Notes

- The parser preserves the original `#EXTINF` line for each entry (attributes, order, spacing) when writing the filtered file.
- Non-`#EXTINF` tags and additional metadata are ignored for now.
- Future: local start time will be derived from entry titles if present.

# iptv-m3u-enhancer
Filter a group and sort by event start time your  daily generated IPTV m3u file
