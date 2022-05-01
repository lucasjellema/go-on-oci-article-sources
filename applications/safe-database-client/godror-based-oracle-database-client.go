package main

import (
	"database/sql"
	"fmt"

	_ "github.com/godror/godror"
)

func GetSqlDBWithGoDrOrDriver(dbConnectDetails DatabaseConnectDetails) *sql.DB {
	var db *sql.DB
	var err error
	if val := dbConnectDetails.WalletLocation; val != "" {
		db, err = sql.Open("godror", fmt.Sprintf(`user="%s" password="%s"
		connectString="tcps://%s:%s/%s?wallet_location=%s"
		   `, dbConnectDetails.Username, dbConnectDetails.Password, dbConnectDetails.Server, dbConnectDetails.Port, dbConnectDetails.Service, dbConnectDetails.WalletLocation))
	}
	if val := dbConnectDetails.WalletLocation; val == "" {
		connectionString := "oracle://" + dbConnectDetails.Username + ":" + dbConnectDetails.Password + "@" + dbConnectDetails.Server + ":" + dbConnectDetails.Port + "/" + dbConnectDetails.Service
		db, err = sql.Open("oracle", connectionString)
	}

	err = db.Ping()
	if err != nil {
		panic(fmt.Errorf("error pinging db: %w", err))
	}
	return db
}
