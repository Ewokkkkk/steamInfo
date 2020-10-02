package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/stretchr/objx"
)

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
		} `json:"games"`
	} `json:"response"`
}

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
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

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ =
			template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

func main() {
	// apiKey := os.Getenv("STEAM_APIKEY")
	apiKey := "B4E99A78E917F2BE6E533AB59ACEF66F"
	steamid := "76561198051101724" // my id
	playerSummaries := GetPlayerSummaries(apiKey, steamid)
	ownedGames := getUserGameList(apiKey, steamid)
	playTime := 0
	gameCount := ownedGames.Response.GameCount

	fmt.Println(playerSummaries.Response.Players[0].Personaname)
	fmt.Println("game count:", gameCount)
	for _, val := range ownedGames.Response.Games {
		fmt.Printf("%-40v% 5vh\n", val.Name, val.PlaytimeForever/60)
	}
	for _, val := range ownedGames.Response.Games {
		playTime += val.PlaytimeForever
	}
	fmt.Printf("total:% 40vh\n", playTime/60)

	http.Handle("/", &templateHandler{filename: "index.html"})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
