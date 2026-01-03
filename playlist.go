package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func (p *Playlist) sortEntries() {
	sort.Slice(p.Entries, func(i, j int) bool {
		a := p.Entries[i]
		b := p.Entries[j]
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

func (p *Playlist) filterScheduledEntries(withLocalTime, applyRange bool, expiredAfter, includeUntil time.Duration) {
	past := time.Now().Add(-expiredAfter)
	future := time.Now().Add(includeUntil)
	out := p.Entries[:0]
	for _, e := range p.Entries {
		if e.Info.StartTimeLocal == nil && withLocalTime {
			continue
		}
		if applyRange && e.Info.StartTimeLocal != nil &&
			(e.Info.StartTimeLocal.Before(past) || e.Info.StartTimeLocal.After(future)) {
			continue
		}
		out = append(out, e)
	}
	p.Entries = out
}

func (p *Playlist) filterExcludeTitles(substrs []string) {
	out := p.Entries[:0]
	for _, e := range p.Entries {
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
	p.Entries = out
}

func (p *Playlist) cleanseAwayHomeStream() {
	for f := range p.Entries {
		if strings.Contains(p.Entries[f].Info.Title, "Away") {
			p.Entries[f].Info.Title = strings.ReplaceAll(p.Entries[f].Info.Title, "| Away Stream", "(A)")
			p.Entries[f].Info.Title = strings.ReplaceAll(p.Entries[f].Info.Title, "(Away)", "(A)")
			fmt.Println(p.Entries[f].Info.Title)
		}
		if strings.Contains(p.Entries[f].Info.Title, "Home") {
			p.Entries[f].Info.Title = strings.ReplaceAll(p.Entries[f].Info.Title, "| Home Stream", "(H)")
			p.Entries[f].Info.Title = strings.ReplaceAll(p.Entries[f].Info.Title, "(Home)", "(H)")
			fmt.Println(p.Entries[f].Info.Title)
		}
	}
}
