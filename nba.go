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
	NBAFranchises = map[NBAFranchiseName]NBAFranchise{
		NBAFranchiseAtlantaHawks:          {Acronym: "ATL", AcronymAlt: "", City: "Atlanta", TeamName: "Hawks"},
		NBAFranchiseBostonCeltics:         {Acronym: "BOS", AcronymAlt: "", City: "Boston", TeamName: "Celtics"},
		NBAFranchiseBrooklynNets:          {Acronym: "BKN", AcronymAlt: "", City: "Brooklyn", TeamName: "Nets"},
		NBAFranchiseCharlotteHornets:      {Acronym: "CHA", AcronymAlt: "", City: "Charlotte", TeamName: "Hornets"},
		NBAFranchiseChicagoBulls:          {Acronym: "CHI", AcronymAlt: "", City: "Chicago", TeamName: "Bulls"},
		NBAFranchiseClevelandCavaliers:    {Acronym: "CLE", AcronymAlt: "", City: "Cleveland", TeamName: "Cavaliers"},
		NBAFranchiseDallasMavericks:       {Acronym: "DAL", AcronymAlt: "", City: "Dallas", TeamName: "Mavericks"},
		NBAFranchiseDenverNuggets:         {Acronym: "DEN", AcronymAlt: "", City: "Denver", TeamName: "Nuggets"},
		NBAFranchiseDetroitPistons:        {Acronym: "DET", AcronymAlt: "", City: "Detroit", TeamName: "Pistons"},
		NBAFranchiseGoldenStateWarriors:   {Acronym: "GSW", AcronymAlt: "GS", City: "San Francisco", TeamName: "Golden State Warriors"},
		NBAFranchiseHoustonRockets:        {Acronym: "HOU", AcronymAlt: "", City: "Houston", TeamName: "Rockets"},
		NBAFranchiseIndianaPacers:         {Acronym: "IND", AcronymAlt: "", City: "Indiana", TeamName: "Pacers"},
		NBAFranchiseLosAngelesClippers:    {Acronym: "LAC", AcronymAlt: "", City: "Los Angeles", TeamName: "Clippers"},
		NBAFranchiseLosAngelesLakers:      {Acronym: "LAL", AcronymAlt: "", City: "Los Angeles", TeamName: "Lakers"},
		NBAFranchiseMemphisGrizzlies:      {Acronym: "MEM", AcronymAlt: "", City: "Memphis", TeamName: "Grizzlies"},
		NBAFranchiseMiamiHeat:             {Acronym: "MIA", AcronymAlt: "", City: "Miami", TeamName: "Heat"},
		NBAFranchiseMilwaukeeBucks:        {Acronym: "MIL", AcronymAlt: "", City: "Milwaukee", TeamName: "Bucks"},
		NBAFranchiseMinnesotaTimberwolves: {Acronym: "MIN", AcronymAlt: "", City: "Minnesota", TeamName: "Timberwolves"},
		NBAFranchiseNewOrleansPelicans:    {Acronym: "NOP", AcronymAlt: "NO", City: "New Orleans", TeamName: "Pelicans"},
		NBAFranchiseNewYorkKnicks:         {Acronym: "NYK", AcronymAlt: "NY", City: "New York", TeamName: "York Knicks"},
		NBAFranchiseOklahomaCityThunder:   {Acronym: "OKC", AcronymAlt: "", City: "Oklahoma", TeamName: "Thunder"},
		NBAFranchiseOrlandoMagic:          {Acronym: "ORL", AcronymAlt: "", City: "Orlando", TeamName: "Magic"},
		NBAFranchisePhiladelphia76ers:     {Acronym: "PHI", AcronymAlt: "", City: "Philadelphia", TeamName: "76ers"},
		NBAFranchisePhoenixSuns:           {Acronym: "PHX", AcronymAlt: "", City: "Phoenix", TeamName: "Suns"},
		NBAFranchisePortlandTrailBlazers:  {Acronym: "POR", AcronymAlt: "", City: "Portland", TeamName: "Trail Blazers"},
		NBAFranchiseSacramentoKings:       {Acronym: "SAC", AcronymAlt: "", City: "Sacramento", TeamName: "Kings"},
		NBAFranchiseSanAntonioSpurs:       {Acronym: "SAS", AcronymAlt: "SA", City: "San Antonio", TeamName: "Spurs"},
		NBAFranchiseTorontoRaptors:        {Acronym: "TOR", AcronymAlt: "", City: "Toronto", TeamName: "Raptors"},
		NBAFranchiseUtahJazz:              {Acronym: "UTA", AcronymAlt: "UTAH", City: "Utah", TeamName: "Jazz"},
		NBAFranchiseWashingtonWizards:     {Acronym: "WAS", AcronymAlt: "", City: "Washington", TeamName: "Wizards"},
	}
)

type NBAFranchise struct {
	Acronym    string
	AcronymAlt string
	City       string
	TeamName   string
}
type NBAFranchiseSlice struct {
	Name      NBAFranchiseName
	Franchise NBAFranchise
}

func generateMatchIdFromTitle(title string, sortedNbaFranchiseSlice []NBAFranchiseSlice) string {
	team1, team2 := parseTeamsFromTitle(title, sortedNbaFranchiseSlice)
	if team1 == nil || team2 == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", team1.Acronym, team2.Acronym)
}

func parseTeamsFromTitle(title string, sortedNbaFranchiseSlice []NBAFranchiseSlice) (*NBAFranchise, *NBAFranchise) {
	var team1 *NBAFranchise
	var team2 *NBAFranchise

	// Ensure a match is the same regardless of order in title, search from NBAFranchises
	for _, franchise := range sortedNbaFranchiseSlice {

		if team1 == nil && titleIsNBAFranchise(title, franchise) {
			team1 = &franchise.Franchise
		} else if team2 == nil && titleIsNBAFranchise(title, franchise) {
			team2 = &franchise.Franchise
		}

		if team1 != nil && team2 != nil {
			break
		}
	}
	return team1, team2
}

func titleIsNBAFranchise(title string, franchise NBAFranchiseSlice) bool {
	acronym := "(" + franchise.Franchise.Acronym + ")"
	if strings.Contains(title, string(franchise.Name)) || strings.Contains(title, acronym) || strings.Contains(title, franchise.Franchise.TeamName) {
		return true
	}
	return false

}
