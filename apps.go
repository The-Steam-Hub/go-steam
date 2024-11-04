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

type App struct {
	AppID string `json:"appid"`
	Name  string `json:"name"`
}

type AppNews struct {
	AppID     string `json:"appid"`
	GID       string `json:"gid"`
	Title     string `json:"title"`
	URL       string `json:"URL"`
	Author    string `json:"author"`
	Contents  string `json:"contents"`
	FeedLabel string `json:"feedlabel"`
	Date      int    `json:"date"`
	FeedName  string `json:"feedname"`
	FeedType  int    `json:"feed_type"`
}

type AppGlobalAchievements struct {
	Name    string  `json:"name"`
	Percent float32 `json:"percent"`
}

type AppPlayTime struct {
	AppID                  int    `json:"appid"`
	Name                   string `json:"name"`
	PlayTimeForever        int    `json:"playtime_forever"`
	PlayTimeWindowsForever int    `json:"playtime_windows_forever"`
	PlayTimeMacForever     int    `json:"playtime_mac_forever"`
	PlayTimeLinuxForever   int    `json:"playtime_linux_forever"`
	PlayTimeDeckForever    int    `json:"playtime_deck_forever"`
	PlayTime2Weeks         int    `json:"playtime_2weeks"`
}

type AppDetails struct {
	Name             string   `json:"name"`
	AppID            int      `json:"steam_appid"`
	ShortDescription string   `json:"short_description"`
	Developers       []string `json:"developers"`
	Publishers       []string `json:"publishers"`
	HeaderImage      string   `json:"header_image"`
	IsFree           bool     `json:"is_free"`
	DLC              []string `json:"dlc"`
	PriceOverview    struct {
		FinalFormatted   string `json:"final_formatted"`
		InitialFormatted string `json:"initial_formatted"`
		DiscountPercent  int    `json:"discount_percent"`
	} `json:"price_overview"`
	ReleaseDate struct {
		ComingSoon bool   `json:"coming_soon"`
		Date       string `json:"date"`
	} `json:"release_date"`
	Genres []struct {
		ID          string `json:"id"`
		Description string `json:"description"`
	} `json:"genres"`
}

var (
	ErrAppNotFound  = errors.New("app not found")
	ErrNewsNotFound = errors.New("app news not found")
)

func (s Steam) GetAppList() (*[]App, error) {
	baseURL, _ := url.Parse(SteamWebAPIISteamApps)
	baseURL.Path += "GetAppList/v2/"

	params := url.Values{}
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
		AppList struct {
			Apps []App `json:"apps"`
		} `json:"applist"`
	}

	json.Unmarshal(b, &response)
	return &response.AppList.Apps, nil
}

func (s Steam) GetAppsOwnedByPlayer(steamID string) (*[]AppPlayTime, error) {
	baseURL, _ := url.Parse(SteamWebAPIIPlayerService)
	baseURL.Path += "GetOwnedGames/v0001"

	params := url.Values{}
	params.Add("key", s.Key)
	params.Add("steamid", steamID)
	params.Add("include_appinfo", "true")
	params.Add("include_free_games", "true")
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
		Games struct {
			PlayTimeStatistics []AppPlayTime `json:"games"`
		} `json:"response"`
	}

	json.Unmarshal(b, &response)
	return &response.Games.PlayTimeStatistics, nil
}

func (s Steam) GetAppNews(appID int) (*AppNews, error) {
	baseURL, _ := url.Parse(SteamWebAPIISteamNews)
	baseURL.Path += "GetNewsForApp/v2"

	params := url.Values{}
	params.Add("appid", strconv.Itoa(appID))
	params.Add("count", "1")
	params.Add("feeds", "steam_community_announcements")
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
		AppNews struct {
			NewsItems []AppNews `json:"newsitems"`
		} `json:"appnews"`
	}

	json.Unmarshal(b, &response)
	if len(response.AppNews.NewsItems) == 0 {
		return nil, ErrNewsNotFound
	}

	return &response.AppNews.NewsItems[0], nil
}

func (s Steam) AppSearch(appName string) (int, error) {
	baseURL, _ := url.Parse(SteamPoweredAPI)
	baseURL.Path += "api/storesearch"

	params := url.Values{}
	params.Add("term", appName)
	params.Add("format", "json")
	params.Add("cc", s.CountryCode)
	params.Add("l", s.Langauge)
	baseURL.RawQuery = params.Encode()

	resp, err := http.Get(baseURL.String())
	if err != nil {
		return -1, err
	}

	if err := HandleStatus(resp); err != nil {
		return -1, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	var response struct {
		Items []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}

	json.Unmarshal(b, &response)
	if len(response.Items) == 0 {
		return -1, ErrAppNotFound
	}

	// Steam fails to return the correct game if there are multiple games that share similar names
	// For example: "Frostpunk" and "Frostpunk 2" both start with "Frostpunk". Searching for "Frostpunk" can
	// result in "Frostpunk 2" being returned as the first index
	for _, v := range response.Items {
		if strings.EqualFold(v.Name, appName) {
			return v.ID, nil
		}
	}

	return response.Items[0].ID, nil
}

func (s Steam) GetAppGlobalAchievements(appID int) (*[]AppGlobalAchievements, error) {
	baseURL, _ := url.Parse(SteamWebAPIISteamUserStats)
	baseURL.Path += "GetGlobalAchievementPercentagesForApp/v0002"

	params := url.Values{}
	params.Add("gameid", strconv.Itoa(appID))
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
		AchievementPercentages struct {
			AppGlobalAchievements []AppGlobalAchievements `json:"achievements"`
		} `json:"achievementpercentages"`
	}

	json.Unmarshal(b, &response)
	return &response.AchievementPercentages.AppGlobalAchievements, nil
}

func (s Steam) GetAppDetails(appID int) (*AppDetails, error) {
	baseURL, _ := url.Parse(SteamPoweredAPI)
	baseURL.Path += "api/appdetails"

	params := url.Values{}
	params.Add("appids", strconv.Itoa(appID))
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

	var response map[string]struct {
		AppData AppDetails `json:"data"`
	}

	json.Unmarshal(b, &response)
	appData := response[strconv.Itoa(appID)].AppData
	return &appData, nil
}

func (s Steam) GetAppsRecentlyPlayedByPlayer(steamID string) (*[]AppPlayTime, error) {
	baseURL, _ := url.Parse(SteamWebAPIIPlayerService)
	baseURL.Path += "GetRecentlyPlayedGames/v0001"

	params := url.Values{}
	params.Add("key", s.Key)
	params.Add("steamid", steamID)
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
		Games struct {
			PlayTimeStatistics []AppPlayTime `json:"games"`
		} `json:"response"`
	}

	json.Unmarshal(b, &response)
	return &response.Games.PlayTimeStatistics, nil
}
