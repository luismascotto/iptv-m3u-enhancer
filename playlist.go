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

func (p *Playlist) filterRemoveWithTitle(substrs []string) {
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

type Cleanser struct {
	Remove        string
	WithSubstring string
	Olds          []string
	New           string
}

func (p *Playlist) cleanseTitles(cleansers []Cleanser) {
	for f := range p.Entries {
		for _, cleanser := range cleansers {
			if cleanser.Remove != "" {
				p.Entries[f].Info.Title = strings.ReplaceAll(p.Entries[f].Info.Title, cleanser.Remove, "")
				continue
			}
			if len(cleanser.Olds) == 0 {
				p.Entries[f].Info.Title = strings.ReplaceAll(p.Entries[f].Info.Title, cleanser.WithSubstring, cleanser.New)
				continue
			}
			if strings.Contains(p.Entries[f].Info.Title, cleanser.WithSubstring) {
				for _, old := range cleanser.Olds {
					p.Entries[f].Info.Title = strings.ReplaceAll(p.Entries[f].Info.Title, old, cleanser.New)
				}
				//fmt.Println(p.Entries[f].Info.Title)
			}
		}
	}
}

func (p *Playlist) processNBAEntries(fallbackYear int) {

	for n := range p.Entries {
		titleGroups := parseTitle(p.Entries[n].Info.TitleCopy)
		if titleGroups == nil {
			continue
		}
		match := parseNBAMatch(titleGroups, fallbackYear)
		if match == nil {
			continue
		}
		p.Entries[n].Info.StartTimeLocal = match.StartTime
		p.Entries[n].Info.Title = fmt.Sprintf("%s | %s (%s) vs %s (%s) > %s",
			match.Channel,
			match.Team1.TeamName,
			match.Team1.Acronym,
			match.Team2.TeamName,
			match.Team2.Acronym,
			match.StartTime.Local().Format("01/02 15:04"))
	}
}
