package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// GameList はストアの全ゲームをsteamAPIから取得し、格納する構造体
type GameList struct {
	Applist struct {
		Apps []struct {
			Appid int    `json:"appid"`
			Name  string `json:"name"`
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
	// fmt.Println(games)

	db, err := sql.Open("mysql", "ew:4253@tcp(192.168.0.6:8889)/steam-info-db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	stmtInsert, err := db.Prepare("INSERT INTO game_list(app_id, app_name) VALUES(?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer stmtInsert.Close()
	for _, val := range games.Applist.Apps {
		id := val.Appid
		name := val.Name

		result, err := stmtInsert.Exec(id, name)
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
