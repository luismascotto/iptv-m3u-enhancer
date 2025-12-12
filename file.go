package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

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

func writeFilteredM3U(outPath string, entries []PlaylistEntry, ignoreRaw bool) error {
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
		if !ignoreRaw && e.Info.StartTimeLocal != nil && line != "" && strings.HasPrefix(line, "#EXTINF:") {
			line = rewriteRawExtinfTitleWithLocalTime(line, *e.Info.StartTimeLocal)
		}
		//If we are ignoring the raw EXTINF or the raw EXTINF is empty, write the new EXTINF
		if ignoreRaw || line == "" {
			line = writeNewExtinf(e)
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
