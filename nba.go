package main

import (
	"fmt"
	"strings"
)

type NBAFranchiseName string

const (
	NBAFranchiseAtlantaHawks          NBAFranchiseName = "Atlanta Hawks"
	NBAFranchiseBostonCeltics         NBAFranchiseName = "Boston Celtics"
	NBAFranchiseBrooklynNets          NBAFranchiseName = "Brooklyn Nets"
	NBAFranchiseCharlotteHornets      NBAFranchiseName = "Charlotte Hornets"
	NBAFranchiseChicagoBulls          NBAFranchiseName = "Chicago Bulls"
	NBAFranchiseClevelandCavaliers    NBAFranchiseName = "Cleveland Cavaliers"
	NBAFranchiseDallasMavericks       NBAFranchiseName = "Dallas Mavericks"
	NBAFranchiseDenverNuggets         NBAFranchiseName = "Denver Nuggets"
	NBAFranchiseDetroitPistons        NBAFranchiseName = "Detroit Pistons"
	NBAFranchiseGoldenStateWarriors   NBAFranchiseName = "Golden State Warriors"
	NBAFranchiseHoustonRockets        NBAFranchiseName = "Houston Rockets"
	NBAFranchiseIndianaPacers         NBAFranchiseName = "Indiana Pacers"
	NBAFranchiseLosAngelesClippers    NBAFranchiseName = "Los Angeles Clippers"
	NBAFranchiseLosAngelesLakers      NBAFranchiseName = "Los Angeles Lakers"
	NBAFranchiseMemphisGrizzlies      NBAFranchiseName = "Memphis Grizzlies"
	NBAFranchiseMiamiHeat             NBAFranchiseName = "Miami Heat"
	NBAFranchiseMilwaukeeBucks        NBAFranchiseName = "Milwaukee Bucks"
	NBAFranchiseMinnesotaTimberwolves NBAFranchiseName = "Minnesota Timberwolves"
	NBAFranchiseNewOrleansPelicans    NBAFranchiseName = "New Orleans Pelicans"
	NBAFranchiseNewYorkKnicks         NBAFranchiseName = "New York Knicks"
	NBAFranchiseOklahomaCityThunder   NBAFranchiseName = "Oklahoma City Thunder"
	NBAFranchiseOrlandoMagic          NBAFranchiseName = "Orlando Magic"
	NBAFranchisePhiladelphia76ers     NBAFranchiseName = "Philadelphia 76ers"
	NBAFranchisePhoenixSuns           NBAFranchiseName = "Phoenix Suns"
	NBAFranchisePortlandTrailBlazers  NBAFranchiseName = "Portland Trail Blazers"
	NBAFranchiseSacramentoKings       NBAFranchiseName = "Sacramento Kings"
	NBAFranchiseSanAntonioSpurs       NBAFranchiseName = "San Antonio Spurs"
	NBAFranchiseTorontoRaptors        NBAFranchiseName = "Toronto Raptors"
	NBAFranchiseUtahJazz              NBAFranchiseName = "Utah Jazz"
	NBAFranchiseWashingtonWizards     NBAFranchiseName = "Washington Wizards"
)

