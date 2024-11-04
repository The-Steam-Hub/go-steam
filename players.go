package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Player struct {
	SteamID                    string `json:"steamid"`
	Name                       string `json:"personaname"`
	TimeCreated                int64  `json:"timecreated"`
	CountryCode                string `json:"loccountrycode"`
	StateCode                  string `json:"locstatecode"`
	AvatarFull                 string `json:"avatarfull"`
	RealName                   string `json:"realname"`
	CommunityBanned            bool   `json:"CommunityBanned"`
	VACBanned                  bool   `json:"VACBanned"`
	NumOfVacBans               int    `json:"NumberOfVACBans"`
	DaysSinceLastBan           int    `json:"DaysSinceLastBan"`
	NumOfGameBans              int    `json:"NumberOfGameBans"`
	EconomyBan                 string `json:"EconomyBan"`
	ProfileURL                 string `json:"profileurl"`
	LastLogOff                 int    `json:"lastlogoff"`
	PlayerXP                   int
	PlayerLevel                int
	PlayerLevelPercentile      float64
	PlayerXPNeededToLevelUp    int
	PlayerXPNeededCurrentLevel int
	PersonaState               int
}

var (
	ErrPlayerNotFound = errors.New("player not found")
)

func (s Steam) GetPlayerSummaries(steamID ...string) (*[]Player, error) {
	baseURL, _ := url.Parse(SteamWebAPIISteamUser)
	baseURL.Path += "GetPlayerSummaries/v0002"

	params := url.Values{}
	params.Add("key", s.Key)
	params.Add("steamids", strings.Join(steamID, ","))
	params.Add("format", "json")
	params.Add("cc", s.CountryCode)
	params.Add("l", s.Langauge)
	baseURL.RawQuery = params.Encode()

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return nil, err
	}

	if err := HandleStatus(resp); err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		Players struct {
			Players []Player `json:"players"`
		} `json:"response"`
	}

	json.Unmarshal(b, &response)

	// Steam will still return a 200 if the user is not found
	// so we need to check if the response is empty
	if len(response.Players.Players) == 0 {
		return nil, ErrPlayerNotFound
	}

	return &response.Players.Players, nil
}

func (s Steam) GetPlayerBans(p *Player) error {
	baseURL, _ := url.Parse(SteamWebAPIISteamUser)
	baseURL.Path += "GetPlayerBans/v1"

	params := url.Values{}
	params.Add("key", s.Key)
	params.Add("steamids", p.SteamID)
	params.Add("format", "json")
	params.Add("cc", s.CountryCode)
	params.Add("l", s.Langauge)
	baseURL.RawQuery = params.Encode()

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return err
	}

	if err := HandleStatus(resp); err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response struct {
		Players []Player `json:"players"`
	}

	json.Unmarshal(b, &response)
	p.CommunityBanned = response.Players[0].CommunityBanned
	p.VACBanned = response.Players[0].VACBanned
	p.NumOfVacBans = response.Players[0].NumOfVacBans
	p.DaysSinceLastBan = response.Players[0].DaysSinceLastBan
	p.NumOfGameBans = response.Players[0].NumOfGameBans
	p.EconomyBan = response.Players[0].EconomyBan
	return nil
}

func (s Steam) GetPlayerBadges(p *Player) error {
	baseURL, _ := url.Parse(SteamWebAPIIPlayerService)
	baseURL.Path += "GetBadges/v1"

	params := url.Values{}
	params.Add("key", s.Key)
	params.Add("steamid", p.SteamID)
	params.Add("format", "json")
	params.Add("cc", s.CountryCode)
	params.Add("l", s.Langauge)
	baseURL.RawQuery = params.Encode()

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return err
	}

	if err := HandleStatus(resp); err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response struct {
		Player struct {
			PlayerXP                   int `json:"player_xp"`
			PlayerLevel                int `json:"player_level"`
			PlayerXPNeededToLevelUp    int `json:"player_xp_needed_to_level_up"`
			PlayerXPNeededCurrentLevel int `json:"player_xp_needed_current_level"`
		} `json:"response"`
	}

	json.Unmarshal(b, &response)
	p.PlayerXP = response.Player.PlayerXP
	p.PlayerLevel = response.Player.PlayerLevel
	p.PlayerXPNeededToLevelUp = response.Player.PlayerXPNeededToLevelUp
	p.PlayerXPNeededCurrentLevel = response.Player.PlayerXPNeededCurrentLevel
	return nil
}

func (s Steam) GetPlayerLevelDistribution(p *Player) error {
	baseURL, _ := url.Parse(SteamWebAPIIPlayerService)
	baseURL.Path += "GetSteamLevelDistribution/v1"

	params := url.Values{}
	params.Add("key", s.Key)
	params.Add("player_level", strconv.Itoa(p.PlayerLevel))
	params.Add("format", "json")
	params.Add("cc", s.CountryCode)
	params.Add("l", s.Langauge)
	baseURL.RawQuery = params.Encode()

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return err
	}

	if err := HandleStatus(resp); err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response struct {
		Player struct {
			PlayerLevelPercentile float64 `json:"player_level_percentile"`
		} `json:"response"`
	}

	json.Unmarshal(b, &response)
	p.PlayerLevelPercentile = response.Player.PlayerLevelPercentile
	return nil
}

func (p Player) Status() string {
	var statusEmoji string
	switch p.PersonaState {
	case 0:
		statusEmoji = "⚫" // Offline
	case 1:
		statusEmoji = "🟢" // Online
	case 2:
		statusEmoji = "🔴" // Busy
	case 3:
		statusEmoji = "🟡" // Away
	case 4:
		statusEmoji = "💤" // Snooze
	case 5:
		statusEmoji = "🔄" // Looking to trade
	case 6:
		statusEmoji = "🎮" // Looking to play
	}
	return statusEmoji
}
