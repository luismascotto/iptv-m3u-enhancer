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
	//Raw        string
	// Parsed times (when present in title). UTC source converted to local as well.
	StartTimeLocal *time.Time
	TitleCopy      string
}

func (e ExtInf) GroupTitle() string {
	if e.Attributes == nil {
		return ""
	}
	return e.Attributes["group-title"]
}

func (e ExtInf) NBAMatchId() string {
	if e.Attributes == nil {
		return ""
	}
	return e.Attributes["nba-match-id"]
}

func (e ExtInf) SetNBAMatchId() {
	nbaMatchId := generateMatchIdFromTitle(e.Title)
	if nbaMatchId != "" {
		if e.Attributes == nil {
			e.Attributes = make(map[string]string)
		}
		e.Attributes["nba-match-id"] = nbaMatchId
	}
}

type PlaylistEntry struct {
	Info ExtInf
	URI  string
}

type Playlist struct {
	Entries       []PlaylistEntry
	HeaderPresent bool
}

func parseM3U(path string, strict bool, groupTitle string) (Playlist, error) {
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
			info, err := parseEXTINF(line)
			if err != nil {
				if strict {
					return Playlist{}, fmt.Errorf("line %d: %w", lineNum, err)
				}
				// Skip malformed EXTINF in non-strict mode
				currentEXTINF = nil
				continue
			}
			// Preserve the raw line for re-emitting later
			//info.Raw = line
			info.TitleCopy = info.Title
			currentEXTINF = &info
			continue
		}

		// Other tags/comments can be ignored for now
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Non-comment non-empty lines should be URIs
		if currentEXTINF != nil {
			if groupTitle != "" && !strings.EqualFold(currentEXTINF.GroupTitle(), groupTitle) {
				continue
			}
			// start := parseTimesFromTitleV2(currentEXTINF.Title)
			// if start != nil {
			// 	currentEXTINF.StartTimeLocal = start
			// } else {

			// 	if start = parseTimesFromTitle(currentEXTINF.Title, fallbackYear); start != nil {
			// 		currentEXTINF.StartTimeLocal = start

			// 		currentEXTINF.Title = replaceStartTimeTokens(currentEXTINF.Title, *start)
			// 	}
			// }

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
)

func parseEXTINF(line string) (ExtInf, error) {
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

// func replaceStartTimeTokens(title string, local time.Time) string {
// 	// 1) Remove all recognizable time tokens from title
// 	res := title
// 	res = rePipeDate12.ReplaceAllString(res, "")
// 	res = reParenTZ12.ReplaceAllString(res, "")
// 	res = reParenTZ.ReplaceAllString(res, "")
// 	res = reStartInTitle.ReplaceAllString(res, "")
// 	res = reStopInTitle.ReplaceAllString(res, "")
// 	res = reDowDomMonth.ReplaceAllString(res, "")
// 	// 2) Cleanup separators and spaces left behind
// 	res = strings.TrimSpace(res)
// 	// Drop dangling separators at end
// 	for {
// 		trimmed := strings.TrimRight(res, " \t")
// 		if strings.HasSuffix(trimmed, "|") || strings.HasSuffix(trimmed, "-") ||
// 			strings.HasSuffix(trimmed, ":") || strings.HasSuffix(trimmed, ";") ||
// 			strings.HasSuffix(trimmed, ",") {
// 			res = strings.TrimRight(strings.TrimRight(trimmed, "|-:;, "), " \t")
// 			continue
// 		}
// 		break
// 	}
// 	// Condense multiple inner spaces
// 	res = regexp.MustCompile(`\s{2,}`).ReplaceAllString(res, " ")
// 	// 3) Append standardized local time suffix " > DD/MM HH:mm"
// 	return strings.TrimSpace(res) + " > " + local.Format("02/01 15:04")
// }

func main() {
	var (
		flagGroupTitle string
		flagOut        string
		flagStrict     bool
		flagStartTime  bool
		flagRecent     bool
		flagNBA        bool
	)
	flag.StringVar(&flagGroupTitle, "group-title", "", "Filter entries by group-title (case-insensitive).")
	flag.StringVar(&flagOut, "out", "", "Output .m3u path. Defaults to '<input>.<group>.m3u' in the same directory.")
	flag.BoolVar(&flagStrict, "strict", false, "Enable strict parsing and fail on malformed lines.")
	flag.BoolVar(&flagStartTime, "start-time", false, "Filter entries with parsed start time.")
	flag.BoolVar(&flagRecent, "recent", false, "Filter entries with start time prior to 6 hours ago or after 24 hours from now")
	flag.BoolVar(&flagNBA, "nba", false, "Parse teams from title to improve sorting by match")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: iptv-m3u-enhancer [--group-title \"<name>\"] [--out <path>] [--strict] [--start-time] [--recent] [--nba] <input.m3u>")
		os.Exit(2)
	}
	inPath := args[0]
	playlist, err := parseM3U(inPath, flagStrict, flagGroupTitle)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	fallbackYear := extractFallbackYear(inPath)
	// Process entries with NBA new generic logic of splitting title into teams and start time
	if flagNBA {
		playlist.processNBAEntries(fallbackYear)
	}

	// Remove entries with undesired titles
	playlist.filterRemoveWithTitle([]string{"no event", "offline", "no games", "no scheduled"})

	// Process entries based on start time information
	if flagStartTime || flagRecent {
		playlist.filterScheduledEntries(flagStartTime, flagRecent, 6*time.Hour, 24*time.Hour)
	}

	if flagNBA {

		playlist.cleanseTitles([]Cleanser{
			{Remove: "ⓧ"},
			{WithSubstring: "USA | NBA", New: "USA"},
			{WithSubstring: " : ", New: ": "},
			{WithSubstring: "Away", Olds: []string{"| Away Stream", "(Away)"}, New: "(A)"},
			{WithSubstring: "Home", Olds: []string{"| Home Stream", "(Home)"}, New: "(H)"},
			//
		})
	}

	// Sort: by parsed local start time (when present) then by title
	playlist.sortEntries()

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

	if err := writeFilteredM3U(outPath, playlist.Entries); err != nil {
		fmt.Fprintln(os.Stderr, "write error:", err)
		os.Exit(1)
	}
}

func writeNewEntry(e PlaylistEntry) string {
	strbExtinf := strings.Builder{}
	strbExtinf.WriteString("#EXTINF:")
	strbExtinf.WriteString(strconv.Itoa(e.Info.Duration))
	for key, value := range e.Info.Attributes {
		strbExtinf.WriteString(" ")
		strbExtinf.WriteString(key)
		strbExtinf.WriteString("=\"")
		strbExtinf.WriteString(value)
		strbExtinf.WriteString("\"")
	}
	strbExtinf.WriteString(",")
	strbExtinf.WriteString(e.Info.Title)
	strbExtinf.WriteString("\n")
	strbExtinf.WriteString(e.URI)
	strbExtinf.WriteString("\n")
	return strbExtinf.String()
}
