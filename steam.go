package main

import (
	"errors"
	"fmt"
	"net/http"
)

type Steam struct {
	Key         string
	Langauge    string
	CountryCode string
}

const (
	SteamWebAPI                = "http://api.steampowered.com/"
	SteamPoweredAPI            = "https://store.steampowered.com/"
	SteamCommunityAPI          = "https://steamcommunity.com/"
	SteamWebAPIIPlayerService  = SteamWebAPI + "IPlayerService/"
	SteamWebAPIISteamUser      = SteamWebAPI + "ISteamUser/"
	SteamWebAPIISteamUserStats = SteamWebAPI + "ISteamUserStats/"
	SteamWebAPIISteamApps      = SteamWebAPI + "ISteamApps/"
	SteamWebAPIISteamNews      = SteamWebAPI + "ISteamNews/"
)

var (
	ErrTooManyRequests = errors.New("rate limit exceeded: too many requests from your IP address")
	ErrForbidden       = errors.New("access denied: missing API key or IP address temporarily banned")
)

// HandleStatus will handle common HTTP status's returned by Steams API
func HandleStatus(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	default:
		return fmt.Errorf("request failed with statuscode %d", resp.StatusCode)
	}
}
