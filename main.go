package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	// stop:YYYY MM DD HH:mm(:SS)?
	reStopInTitle = regexp.MustCompile(`(?i)stop:\s*(\d{4})\s+(\d{1,2})\s+(\d{1,2})\s+(\d{1,2}):(\d{2})(?::(\d{2}))?`)
	// 2) (MM.DD H:mmTZ) where TZ is a US time band like ET, CT, MT, PT (also EST/EDT, etc.)
	reParenTZ = regexp.MustCompile(`(?i)\((\d{1,2})\.(\d{1,2})\s+(\d{1,2}):(\d{2})\s*([A-Z]{1,4})\)`)
	// 2b) (MM.DD h:mm(AM|PM) TZ)
	reParenTZ12 = regexp.MustCompile(`(?i)\((\d{1,2})\.(\d{1,2})\s+(\d{1,2}):(\d{2})\s*(AM|PM)\s*([A-Z]{1,4})\)`)
	// 3) | MM/DD/YYYY h:mm (AM|PM) TZ
	rePipeDate12 = regexp.MustCompile(`(?i)\|\s*(\d{1,2})\/(\d{1,2})\/(\d{4})\s+(\d{1,2}):(\d{2})\s*(AM|PM)\s*([A-Z]{1,4})`)
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
	if local := parseTimesFromTitle(title, fallbackYear); local != nil {
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

func parseTimesFromTitle(title string, fallbackYear int) *time.Time {
	// Prefer explicit 'start:' form if present
	if m := reStartInTitle.FindStringSubmatch(title); m != nil {
		year, _ := strconv.Atoi(m[1])
		mon, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])
		hh, _ := strconv.Atoi(m[4])
		mm, _ := strconv.Atoi(m[5])
		// ss := 0
		// if len(m) > 6 && m[6] != "" {
		// 	ss, _ = strconv.Atoi(m[6])
		// }
		//t := time.Date(year, time.Month(mon), day, hh, mm, ss, 0, time.UTC)
		return getTimeFromLocation("", year, mon, day, hh, mm)
	}
	// Pipe format with explicit date and AM/PM: | MM/DD/YYYY h:mm AM TZ
	if m := rePipeDate12.FindStringSubmatch(title); m != nil {
		mon, _ := strconv.Atoi(m[1])
		day, _ := strconv.Atoi(m[2])
		year, _ := strconv.Atoi(m[3])
		hh, _ := strconv.Atoi(m[4])
		mm, _ := strconv.Atoi(m[5])
		ampm := strings.ToUpper(m[6])
		tz := strings.ToUpper(m[7])
		hh = to24h(hh, ampm)
		return getTimeFromLocation(tz, year, mon, day, hh, mm)
	}
	// Parenthetical with AM/PM and US time band: (MM.DD h:mmPM TZ)
	if m := reParenTZ12.FindStringSubmatch(title); m != nil {
		mon, _ := strconv.Atoi(m[1])
		day, _ := strconv.Atoi(m[2])
		hh, _ := strconv.Atoi(m[3])
		mm, _ := strconv.Atoi(m[4])
		ampm := strings.ToUpper(m[5])
		tz := strings.ToUpper(m[6])
		hh = to24h(hh, ampm)
		return getTimeFromLocation(tz, fallbackYear, mon, day, hh, mm)
	}
	// Fallback: parenthetical with US time band like (MM.DD H:mmET)
	if m := reParenTZ.FindStringSubmatch(title); m != nil {
		mon, _ := strconv.Atoi(m[1])
		day, _ := strconv.Atoi(m[2])
		hh, _ := strconv.Atoi(m[3])
		mm, _ := strconv.Atoi(m[4])
		tz := strings.ToUpper(m[5])
		return getTimeFromLocation(tz, fallbackYear, mon, day, hh, mm)
	}
	return nil
}

func getTimeFromLocation(tz string, year int, mon int, day int, hh int, mm int) *time.Time {
	var loc *time.Location
	if loc = resolveUSTimeBand(tz); loc == nil {
		loc = time.UTC
	}
	tLocation := time.Date(year, time.Month(mon), day, hh, mm, 0, 0, loc)
	// Round up
	tLocation = tLocation.Add(time.Duration(RoundUpMinutesTo60or30(tLocation.Minute())) * time.Minute)
	tClient := tLocation.In(time.Local)
	return &tClient
}

