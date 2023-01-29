package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var db *sql.DB

func main() {
	var dbName string = "bookcase"
	var dbUser string = "admin"
	var dbHost string = "database-1.cluster-cxw8iq8t33nv.ap-northeast-1.rds.amazonaws.com"
	var dbPort string = "3306"
	pw := os.Getenv("DB_PASSWORD")

	db, err := sql.Open("mysql", dbUser+":"+pw+"@tcp("+dbHost+":"+dbPort+")/"+dbName)
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
