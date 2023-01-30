package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var db *sql.DB

type RDS struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Engine              string `json:"engine"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	DbClusterIdentifier string `json:"dbClusterIdentifier"`
}

func main() {
	dbName := "bookcase"
	jsonStr := os.Getenv("DB_SETTINGS")
	var rds RDS
	if err := json.Unmarshal([]byte(jsonStr), &rds); err != nil {
		panic(err)
	}

	db, err := sql.Open("mysql", rds.Username+":"+rds.Password+"@tcp("+rds.Host+":"+string(rds.Port)+")/"+dbName)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	handler := func(w http.ResponseWriter, _ *http.Request) {
		// SQLの実行
		rows, err := db.Query("SELECT * FROM authors")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		names := []string{}
		// SQLの実行
		for rows.Next() {
			var author Author
			if err := rows.Scan(&author.authorId, &author.name); err != nil {
				panic(err.Error())
			}
			names = append(names, author.name)
		}

		io.WriteString(w, fmt.Sprintf("Hello AppRunner! %s", strings.Join(names, ",")))
	}
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Author struct {
	authorId int
	name     string
}
