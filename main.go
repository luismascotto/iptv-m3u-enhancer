package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ExtInf struct {
	Duration   int
	Title      string
	Attributes map[string]string
	Raw        string
	// Parsed times (when present in title). UTC source converted to local as well.
	StartTimeUTC   *time.Time
	StartTimeLocal *time.Time
}

func (e ExtInf) GroupTitle() string {
	if e.Attributes == nil {
		return ""
	}
	return e.Attributes["group-title"]
}

type PlaylistEntry struct {
	Info ExtInf
	URI  string
}

type Playlist struct {
	Entries       []PlaylistEntry
	HeaderPresent bool
}

func parseM3U(path string, strict bool) (Playlist, error) {
	f, err := os.Open(path)
	if err != nil {
		return Playlist{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Increase the scanner buffer to handle long attribute lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var (
		lineNum          int
		firstNonEmptySet bool
		headerSeen       bool
		currentEXTINF    *ExtInf
		entries          []PlaylistEntry
	)

	fallbackYear := extractFallbackYear(path)

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if !firstNonEmptySet {
			firstNonEmptySet = true
			if strings.EqualFold(line, "#EXTM3U") {
				headerSeen = true
				continue
			}
			// If the first non-empty line isn't #EXTM3U
			if strict {
				return Playlist{}, fmt.Errorf("line %d: expected #EXTM3U header", lineNum)
			}
			// Non-strict: continue parsing
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			info, err := parseEXTINF(line, fallbackYear)
			if err != nil {
				if strict {
					return Playlist{}, fmt.Errorf("line %d: %w", lineNum, err)
				}
				// Skip malformed EXTINF in non-strict mode
				currentEXTINF = nil
				continue
			}
			// Preserve the raw line for re-emitting later
			info.Raw = line
			currentEXTINF = &info
			continue
		}

		// Other tags/comments can be ignored for now
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Non-comment non-empty lines should be URIs
		if currentEXTINF != nil {
			entries = append(entries, PlaylistEntry{
				Info: *currentEXTINF,
				URI:  line,
			})
			currentEXTINF = nil
			continue
		}
		// URI without prior EXTINF
		if strict {
			return Playlist{}, fmt.Errorf("line %d: URI without preceding #EXTINF", lineNum)
		}
		// Non-strict: ignore
	}

	if err := scanner.Err(); err != nil {
		return Playlist{}, err
	}
	if currentEXTINF != nil && strict {
		return Playlist{}, errors.New("file ended after #EXTINF without URI")
	}
	return Playlist{
		Entries:       entries,
		HeaderPresent: headerSeen,
	}, nil
}

var (
	attrKVQuoted = regexp.MustCompile(`(?i)([a-z0-9\-]+)="([^"]*)"`)
	attrKVPlain  = regexp.MustCompile(`(?i)\b([a-z0-9\-]+)=([^\s,]+)`)
	// Title time patterns:
	// 1) start:YYYY MM DD HH:mm(:SS)?
	reStartInTitle = regexp.MustCompile(`(?i)start:\s*(\d{4})\s+(\d{1,2})\s+(\d{1,2})\s+(\d{1,2}):(\d{2})(?::(\d{2}))?`)
	// 2) (MM.DD H:mmTZ) where TZ is a US time band like ET, CT, MT, PT (also EST/EDT, etc.)
	reParenTZ = regexp.MustCompile(`(?i)\((\d{1,2})\.(\d{1,2})\s+(\d{1,2}):(\d{2})\s*([A-Z]{1,4})\)`)
	// Date in filename/path: YYYY-MM-DD
	reDateInPath = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
)

func parseEXTINF(line string, fallbackYear int) (ExtInf, error) {
	// Expect: #EXTINF:<duration> [attributes],<title>
	const prefix = "#EXTINF:"
	if !strings.HasPrefix(line, prefix) {
		return ExtInf{}, fmt.Errorf("not an EXTINF line: %q", line)
	}
	payload := strings.TrimSpace(line[len(prefix):])
	// Split into meta and title by the last comma or first? Standard uses last part after the last comma as title
	// but attributes shouldn't contain commas outside of quotes. Safer: split on the first comma not within quotes.
	metaPart, titlePart, err := splitMetaAndTitle(payload)
	if err != nil {
		return ExtInf{}, err
	}
	title := strings.TrimSpace(titlePart)

	// Duration is first token in metaPart before any space
	metaPart = strings.TrimSpace(metaPart)
	if metaPart == "" {
		return ExtInf{}, fmt.Errorf("missing duration and attributes")
	}
	firstSpace := strings.IndexByte(metaPart, ' ')
	var durationStr string
	var attrsStr string
	if firstSpace == -1 {
		durationStr = metaPart
		attrsStr = ""
	} else {
		durationStr = strings.TrimSpace(metaPart[:firstSpace])
		attrsStr = strings.TrimSpace(metaPart[firstSpace+1:])
	}

	dur, err := strconv.Atoi(durationStr)
	if err != nil {
		return ExtInf{}, fmt.Errorf("invalid duration %q", durationStr)
	}

	attributes := make(map[string]string, 8)
	// Prefer quoted key="value" matches
	for _, m := range attrKVQuoted.FindAllStringSubmatch(attrsStr, -1) {
		key := strings.ToLower(m[1])
		val := m[2]
		attributes[key] = val
	}
	// Add plain key=value matches not already set (so quoted wins)
	for _, m := range attrKVPlain.FindAllStringSubmatch(attrsStr, -1) {
		key := strings.ToLower(m[1])
		val := m[2]
		if _, exists := attributes[key]; !exists {
			attributes[key] = val
		}
	}

	ext := ExtInf{
		Duration:   dur,
		Title:      title,
		Attributes: attributes,
	}
	if utc, local := parseTimesFromTitle(title, fallbackYear); utc != nil && local != nil {
		ext.StartTimeUTC = utc
		ext.StartTimeLocal = local
	}
	return ext, nil
}

func splitMetaAndTitle(payload string) (string, string, error) {
	// Split on the first comma not inside double quotes
	inQuotes := false
	for i := 0; i < len(payload); i++ {
		ch := payload[i]
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if ch == ',' && !inQuotes {
			return payload[:i], payload[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("invalid EXTINF, missing title separator ','")
}

func extractFallbackYear(path string) int {
	base := filepath.Base(path)
	if m := reDateInPath.FindStringSubmatch(base); m != nil {
		if y, err := strconv.Atoi(m[1]); err == nil {
			return y
		}
	}
	// As a secondary attempt, search full path in case the date is elsewhere.
	if m := reDateInPath.FindStringSubmatch(path); m != nil {
		if y, err := strconv.Atoi(m[1]); err == nil {
			return y
		}
	}
	return time.Now().Year()
}

func parseTimesFromTitle(title string, fallbackYear int) (*time.Time, *time.Time) {
	// Prefer explicit 'start:' form if present
	if m := reStartInTitle.FindStringSubmatch(title); m != nil {
		year, _ := strconv.Atoi(m[1])
		mon, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])
		hh, _ := strconv.Atoi(m[4])
		mm, _ := strconv.Atoi(m[5])
		ss := 0
		if len(m) > 6 && m[6] != "" {
			ss, _ = strconv.Atoi(m[6])
		}
		t := time.Date(year, time.Month(mon), day, hh, mm, ss, 0, time.UTC)
		loc := t.In(time.Local)
		return &t, &loc
	}
	// Fallback: parenthetical with US time band like (MM.DD H:mmET)
	if m := reParenTZ.FindStringSubmatch(title); m != nil {
		mon, _ := strconv.Atoi(m[1])
		day, _ := strconv.Atoi(m[2])
		hh, _ := strconv.Atoi(m[3])
		mm, _ := strconv.Atoi(m[4])
		tz := strings.ToUpper(m[5])
		if loc := resolveUSTimeBand(tz); loc != nil {
			tLocalBand := time.Date(fallbackYear, time.Month(mon), day, hh, mm, 0, 0, loc)
			tUTC := tLocalBand.UTC()
			tClient := tUTC.In(time.Local)
			return &tUTC, &tClient
		}
	}
	return nil, nil
}

func resolveUSTimeBand(tz string) *time.Location {
	// Map common US time band labels and abbreviations to IANA locations
	// Using representative cities so DST is applied correctly when applicable.
	switch tz {
	case "ET", "EST", "EDT":
		if l, err := time.LoadLocation("America/New_York"); err == nil {
			return l
		}
	case "CT", "CST", "CDT":
		if l, err := time.LoadLocation("America/Chicago"); err == nil {
			return l
		}
	case "MT", "MST", "MDT":
		// Phoenix (America/Phoenix) does not observe DST; most MT does.
		if l, err := time.LoadLocation("America/Denver"); err == nil {
			return l
		}
	case "PT", "PST", "PDT":
		if l, err := time.LoadLocation("America/Los_Angeles"); err == nil {
			return l
		}
	case "AKT", "AKST", "AKDT":
		if l, err := time.LoadLocation("America/Anchorage"); err == nil {
			return l
		}
	case "HST", "HDT", "HT":
		if l, err := time.LoadLocation("Pacific/Honolulu"); err == nil {
			return l
		}
	}
	return nil
}

func sanitizeForFilename(s string) string {
	if s == "" {
		return "filtered"
	}
	// Replace any non-alphanumeric with underscore
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

func writeFilteredM3U(outPath string, entries []PlaylistEntry) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	if _, err := w.WriteString("#EXTM3U\n"); err != nil {
		return err
	}
	for _, e := range entries {
		line := e.Info.Raw
		if line == "" {
			// Reconstruct minimal EXTINF if raw was not preserved
			line = fmt.Sprintf("#EXTINF:%d,%s", e.Info.Duration, e.Info.Title)
		}
		if _, err := w.WriteString(line + "\n"); err != nil {
			return err
		}
		if _, err := w.WriteString(e.URI + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var (
		flagGroupTitle string
		flagOut        string
		flagStrict     bool
	)
	flag.StringVar(&flagGroupTitle, "group-title", "", "Filter entries by group-title (case-insensitive). If empty, include all.")
	flag.StringVar(&flagOut, "out", "", "Output .m3u path. Defaults to '<input>.<group>.m3u' in the same directory.")
	flag.BoolVar(&flagStrict, "strict", false, "Enable strict parsing and fail on malformed lines.")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: iptv-m3u-enhancer [--group-title \"<name>\"] [--out <path>] [--strict] <input.m3u>")
		os.Exit(2)
	}
	inPath := args[0]
	pl, err := parseM3U(inPath, flagStrict)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	// Filter entries by group-title (if provided)
	var filtered []PlaylistEntry
	if flagGroupTitle == "" {
		filtered = pl.Entries
	} else {
		want := strings.ToLower(flagGroupTitle)
		for _, e := range pl.Entries {
			if strings.ToLower(e.Info.GroupTitle()) == want {
				filtered = append(filtered, e)
			}
		}
	}

	// Derive default output path if needed
	outPath := flagOut
	if outPath == "" {
		dir := filepath.Dir(inPath)
		base := filepath.Base(inPath)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		suffix := sanitizeForFilename(flagGroupTitle)
		outPath = filepath.Join(dir, fmt.Sprintf("%s.%s%s", name, suffix, ext))
	}

	if err := writeFilteredM3U(outPath, filtered); err != nil {
		fmt.Fprintln(os.Stderr, "write error:", err)
		os.Exit(1)
	}
}
