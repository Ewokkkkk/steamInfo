package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

type Data struct {
	Personaname   string
	Profileurl    string
	Avatarfull    string
	GameCount     int
	TotalPlaytime int
	Games         []struct {
		// Appid                    int
		// Name                     string
		// PlaytimeForever          int
		// ImgIconURL               string
		// ImgLogoURL               string
		// PlaytimeWindowsForever   int
		// PlaytimeMacForever       int
		// PlaytimeLinuxForever     int
		// HasCommunityVisibleStats bool
		// Playtime2Weeks           int
		Appid                    int    `json:"appid"`
		Name                     string `json:"name"`
		PlaytimeForever          int    `json:"playtime_forever"`
		ImgIconURL               string `json:"img_icon_url"`
		ImgLogoURL               string `json:"img_logo_url"`
		PlaytimeWindowsForever   int    `json:"playtime_windows_forever"`
		PlaytimeMacForever       int    `json:"playtime_mac_forever"`
		PlaytimeLinuxForever     int    `json:"playtime_linux_forever"`
		HasCommunityVisibleStats bool   `json:"has_community_visible_stats,omitempty"`
		Playtime2Weeks           int    `json:"playtime_2weeks,omitempty"`
		Price                    string
	}
}

type PlayerSummaries struct {
	Response struct {
		Players []struct {
			Steamid     string `json:"steamid"`
			Personaname string `json:"personaname"`
			Profileurl  string `json:"profileurl"`
			Avatarfull  string `json:"avatarfull"`
		} `json:"players"`
	} `json:"response"`
}

type OwnedGames struct {
	Response struct {
		GameCount int `json:"game_count"`
		Games     []struct {
			Appid                    int    `json:"appid"`
			Name                     string `json:"name"`
			PlaytimeForever          int    `json:"playtime_forever"`
			ImgIconURL               string `json:"img_icon_url"`
			ImgLogoURL               string `json:"img_logo_url"`
			PlaytimeWindowsForever   int    `json:"playtime_windows_forever"`
			PlaytimeMacForever       int    `json:"playtime_mac_forever"`
			PlaytimeLinuxForever     int    `json:"playtime_linux_forever"`
			HasCommunityVisibleStats bool   `json:"has_community_visible_stats,omitempty"`
			Playtime2Weeks           int    `json:"playtime_2weeks,omitempty"`
			Price                    string
		} `json:"games"`
	} `json:"response"`
}

func GetPlayerSummaries(apiKey, steamid string) PlayerSummaries {
	url := "http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	params := request.URL.Query()
	params.Add("key", apiKey)
	params.Add("steamids", steamid)
	request.URL.RawQuery = params.Encode()

	fmt.Println(request.URL.String())

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var player PlayerSummaries
	if err := json.Unmarshal(body, &player); err != nil {
		log.Fatal(err)
	}

	return player
}

func getUserGameList(apiKey, steamid string) OwnedGames {
	url := "http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	params := request.URL.Query()
	params.Add("key", apiKey)
	params.Add("steamid", steamid)
	params.Add("include_appinfo", "true")
	params.Add("format", "json")
	request.URL.RawQuery = params.Encode()
	fmt.Println(request.URL.String())

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var games OwnedGames
	if err := json.Unmarshal(body, &games); err != nil {
		log.Fatal(err)
	}

	return games
}

func getPriceJPY(Appids []string) string {
	url := "https://store.steampowered.com/api/appdetails"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	id := strings.Join(Appids, ",")
	// id = id + id + id + id
	params := request.URL.Query()
	params.Add("appids", id)
	params.Add("filters", "price_overview")
	request.URL.RawQuery = params.Encode()
	fmt.Println(request.URL.String())
	timeout := time.Duration(5 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	js, err := simplejson.NewJson(body)
	if err != nil {
		log.Fatal(err)
	}
	m, err := js.Map()
	if err != nil {
		log.Fatal(err)
	}
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}

	price, err := js.GetPath(keys[0], "data", "price_overview", "initial_formatted").String()
	if err != nil {
		return "--"
	}
	// セール中でなければinitial_formattedは空なのでfinal_formattedから取得する
	if price == "" {
		price, err = js.GetPath(keys[0], "data", "price_overview", "final_formatted").String()
		if err != nil {
			return "--"
		}
	}
	return price

}

// テンプレート用独自関数
func HourTimes(i int) int {
	return i / 60
}

func user(w http.ResponseWriter, r *http.Request) {
	var appids []string

	fmt.Println("method:", r.Method)
	id := r.FormValue("userid")
	playerSummaries := GetPlayerSummaries(apiKey, id)
	ownedGames := getUserGameList(apiKey, id)
	TotalPlaytime := 0
	gameCount := ownedGames.Response.GameCount

	fmt.Println(playerSummaries.Response.Players[0].Personaname)
	fmt.Println("game count:", gameCount)
	for _, val := range ownedGames.Response.Games {
		fmt.Printf("%-40v% 5vh\n", val.Name, val.PlaytimeForever/60)
		TotalPlaytime += val.PlaytimeForever
		// val.Price = getPriceJPY(val.Appid)
		// time.Sleep(time.Second * 5)
		appids = append(appids, strconv.Itoa(val.Appid))
	}
	// getPriceJPY(appids) // test

	fmt.Printf("total:% 40vh\n", TotalPlaytime/60)

	funcs := template.FuncMap{
		"HourTimes": HourTimes,
	}

	t, err := template.New("user.html").Funcs(funcs).ParseFiles("templates/user.html")
	if err != nil {
		log.Fatal(err)
	}
	data := Data{
		Personaname:   playerSummaries.Response.Players[0].Personaname,
		Profileurl:    playerSummaries.Response.Players[0].Profileurl,
		Avatarfull:    playerSummaries.Response.Players[0].Avatarfull,
		GameCount:     ownedGames.Response.GameCount,
		TotalPlaytime: TotalPlaytime / 60,
		Games:         ownedGames.Response.Games,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}