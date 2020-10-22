package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Data struct {
	Personaname string
	Profileurl  string
	// Avatarfull    string
	Avatarmedium  string
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
		Price                    int
		ReleaseDate              string
	}
}

type PlayerSummaries struct {
	Response struct {
		Players []struct {
			Steamid     string `json:"steamid"`
			Personaname string `json:"personaname"`
			Profileurl  string `json:"profileurl"`
			// Avatarfull   string `json:"avatarfull"`
			Avatarmedium string `json:"avatarmedium"`
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
			Price                    int
			ReleaseDate              string
		} `json:"games"`
	} `json:"response"`
}
type VanityURL struct {
	Response struct {
		Success int    `json:"success"`
		Steamid string `json:"steamid"`
		Message string `json:"message"`
	} `json:"response"`
}

func getPlayerSummaries(apiKey, steamid string) PlayerSummaries {
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

func getGamesInfo(appid int) (price int, releaseDate string) {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/steam-info-db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	var d string
	var p int
	err = db.QueryRow("SELECT release_date, price FROM games_info WHERE id = ? LIMIT 1", appid).Scan(&d, &p)
	if err == sql.ErrNoRows { //  見つからなかった
		d = "-"
		p = -1
	} else if err != nil { // それ以外のエラー
		log.Fatalln(err)
	}

	fmt.Println(d, p)
	return p, d
}

// url末尾の値(steamid or customURL)を入力し、apiからsteamidを取得する
func getSteamID(apiKey, val string) string {
	url := "http://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	params := request.URL.Query()
	params.Add("key", apiKey)
	params.Add("vanityurl", val)
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
	var responseID VanityURL
	if err := json.Unmarshal(body, &responseID); err != nil {
		log.Fatal(err)
	}
	if responseID.Response.Success != 1 {
		return val
	}

	return responseID.Response.Steamid
}

// テンプレート用独自関数
func HourTimes(i int) int {
	return i / 60
}

func user(w http.ResponseWriter, r *http.Request) {
	var appids []string

	fmt.Println("method:", r.Method)

	// url→idの処理
	//またはカスタムurl→ResolveVanityURL→id

	segs := strings.Split(r.FormValue("userid"), "/")
	val := segs[4]
	id := getSteamID(apiKey, val)

	playerSummaries := getPlayerSummaries(apiKey, id)
	ownedGames := getUserGameList(apiKey, id)
	TotalPlaytime := 0
	gameCount := ownedGames.Response.GameCount

	fmt.Println(playerSummaries.Response.Players[0].Personaname)
	fmt.Println("game count:", gameCount)
	for i, val := range ownedGames.Response.Games {
		fmt.Printf("%-40v% 5vh\n", val.Name, val.PlaytimeForever/60)
		TotalPlaytime += val.PlaytimeForever
		price, releaseDate := getGamesInfo(val.Appid)
		ownedGames.Response.Games[i].Price = price
		ownedGames.Response.Games[i].ReleaseDate = releaseDate
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
		Personaname: playerSummaries.Response.Players[0].Personaname,
		Profileurl:  playerSummaries.Response.Players[0].Profileurl,
		// Avatarfull:    playerSummaries.Response.Players[0].Avatarfull,
		Avatarmedium:  playerSummaries.Response.Players[0].Avatarmedium,
		GameCount:     ownedGames.Response.GameCount,
		TotalPlaytime: TotalPlaytime / 60,
		Games:         ownedGames.Response.Games,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}
