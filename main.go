package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
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
	var dbPort int = 3306
	var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, dbPort)

	tlsName := "rds"
	if err := registerTlsConfig("./ap-northeast-1-bundle.pem", tlsName); err != nil {
		panic(err)
	}

	pw := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=%s&allowCleartextPasswords=true",
		dbUser, pw, dbEndpoint, dbName, tlsName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err) // 証明書を設定しないと panic: x509: certificate signed by unknown authority が発生
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
	pem, err := ioutil.ReadFile(pemPath)
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
