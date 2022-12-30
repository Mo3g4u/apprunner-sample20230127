package main

import (
	"context"
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

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
)

func main() {
	var dbName string = "bookcase"
	var dbUser string = "bookuser"
	var dbHost string = "database-1.cluster-cxw8iq8t33nv.ap-northeast-1.rds.amazonaws.com"
	var dbPort int = 3306
	var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, dbPort)
	var region string = "ap-northeast-1"

	tlsName := "rds"
	if err := registerTlsConfig("./ap-northeast-1-bundle.pem", tlsName); err != nil {
		panic(err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	authenticationToken, err := auth.BuildAuthToken(
		context.TODO(), dbEndpoint, region, dbUser, cfg.Credentials)
	if err != nil {
		panic("failed to create authentication token: " + err.Error())
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=%s&allowCleartextPasswords=true",
		dbUser, authenticationToken, dbEndpoint, dbName, tlsName,
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
		io.WriteString(w, "Hello AppRunner!\n")
	}
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
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
