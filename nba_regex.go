package main

import (
	"regexp"
	"strings"
	"time"
)

var (
	// Patterns for extracting teams and start time from title. Times will be parsed later to handle different formats/zones.
	// Title patterns:
	// 1) <channel name>: <team1> vs/x/@ <team2> | <start time>
	// Example: USA | NBA 03: Atlanta Hawks vs Toronto Raptors | Sat 3rd Jan 7:30PM ET
	reTitle1 = regexp.MustCompile(`(?i)(.*): (.*) (?:vs|x|@) (.*) \| (.*)`)

	// 2) <channel name>: <team1> vs/x/@ <team2> (<stream type>) (<start time>)
	// Example: NBA 27: Rockets vs Clippers (Home) (12.23 5:30PM ET)
	reTitle2 = regexp.MustCompile(`(?i)(.*): (.*) (?:vs|x|@) (.*) \((.*)\) \((.*)\)`)

	// 2.1) <channel name>: <team1> vs/x/@ <team2> | <stream type> Stream | (<start time>)
	// Example: NBA 27: Rockets vs Clippers | Away Stream | (12.23 5:30PM ET)
	reTitle21 = regexp.MustCompile(`(?i)(.*): (.*) (?:vs|x|@) (.*) \| (.*) Stream \| (.*)`)

	// 3) <channel name>: <team1> vs/x/@ <team2> start:<start time> stop:<stop time>
	//Example: NBA 10 : Spurs (SAS) x Cavaliers (CLE) start:2025 12 30 00:50:00 stop:2025 12 30 04:50:00
	reTitle3 = regexp.MustCompile(`(?i)(.*): (.*) (?:vs|x|@) (.*) start:(.*) stop:(.*)`)

	// 3) <channel name>: <team1> vs/x/@ <team2> // <start time> // <start time>
	//Example: NBA 02 : Brooklyn Nets @ Washington Wizards // UK Fri 2 Jan 11:45pm // ET Fri 2 Jan 6:45pm
	reTitle4 = regexp.MustCompile(`(?i)(.*): (.*) (?:vs|x|@) (.*) \/\/ (.*) \/\/ (.*)`)

	titleRegexes = []TitleRegex{
		{
			Regex:  reTitle1,
			Format: "<channel name>: <team1> vs/x/@ <team2> | <start time>",
			Groups: []string{"channel", "team1", "team2", "start time"},
		},
		{
			Regex:  reTitle2,
			Format: "<channel name>: <team1> vs/x/@ <team2> (<stream type>) (<start time>)",
			Groups: []string{"channel", "team1", "team2", "stream type", "start time"},
		},
		{
			Regex:  reTitle21,
			Format: "<channel name>: <team1> vs/x/@ <team2> | <stream type> Stream | (<start time>)",
			Groups: []string{"channel", "team1", "team2", "stream type", "start time"},
		},
		{
			Regex:  reTitle3,
			Format: "<channel name>: <team1> vs/x/@ <team2> start:<start time> stop:<stop time>",
			Groups: []string{"channel", "team1", "team2", "start time", "stop time"},
		},
		{
			Regex:  reTitle4,
			Format: "<channel name>: <team1> vs/x/@ <team2> // <uk start time> // <start time>",
			Groups: []string{"channel", "team1", "team2", "uk start time", "start time"},
		},
	}
)

type TitleRegex struct {
	Regex  *regexp.Regexp
	Format string
	Groups []string
}

type NBATitleGroups struct {
	Channel    string
	Team1      string
	Team2      string
	StreamType string
	StartTime  string
}

var nbaTitleGroupKeys = NBATitleGroups{
	Channel:    "channel",
	Team1:      "team1",
	Team2:      "team2",
	StreamType: "stream type",
	StartTime:  "start time",
}

type NBAMatch struct {
	Channel    string
	Team1      NBAFranchise
	Team2      NBAFranchise
	StreamType string
	StartTime  *time.Time
	// StartTime2 *time.Time
	// EndTime   *time.Time
}

func parseTitle(title string) *NBATitleGroups {
	for _, regex := range titleRegexes {
		if m := regex.Regex.FindStringSubmatch(title); m != nil {
			mapGroups := make(map[string]string)
			for i, group := range m {
				if i > len(regex.Groups) {
					break
				}
				if i > 0 {
					mapGroups[regex.Groups[i-1]] = strings.TrimSpace(group)
				}
			}
			if mapGroups[nbaTitleGroupKeys.Team1] == "" ||
				mapGroups[nbaTitleGroupKeys.Team2] == "" ||
				mapGroups[nbaTitleGroupKeys.StartTime] == "" {
				continue
			}

			return &NBATitleGroups{
				Channel:    mapGroups[nbaTitleGroupKeys.Channel],
				Team1:      mapGroups[nbaTitleGroupKeys.Team1],
				Team2:      mapGroups[nbaTitleGroupKeys.Team2],
				StreamType: mapGroups[nbaTitleGroupKeys.StreamType],
				StartTime:  mapGroups[nbaTitleGroupKeys.StartTime],
			}
		}
	}
	return nil
}

func parseNBAMatch(titleGroups *NBATitleGroups, fallbackYear int) *NBAMatch {
	team1 := parseNBAFranchise(titleGroups.Team1)
	team2 := parseNBAFranchise(titleGroups.Team2)
	if team1 == nil || team2 == nil {
		return nil
	}
	titleGroups.StartTime = strings.ReplaceAll(titleGroups.StartTime, "st ", " ")
	titleGroups.StartTime = strings.ReplaceAll(titleGroups.StartTime, "nd ", " ")
	titleGroups.StartTime = strings.ReplaceAll(titleGroups.StartTime, "rd ", " ")
	titleGroups.StartTime = strings.ReplaceAll(titleGroups.StartTime, "th ", " ")
	startTime := parseTimesFromTitleV2(titleGroups.StartTime)
	if startTime == nil {
		return nil
	}

	if startTime.Year() == 0 {
		// new time with fallback year
		t := time.Date(fallbackYear, startTime.Month(), startTime.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), startTime.Location())
		startTime = &t
	}

	return &NBAMatch{
		Channel:    titleGroups.Channel,
		Team1:      *team1,
		Team2:      *team2,
		StreamType: titleGroups.StreamType,
		StartTime:  startTime,
	}
}

func parseNBAFranchise(name string) *NBAFranchise {
	for _, franchise := range NBAFranchises {
		if strings.Contains(name, string(franchise.Name)) ||
			strings.Contains(name, franchise.TeamName) ||
			strings.Contains(name, "("+franchise.Acronym+")") ||
			(franchise.AcronymAlt != "" && strings.Contains(name, "("+franchise.AcronymAlt+")")) {
			return &franchise
		}
	}
	return nil
}
