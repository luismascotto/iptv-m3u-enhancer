package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
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
	// 4) DayOfWeek DD(st|nd|rd|th) Month HH:mm TZ
	// Example: Tue 9th Dec 6:00PM ET
	reDowDomMonth = regexp.MustCompile(`(?i)(Mon|Tue|Wed|Thu|Fri|Sat|Sun)\s+(\d{1,2})(th|nd|rd|st)\s+(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+(\d{1,2}):(\d{2})(AM|PM)\s*([A-Z]{1,4})`)
	// Date in filename/path: YYYY-MM-DD
	reDateInPath = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
)

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
		day, _ := strconv.Atoi(m[2])
		mon, _ := strconv.Atoi(m[1])
		hh, _ := strconv.Atoi(m[3])
		mm, _ := strconv.Atoi(m[4])
		tz := strings.ToUpper(m[5])
		return getTimeFromLocation(tz, fallbackYear, mon, day, hh, mm)
	}
	// // Fallback2: day of week, day of month, month, hour:minute, time band
	if m := reDowDomMonth.FindStringSubmatch(title); m != nil {
		//dow := strings.ToUpper(m[1])
		day, _ := strconv.Atoi(m[2])
		mon := getMonthNumber(m[4])
		hh, _ := strconv.Atoi(m[5])
		mm, _ := strconv.Atoi(m[6])
		ampm := strings.ToUpper(m[7])
		tz := strings.ToUpper(m[8])
		hh = to24h(hh, ampm)
		return getTimeFromLocation(tz, fallbackYear, mon, day, hh, mm)
	}
	return nil
}

func getMonthNumber(month string) int {

	switch month {
	case "Jan":
		return 1
	case "Feb":
		return 2
	case "Mar":
		return 3
	case "Apr":
		return 4
	case "May":
		return 5
	case "Jun":
		return 6
	case "Jul":
		return 7
	case "Aug":
		return 8
	case "Sep":
		return 9
	case "Oct":
		return 10
	case "Nov":
		return 11
	case "Dec":
		return 12
	}
	return 0
}

func getTimeFromLocation(tz string, year int, mon int, day int, hh int, mm int) *time.Time {
	var loc *time.Location
	if loc = resolveUSTimeBand(tz); loc == nil {
		loc = time.UTC
	}
	tLocation := time.Date(year, time.Month(mon), day, hh, mm, 0, 0, loc)
	// Round up
	tLocation = tLocation.Add(time.Duration(roundUpMinutesToHourOrHalf(tLocation.Minute())) * time.Minute)
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

func roundUpMinutesToHourOrHalf(minutes int) int {
	if minutes >= 50 {
		return (60 - minutes)
	}
	if minutes < 30 && minutes >= 20 {
		return (30 - minutes)
	}
	return 0
}
