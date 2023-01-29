package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
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

	tlsName := "rds"
	if err := registerTlsConfig("./ap-northeast-1-bundle.pem", tlsName); err != nil {
		panic(err)
	}

	db, err := sql.Open("mysql", dbUser+":"+pw+"@tcp("+dbHost+":"+dbPort+")/"+dbName+"?tls=rds&allowCleartextPasswords=true")
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

func registerTlsConfig(pemPath, tlsConfigKey string) (err error) {
	caCertPool := x509.NewCertPool()
	pem, err := os.ReadFile(pemPath)
	if err != nil {
		return
	}

	if ok := caCertPool.AppendCertsFromPEM(pem); !ok {
		return errors.New("pem error")
	}
	mysql.RegisterTLSConfig(tlsConfigKey, &tls.Config{
		ClientCAs:          caCertPool,
		InsecureSkipVerify: true, // 必要に応じて
	})

	return
}
