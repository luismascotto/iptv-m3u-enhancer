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

func writeFilteredM3U(outPath string, entries []PlaylistEntry) error {
	var err error
	var f *os.File
	if err = os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	if f, err = os.Create(outPath); err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	if _, err = w.WriteString("#EXTM3U\n"); err != nil {
		return err
	}
	for _, e := range entries {
		line := writeNewEntry(e)
		if line == "" {
			continue
			// line = e.Info.Raw
			// if line == "" {
			// 	continue
			// }
		}

		if _, err = w.WriteString(line); err != nil {
			return err
		}
		// if _, err = w.WriteString("\n"); err != nil {
		// 	return err
		// }
		// if _, err = w.WriteString(e.URI); err != nil {
		// 	return err
		// }
		// if _, err = w.WriteString("\n"); err != nil {
		// 	return err
		// }
	}
	return nil
}
