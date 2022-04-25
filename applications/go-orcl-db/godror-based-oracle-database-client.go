package main

import (
	"database/sql"
	"fmt"

	_ "github.com/godror/godror"
)

func GetSqlDBWithGoDrOrDriver(dbParams map[string]string) *sql.DB {
	var db *sql.DB
	var err error
	if val, ok := dbParams["walletLocation"]; ok && val != "" {
		db, err = sql.Open("godror", fmt.Sprintf(`user="%s" password="%s"
		connectString="tcps://%s:%s/%s?wallet_location=%s"
		   `, dbParams["username"], dbParams["password"], dbParams["server"], dbParams["port"], dbParams["service"], dbParams["walletLocation"]))
	}
	if val, ok := dbParams["walletLocation"]; !ok || val == "" {
		connectionString := "oracle://" + dbParams["username"] + ":" + dbParams["password"] + "@" + dbParams["server"] + ":" + dbParams["port"] + "/" + dbParams["service"]
		db, err = sql.Open("oracle", connectionString)
	}

	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("error pinging db: %w", err))
	}
	return db
}
