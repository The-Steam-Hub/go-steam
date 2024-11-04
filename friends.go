package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type Friend struct {
	ID           string `json:"steamid"`
	Relationship string `json:"relationship"`
	FriendsSince int64  `json:"friend_since"`
}

func (s Steam) GetFriendsList(steamID string) (*[]Friend, error) {
	baseURL, _ := url.Parse(SteamWebAPIISteamUser)
	baseURL.Path += "GetFriendList/v0001"

	params := url.Values{}
	params.Add("key", s.Key)
	params.Add("steamid", steamID)
	params.Add("relationship", "friend")
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
		FriendsList struct {
			Friends []Friend `json:"friends"`
		} `json:"friendslist"`
	}

	json.Unmarshal(b, &response)
	return &response.FriendsList.Friends, nil
}
