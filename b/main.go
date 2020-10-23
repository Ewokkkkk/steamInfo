package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
)

// GameList はストアの全ゲームをsteamAPIから取得し、格納する構造体
type GameList struct {
	Applist struct {
		Apps []struct {
			Appid  int    `json:"appid"`
			Name   string `json:"name"`
			IsFree bool
			Price  int
			Date   string
		} `json:"apps"`
	} `json:"applist"`
}

// type GameInfo struct {
// 	isFree bool
// 	Price  int
// 	Date   string
// }

// var GamesInfo []GameInfo

func main() {
	games := getGameList()
	getGamesInfo(games)
	// fmt.Println(games)

	db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/steam-info-db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	stmtInsert, err := db.Prepare("INSERT INTO games_info(id, name, release_date, price, is_free) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer stmtInsert.Close()
	for _, val := range games.Applist.Apps {
		id := val.Appid
		name := val.Name
		releaseDate := val.Date
		price := val.Price
		isFree := val.IsFree

		result, err := stmtInsert.Exec(id, name, releaseDate, price, isFree)
		if err != nil {
			panic(err.Error())
		}
		lastInsertID, err := result.LastInsertId()
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(lastInsertID)

	}
}

func getGameList() GameList {
	url := "https://api.steampowered.com/ISteamApps/GetAppList/v2/"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(request.URL.String())
	timeout := time.Duration(10 * time.Second)
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
	var gameList GameList
	if err := json.Unmarshal(body, &gameList); err != nil {
		log.Fatal(err)
	}

	return gameList
}

func getGamesInfo(games GameList) {
	for i, val := range games.Applist.Apps {
		url := "https://store.steampowered.com/api/appdetails"
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		params := request.URL.Query()
		// params.Add("filters", "price_overview,release_date")
		params.Add("cc", "jpy")
		params.Add("l", "japanese")
		params.Add("appids", strconv.Itoa(val.Appid))
		request.URL.RawQuery = params.Encode()
		fmt.Println(i, request.URL.String())

		timeout := time.Duration(10 * time.Second)
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
		for k := range m {
			keys = append(keys, k)
		}

		isFree, err := js.GetPath(keys[0], "data", "is_free").Bool()
		if err != nil {
			isFree = false
		}
		price, err := js.GetPath(keys[0], "data", "price_overview", "initial").Int()
		if err != nil {
			price = -1
		}
		date, err := js.GetPath(keys[0], "data", "release_date", "date").String()
		if err != nil {
			date = "-"
		}

		if isFree == true {
			games.Applist.Apps[i].Price = 0
		} else if price != -1 {
			price /= 100
			games.Applist.Apps[i].Price = price
		}
		games.Applist.Apps[i].IsFree = isFree
		games.Applist.Apps[i].Date = date

		time.Sleep(time.Second * 2)

		// テスト用
		if i > 1000 {
			return
		}

	}
}
