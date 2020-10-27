package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
)

// GameList はストアの全ゲームをsteamAPIから取得し、格納する構造体
type GameList struct {
	Appid  int
	Name   string
	IsFree bool
	Price  int
	Date   string
}

func main() {
	games := selectGameList()
	getGamesInfo(games)

}

func selectGameList() []GameList {
	db, err := sql.Open("mysql", "admin:"+os.Getenv("RDS_PASS")+"@tcp(database-1.cop2pvzm3623.ap-northeast-1.rds.amazonaws.com)/steam-info-db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	rows, err := db.Query("select app_id,app_name from game_list order by updated_at")
	if err != nil {
		log.Fatal(err)
	}
	var games []GameList
	for rows.Next() {
		game := GameList{}
		if err := rows.Scan(&game.Appid, &game.Name); err != nil {
			log.Fatal(err)
		}
		games = append(games, game)
	}
	return games
}

func getGamesInfo(games []GameList) {
	url := "https://store.steampowered.com/api/appdetails"
	for i, val := range games {
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
			games[i].Price = 0
		} else if price != -1 {
			price /= 100
			games[i].Price = price
		}
		games[i].IsFree = isFree
		games[i].Date = date

		insertGamesInfo(games[i])

		time.Sleep(time.Second * 2)

	}
}
func insertGamesInfo(game GameList) {
	db, err := sql.Open("mysql", "ew:4253@tcp(192.168.0.6:8889)/steam-info-db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	stmtInsert, err := db.Prepare("INSERT INTO game_info(app_id, app_name, release_date, price, is_free) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer stmtInsert.Close()

	id := game.Appid
	name := game.Name
	releaseDate := game.Date
	price := game.Price
	isFree := game.IsFree

	result, err := stmtInsert.Exec(id, name, releaseDate, price, isFree)
	if err != nil {
		panic(err.Error())
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(lastInsertID, id, name, releaseDate, price, isFree)

	upd, err := db.Prepare("UPDATE game_list set updated_at = ? where app_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	t := time.Now()
	ts := t.Format("2006-01-02 15:04:05")
	upd.Exec(ts, id)
}