var (
	NBAFranchises = []NBAFranchise{
		{Name: NBAFranchiseAtlantaHawks, Acronym: "ATL", City: "Atlanta", TeamName: "Hawks"},
		{Name: NBAFranchiseBostonCeltics, Acronym: "BOS", City: "Boston", TeamName: "Celtics"},
		{Name: NBAFranchiseBrooklynNets, Acronym: "BKN", City: "Brooklyn", TeamName: "Nets"},
		{Name: NBAFranchiseCharlotteHornets, Acronym: "CHA", City: "Charlotte", TeamName: "Hornets"},
		{Name: NBAFranchiseChicagoBulls, Acronym: "CHI", City: "Chicago", TeamName: "Bulls"},
		{Name: NBAFranchiseClevelandCavaliers, Acronym: "CLE", City: "Cleveland", TeamName: "Cavaliers"},
		{Name: NBAFranchiseDallasMavericks, Acronym: "DAL", City: "Dallas", TeamName: "Mavericks"},
		{Name: NBAFranchiseDenverNuggets, Acronym: "DEN", City: "Denver", TeamName: "Nuggets"},
		{Name: NBAFranchiseDetroitPistons, Acronym: "DET", City: "Detroit", TeamName: "Pistons"},
		{Name: NBAFranchiseGoldenStateWarriors, Acronym: "GSW", City: "San Francisco", TeamName: "Warriors", AcronymAlt: "GS"},
		{Name: NBAFranchiseHoustonRockets, Acronym: "HOU", City: "Houston", TeamName: "Rockets"},
		{Name: NBAFranchiseIndianaPacers, Acronym: "IND", City: "Indiana", TeamName: "Pacers"},
		{Name: NBAFranchiseLosAngelesClippers, Acronym: "LAC", City: "Los Angeles", TeamName: "Clippers"},
		{Name: NBAFranchiseLosAngelesLakers, Acronym: "LAL", City: "Los Angeles", TeamName: "Lakers"},
		{Name: NBAFranchiseMemphisGrizzlies, Acronym: "MEM", City: "Memphis", TeamName: "Grizzlies"},
		{Name: NBAFranchiseMiamiHeat, Acronym: "MIA", City: "Miami", TeamName: "Heat"},
		{Name: NBAFranchiseMilwaukeeBucks, Acronym: "MIL", City: "Milwaukee", TeamName: "Bucks"},
		{Name: NBAFranchiseMinnesotaTimberwolves, Acronym: "MIN", City: "Minnesota", TeamName: "Timberwolves"},
		{Name: NBAFranchiseNewOrleansPelicans, Acronym: "NOP", City: "New Orleans", TeamName: "Pelicans", AcronymAlt: "NO"},
		{Name: NBAFranchiseNewYorkKnicks, Acronym: "NYK", City: "New York", TeamName: "Knicks", AcronymAlt: "NY"},
		{Name: NBAFranchiseOklahomaCityThunder, Acronym: "OKC", City: "Oklahoma", TeamName: "Thunder"},
		{Name: NBAFranchiseOrlandoMagic, Acronym: "ORL", City: "Orlando", TeamName: "Magic"},
		{Name: NBAFranchisePhiladelphia76ers, Acronym: "PHI", City: "Philadelphia", TeamName: "76ers"},
		{Name: NBAFranchisePhoenixSuns, Acronym: "PHX", City: "Phoenix", TeamName: "Suns"},
		{Name: NBAFranchisePortlandTrailBlazers, Acronym: "POR", City: "Portland", TeamName: "Trail Blazers"},
		{Name: NBAFranchiseSacramentoKings, Acronym: "SAC", City: "Sacramento", TeamName: "Kings"},
		{Name: NBAFranchiseSanAntonioSpurs, Acronym: "SAS", City: "San Antonio", TeamName: "Spurs", AcronymAlt: "SA"},
		{Name: NBAFranchiseTorontoRaptors, Acronym: "TOR", City: "Toronto", TeamName: "Raptors"},
		{Name: NBAFranchiseUtahJazz, Acronym: "UTA", City: "Utah", TeamName: "Jazz", AcronymAlt: "UTAH"},
		{Name: NBAFranchiseWashingtonWizards, Acronym: "WAS", City: "Washington", TeamName: "Wizards"},
	}
)

type NBAFranchise struct {
	Name       NBAFranchiseName
	Acronym    string
	AcronymAlt string
	City       string
	TeamName   string
}

func generateMatchIdFromTitle(title string) string {
	team1, team2 := parseTeamsFromTitle(title)
	if team1 == nil || team2 == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", team1.Acronym, team2.Acronym)
}

func parseTeamsFromTitle(title string) (*NBAFranchise, *NBAFranchise) {
	var team1 *NBAFranchise
	var team2 *NBAFranchise

	// Ensure a match is the same regardless of order in title, search from NBAFranchises
	for _, franchise := range NBAFranchises {

		if team1 == nil && titleHasNBAFranchiseInfo(title, franchise) {
			team1 = &franchise
		} else if team2 == nil && titleHasNBAFranchiseInfo(title, franchise) {
			team2 = &franchise
		}

		if team1 != nil && team2 != nil {
			break
		}
	}
	return team1, team2
}

func titleHasNBAFranchiseInfo(title string, franchise NBAFranchise) bool {
	acronym := "(" + franchise.Acronym + ")"
	acronymAlt := "(" + franchise.AcronymAlt + ")"
	if strings.Contains(title, string(franchise.Name)) ||
		strings.Contains(title, franchise.TeamName) ||
		strings.Contains(title, acronym) ||
		(franchise.AcronymAlt != "" && strings.Contains(title, acronymAlt)) {
		return true
	}
	return false

}

func NewNBAFranchisesMap() map[string]NBAFranchise {
	nbaFranchisesMap := make(map[string]NBAFranchise)
	for _, franchise := range NBAFranchises {
		nbaFranchisesMap[franchise.Acronym] = franchise
	}
	return nbaFranchisesMap
}
