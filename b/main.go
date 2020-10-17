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

type GameList struct {
	Applist struct {
		Apps []struct {
			Appid int    `json:"appid"`
			Name  string `json:"name"`
		} `json:"apps"`
	} `json:"applist"`
}

type GameInfo struct {
	Success bool `json:"success"`
	Data    struct {
		PriceOverview struct {
			// Currency         string `json:"currency"`
			Initial int `json:"initial"`
			// Final            int    `json:"final"`
			// DiscountPercent  int    `json:"discount_percent"`
			// InitialFormatted string `json:"initial_formatted"`
			// FinalFormatted   string `json:"final_formatted"`
		} `json:"price_overview"`
		ReleaseDate struct {
			// ComingSoon bool   `json:"coming_soon"`
			Date string `json:"date"`
		} `json:"release_date"`
	} `json:"data"`
}

func main() {
	games := getGameList()

	db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/steam-info-db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	stmtInsert, err := db.Prepare("INSERT INTO games_info(id, name) VALUES(?, ?)")
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