func to24h(hour12 int, ampm string) int {
	h := hour12 % 12
	if strings.EqualFold(ampm, "PM") {
		h += 12
	}
	return h
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
		// If we have a parsed local start time, rewrite the title segment with standardized local time
		if e.Info.StartTimeLocal != nil && line != "" && strings.HasPrefix(line, "#EXTINF:") {
			line = rewriteExtinfTitleWithLocalTime(line, *e.Info.StartTimeLocal)
		} else if line == "" {
			// Reconstruct minimal EXTINF if raw was not preserved
			title := e.Info.Title
			if e.Info.StartTimeLocal != nil {
				title = replaceStartTimeTokens(title, *e.Info.StartTimeLocal)
			}
			line = fmt.Sprintf("#EXTINF:%d,%s", e.Info.Duration, title)
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

func rewriteExtinfTitleWithLocalTime(rawLine string, local time.Time) string {
	const prefix = "#EXTINF:"
	if !strings.HasPrefix(rawLine, prefix) {
		return rawLine
	}
	payload := rawLine[len(prefix):]
	// Find first comma not inside quotes to split meta and title, preserving original spacing
	inQuotes := false
	split := -1
	for i := 0; i < len(payload); i++ {
		ch := payload[i]
		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}
		if ch == ',' && !inQuotes {
			split = i
			break
		}
	}
	if split == -1 {
		return rawLine
	}
	meta := payload[:split] // keep as-is
	title := payload[split+1:]
	newTitle := replaceStartTimeTokens(title, local)
	return prefix + meta + "," + newTitle
}

func replaceStartTimeTokens(title string, local time.Time) string {
	// 1) Remove all recognizable time tokens from title
	res := title
	res = rePipeDate12.ReplaceAllString(res, "")
	res = reParenTZ12.ReplaceAllString(res, "")
	res = reParenTZ.ReplaceAllString(res, "")
	res = reStartInTitle.ReplaceAllString(res, "")
	res = reStopInTitle.ReplaceAllString(res, "")
	// 2) Cleanup separators and spaces left behind
	res = strings.TrimSpace(res)
	// Drop dangling separators at end
	for {
		trimmed := strings.TrimRight(res, " \t")
		if strings.HasSuffix(trimmed, "|") || strings.HasSuffix(trimmed, "-") ||
			strings.HasSuffix(trimmed, ":") || strings.HasSuffix(trimmed, ";") ||
			strings.HasSuffix(trimmed, ",") {
			res = strings.TrimRight(strings.TrimRight(trimmed, "|-:;, "), " \t")
			continue
		}
		break
	}
	// Condense multiple inner spaces
	res = regexp.MustCompile(`\s{2,}`).ReplaceAllString(res, " ")
	// 3) Append standardized local time suffix " > DD/MM HH:mm"
	return strings.TrimSpace(res) + " > " + local.Format("02/01 15:04")
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

	// Remove entries with start times earlier than 6 hours ago (keep entries without time)
	filtered = filterRecentEntries(filtered, 6*time.Hour)

	// Remove entries with undesired titles
	filtered = filterExcludeTitles(filtered, []string{"no event", "offline", "no games", "no scheduled"})

	// Sort: by parsed local start time (items with time first, earlier first), then by title
	sortEntries(filtered)

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

func sortEntries(entries []PlaylistEntry) {
	sort.Slice(entries, func(i, j int) bool {
		a := entries[i]
		b := entries[j]
		at := a.Info.StartTimeLocal
		bt := b.Info.StartTimeLocal
		switch {
		case at != nil && bt != nil:
			if at.Equal(*bt) {
				// Tie-breaker: title, case-insensitive
				ai := strings.ToLower(a.Info.Title)
				bi := strings.ToLower(b.Info.Title)
				if ai == bi {
					if strings.Contains(ai, "@") {
						return true
					}
					if strings.Contains(bi, "@") {
						return false
					}
					return false
				}
				return ai < bi
			}
			return at.Before(*bt)
		case at != nil && bt == nil:
			// Items with time come first
			return true
		case at == nil && bt != nil:
			return false
		default:
			// Both without time: sort by title
			ai := strings.ToLower(a.Info.Title)
			bi := strings.ToLower(b.Info.Title)
			if ai == bi {
				return a.Info.Title < b.Info.Title
			}
			return ai < bi
		}
	})
}

func filterRecentEntries(entries []PlaylistEntry, maxAge time.Duration) []PlaylistEntry {
	cutoff := time.Now().UTC().Add(-maxAge)
	out := entries[:0]
	for _, e := range entries {
		if e.Info.StartTimeLocal == nil || !e.Info.StartTimeLocal.Before(cutoff) {
			out = append(out, e)
		}
	}
	return out
}

func filterExcludeTitles(entries []PlaylistEntry, substrs []string) []PlaylistEntry {
	out := entries[:0]
	for _, e := range entries {
		titleLower := strings.ToLower(e.Info.Title)
		exclude := false
		for _, sub := range substrs {
			if strings.Contains(titleLower, sub) {
				exclude = true
				break
			}
		}
		if !exclude {
			out = append(out, e)
		}
	}
	return out
}

func RoundUpMinutesTo60or30(minutes int) int {
	if minutes >= 50 {
		return (60 - minutes)
	}
	if minutes < 30 && minutes >= 20 {
		return (30 - minutes)
	}
	return 0
}
