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
				if a.Info.NBAMatchId() != "" && b.Info.NBAMatchId() != "" && a.Info.NBAMatchId() != b.Info.NBAMatchId() {
					return a.Info.NBAMatchId() < b.Info.NBAMatchId()
				}

				// If no nba-match-id (or equal), sort by title (after colon when present), case-insensitive
				_, ai, _ := strings.Cut(a.Info.Title, ":")
				_, bi, _ := strings.Cut(b.Info.Title, ":")
				ai = strings.ToLower(ai)
				bi = strings.ToLower(bi)
				if ai == bi {
					return a.Info.Title < b.Info.Title
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
	var matchIdStartTimeMap = make(map[string]time.Time)
	for n := range p.Entries {
		titleGroups := parseTitle(p.Entries[n].Info.TitleCopy)
		if titleGroups == nil {
			continue
		}
		match := parseNBAMatch(titleGroups, fallbackYear)
		if match == nil {
			continue
		}
		t := match.StartTime.Add(time.Duration(roundUpMinutesToHourOrHalf(match.StartTime.Minute())) * time.Minute)
		tLocal := t.In(time.Local)
		p.Entries[n].Info.StartTimeLocal = &tLocal

		streamType := ""
		if match.StreamType != "" {
			streamType = fmt.Sprintf("%c ", match.StreamType[0])
		}
		p.Entries[n].Info.Title = fmt.Sprintf("%s: %s (%s) vs %s (%s) %s> ",
			match.Channel,
			match.Team1.TeamName,
			match.Team1.Acronym,
			match.Team2.TeamName,
			match.Team2.Acronym,
			streamType)
		p.Entries[n].Info.SetNBAMatchId()

		matchId := p.Entries[n].Info.NBAMatchId()

		// update matchIdStartTimeMap with the latest start time
		if matchId != "" {
			if _, ok := matchIdStartTimeMap[matchId]; !ok {
				matchIdStartTimeMap[matchId] = tLocal
			} else {
				if tLocal.After(matchIdStartTimeMap[matchId]) {
					matchIdStartTimeMap[matchId] = tLocal
				}
			}
		}
	}

	// Finalize the titles with the latest start time + Home/Away Stream info
	for n := range p.Entries {
		matchId := p.Entries[n].Info.NBAMatchId()
		if matchId != "" {
			if tLocal, ok := matchIdStartTimeMap[matchId]; ok {
				diffDays := int(time.Since(tLocal).Hours() / 24)
				suffix := ""
				if diffDays != 0 {
					// Add + or - to indicate past or future
					suffix = fmt.Sprintf(" (%+d)", diffDays)
				}
				p.Entries[n].Info.StartTimeLocal = &tLocal
				p.Entries[n].Info.Title = fmt.Sprintf("%s%s%s",
					p.Entries[n].Info.Title,
					tLocal.Format("15:04"),
					suffix)
			}
		}
	}
}
