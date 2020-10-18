package main

import (
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

type GameList struct {
	Applist struct {
		Apps []struct {
			Appid int    `json:"appid"`
			Name  string `json:"name"`
			Price int
			Date  string
		} `json:"apps"`
	} `json:"applist"`
}

type GameInfo struct {
	Price int
	Date  string
}

var GamesInfo []GameInfo

func main() {
	games := getGameList()
	// for _, val := range games.Applist.Apps {
	// 	getGamesInfo(val.Appid)
	// 	fmt.Println(GamesInfo)
	// 	time.Sleep(time.Second * 3)
	// }
	getGamesInfo(games)
	fmt.Println(games.Applist.Apps[0].Date)

	// db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/steam-info-db")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer db.Close()

	// stmtInsert, err := db.Prepare("INSERT INTO games_info(id, name) VALUES(?, ?)")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer stmtInsert.Close()
	// for _, val := range games.Applist.Apps {
	// 	id := val.Appid
	// 	name := val.Name
	// 	result, err := stmtInsert.Exec(id, name)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	lastInsertID, err := result.LastInsertId()
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	fmt.Println(lastInsertID)

	// }
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

// func getGamesInfo(Appid int) {
func getGamesInfo(games GameList) {
	cnt := 0
	for _, val := range games.Applist.Apps {

		url := "https://store.steampowered.com/api/appdetails"
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		params := request.URL.Query()
		params.Add("filters", "price_overview,release_date")
		params.Add("cc", "jpy")
		params.Add("l", "japanese")
		params.Add("appids", strconv.Itoa(val.Appid))
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
		price, err := js.GetPath(keys[0], "data", "price_overview", "initial").Int()
		if err != nil {
			price = -1
		}

		date, err := js.GetPath(keys[0], "data", "release_date", "date").String()
		if err != nil {
			date = "-"
		}
		val.Price = price
		val.Date = date

		time.Sleep(time.Second * 3)

	}
}
