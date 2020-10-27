package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func main() {
	games := getGameList()

	db, err := sql.Open("mysql", "admin:"+os.Getenv("RDS_PASS")+"@tcp(database-1.cop2pvzm3623.ap-northeast-1.rds.amazonaws.com)/steam-info-db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	fmt.Println("DB connected.")

	stmtInsert, err := db.Prepare("INSERT INTO game_list(app_id, app_name) VALUES(?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer stmtInsert.Close()
	for _, val := range games.Applist.Apps {
		id := val.Appid
		name := val.Name

		var (
			rowID   int
			appID   int
			appName string
			date    interface{}
		)

		err = db.QueryRow("SELECT * FROM game_list WHERE app_id = ?", id).Scan(&rowID, &appID, &appName, &date)
		switch {
		case err == sql.ErrNoRows:
			result, err := stmtInsert.Exec(id, name)
			if err != nil {
				panic(err.Error())
			}
			lastInsertID, err := result.LastInsertId()
			if err != nil {
				panic(err.Error())
			}
			fmt.Println(lastInsertID)
		case err != nil:
			panic(err.Error())
		default:
			// fmt.Println("already exists.")
		}

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
