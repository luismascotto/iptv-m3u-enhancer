package main

import "testing"

func TestReTitle1_MatchAndGroups(t *testing.T) {
	title := "USA | NBA 03: Atlanta Hawks vs Toronto Raptors | Sat 3rd Jan 7:30PM ET"
	m := reTitle1.FindStringSubmatch(title)
	if m == nil {
		t.Fatalf("reTitle1 did not match: %q", title)
	}
	if len(m) != 5 {
		t.Fatalf("reTitle1 expected 5 groups (including full match), got %d: %#v", len(m), m)
	}
	if m[1] != "USA | NBA 03" {
		t.Errorf("group1 (channel) = %q, want %q", m[1], "USA | NBA 03")
	}
	if m[2] != "Atlanta Hawks" {
		t.Errorf("group2 (team1) = %q, want %q", m[2], "Atlanta Hawks")
	}
	if m[3] != "Toronto Raptors" {
		t.Errorf("group3 (team2) = %q, want %q", m[3], "Toronto Raptors")
	}
	if m[4] != "Sat 3rd Jan 7:30PM ET" {
		t.Errorf("group4 (start time) = %q, want %q", m[4], "Sat 3rd Jan 7:30PM ET")
	}
}

func TestReTitle2_MatchAndGroups(t *testing.T) {
	title := "NBA 27: Rockets vs Clippers (Home) (12.23 5:30PM ET)"
	m := reTitle2.FindStringSubmatch(title)
	if m == nil {
		t.Fatalf("reTitle2 did not match: %q", title)
	}
	if len(m) != 6 {
		t.Fatalf("reTitle2 expected 6 groups (including full match), got %d: %#v", len(m), m)
	}
	if m[1] != "NBA 27" {
		t.Errorf("group1 (channel) = %q, want %q", m[1], "NBA 27")
	}
	if m[2] != "Rockets" {
		t.Errorf("group2 (team1) = %q, want %q", m[2], "Rockets")
	}
	if m[3] != "Clippers" {
		t.Errorf("group3 (team2) = %q, want %q", m[3], "Clippers")
	}
	if m[4] != "Home" {
		t.Errorf("group4 (stream type) = %q, want %q", m[4], "Home")
	}
	if m[5] != "12.23 5:30PM ET" {
		t.Errorf("group5 (start time) = %q, want %q", m[5], "12.23 5:30PM ET")
	}
}

func TestReTitle21_MatchAndGroups(t *testing.T) {
	title := "NBA 27: Rockets vs Clippers | Away Stream | (12.23 5:30PM ET)"
	m := reTitle21.FindStringSubmatch(title)
	if m == nil {
		t.Fatalf("reTitle21 did not match: %q", title)
	}
	if len(m) != 6 {
		t.Fatalf("reTitle21 expected 6 groups (including full match), got %d: %#v", len(m), m)
	}
	if m[1] != "NBA 27" {
		t.Errorf("group1 (channel) = %q, want %q", m[1], "NBA 27")
	}
	if m[2] != "Rockets" {
		t.Errorf("group2 (team1) = %q, want %q", m[2], "Rockets")
	}
	if m[3] != "Clippers" {
		t.Errorf("group3 (team2) = %q, want %q", m[3], "Clippers")
	}
	if m[4] != "Away" {
		t.Errorf("group4 (stream type) = %q, want %q", m[4], "Away")
	}
	if m[5] != "(12.23 5:30PM ET)" {
		t.Errorf("group5 (start time) = %q, want %q", m[5], "(12.23 5:30PM ET)")
	}
}

func TestReTitle3_MatchAndGroups(t *testing.T) {
	title := "NBA 10 : Spurs (SAS) x Cavaliers (CLE) start:2025 12 30 00:50:00 stop:2025 12 30 04:50:00"
	m := reTitle3.FindStringSubmatch(title)
	if m == nil {
		t.Fatalf("reTitle3 did not match: %q", title)
	}
	if len(m) != 6 {
		t.Fatalf("reTitle3 expected 6 groups (including full match), got %d: %#v", len(m), m)
	}
	if m[1] != "NBA 10 " {
		t.Errorf("group1 (channel) = %q, want %q", m[1], "NBA 10 ")
	}
	if m[2] != "Spurs (SAS)" {
		t.Errorf("group2 (team1) = %q, want %q", m[2], "Spurs (SAS)")
	}
	if m[3] != "Cavaliers (CLE)" {
		t.Errorf("group3 (team2) = %q, want %q", m[3], "Cavaliers (CLE)")
	}
	if m[4] != "2025 12 30 00:50:00" {
		t.Errorf("group4 (start) = %q, want %q", m[4], "2025 12 30 00:50:00")
	}
	if m[5] != "2025 12 30 04:50:00" {
		t.Errorf("group5 (stop) = %q, want %q", m[5], "2025 12 30 04:50:00")
	}
}

func TestReTitle4_MatchAndGroups(t *testing.T) {
	title := "NBA 02 : Brooklyn Nets @ Washington Wizards // UK Fri 2 Jan 11:45pm // ET Fri 2 Jan 6:45pm"
	m := reTitle4.FindStringSubmatch(title)
	if m == nil {
		t.Fatalf("reTitle4 did not match: %q", title)
	}
	if len(m) != 6 {
		t.Fatalf("reTitle4 expected 6 groups (including full match), got %d: %#v", len(m), m)
	}
	if m[1] != "NBA 02 " {
		t.Errorf("group1 (channel) = %q, want %q", m[1], "NBA 02 ")
	}
	if m[2] != "Brooklyn Nets" {
		t.Errorf("group2 (team1) = %q, want %q", m[2], "Brooklyn Nets")
	}
	if m[3] != "Washington Wizards" {
		t.Errorf("group3 (team2) = %q, want %q", m[3], "Washington Wizards")
	}
	if m[4] != "UK Fri 2 Jan 11:45pm" {
		t.Errorf("group4 (start time 1) = %q, want %q", m[4], "UK Fri 2 Jan 11:45pm")
	}
	if m[5] != "ET Fri 2 Jan 6:45pm" {
		t.Errorf("group5 (start time 2) = %q, want %q", m[5], "ET Fri 2 Jan 6:45pm")
	}
}

