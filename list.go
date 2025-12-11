package main

import (
	"sort"
	"strings"
	"time"
)

func sortEntries(entries []PlaylistEntry) {
	sort.Slice(entries, func(i, j int) bool {
		a := entries[i]
		b := entries[j]
		at := a.Info.StartTimeLocal
		bt := b.Info.StartTimeLocal
		switch {
		case at != nil && bt != nil:
			if at.Equal(*bt) {
				// On the same time, sort by nba-match-id if it exists
				if a.Info.NBAMatchId() != "" && b.Info.NBAMatchId() != "" {
					return a.Info.NBAMatchId() < b.Info.NBAMatchId()
				}
				// If no nba-match-id, sort by title, case-insensitive
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
	cutoff := time.Now().Add(-maxAge)
	out := entries[:0]
	for _, e := range entries {
		if e.Info.StartTimeLocal != nil && !e.Info.StartTimeLocal.Before(cutoff) {
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
