package main

import (
	"context"
	"database/sql"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/secrets"
)

type DatabaseConnectDetails struct {
	Service        string `json:service`
	Username       string `json:username`
	Server         string `json:server`
	Port           string `json:port`
	Password       string `json:password`
	WalletLocation string `json:walletLocation`
}

const (
	autonomousDatabaseConnectDetailsSecretOCID = "ocid1.vaultsecret.oc1.iad.amaaaaaa6sde7caabn37hbdsu7dczk6wpxvr7euq7j5fmti2zkjcpwzlmowq"
	autonomousDatabaseCwalletSsoSecretOCID     = "ocid1.vaultsecret.oc1.iad.amaaaaaa6sde7caazzhfhfsy2v6tqpr3velezxm4r7ld5alifmggjv3le2cq"
	walletLocation                             = "."
	walletFile                                 = "cwallet.sso"
)

func main() {
	initializeWallet()
	dbConnectDetails := getDatabaseConnectDetails()
	dbConnectDetails.WalletLocation = walletLocation
	db := GetSqlDBWithGoDrOrDriver(dbConnectDetails)
	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()
	sqlOperations(db)
	fmt.Println("DONE")
}

func getDatabaseConnectDetails() DatabaseConnectDetails {
	secretsClient, err := secrets.NewSecretsClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		fmt.Printf("failed to get secretsclient : %s", err)
	}
	secretReq := secrets.GetSecretBundleRequest{SecretId: common.String(autonomousDatabaseConnectDetailsSecretOCID)}
	secretResponse, _ := secretsClient.GetSecretBundle(context.Background(), secretReq)
	contentDetails := secretResponse.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)
	decodedSecretContents, _ := b64.StdEncoding.DecodeString(*contentDetails.Content)
	var dbCredentials DatabaseConnectDetails
	json.Unmarshal(decodedSecretContents, &dbCredentials)
	return dbCredentials
}

func initializeWallet() {
	secretsClient, err := secrets.NewSecretsClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		fmt.Printf("failed to get secretsclient : %s", err)
	}
	secretReq := secrets.GetSecretBundleRequest{SecretId: common.String(autonomousDatabaseCwalletSsoSecretOCID)}
	secretResponse, _ := secretsClient.GetSecretBundle(context.Background(), secretReq)
	contentDetails := secretResponse.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)
	decodedSecretContents, _ := b64.StdEncoding.DecodeString(*contentDetails.Content)
	_ = os.WriteFile(fmt.Sprintf("%s/%s", walletLocation, walletFile), decodedSecretContents, 0644)
}

const createTableStatement = "CREATE TABLE TEMP_TABLE ( NAME VARCHAR2(100), CREATION_TIME TIMESTAMP DEFAULT SYSTIMESTAMP, VALUE  NUMBER(5))"
const dropTableStatement = "DROP TABLE TEMP_TABLE PURGE"
const insertStatement = "INSERT INTO TEMP_TABLE ( NAME , VALUE) VALUES (:name, :value)"
const queryStatement = "SELECT name, creation_time, value FROM TEMP_TABLE"

func sqlOperations(db *sql.DB) {
	_, err := db.Exec(createTableStatement)
	handleError("create table", err)
	defer db.Exec(dropTableStatement) // make sure the table is removed when all is said and done
	stmt, err := db.Prepare(insertStatement)
	handleError("prepare insert statement", err)
	sqlresult, err := stmt.Exec("John", 42)
	handleError("execute insert statement", err)
	rowCount, _ := sqlresult.RowsAffected()
	fmt.Println("Inserted number of rows = ", rowCount)

	var queryResultName string
	var queryResultTimestamp time.Time
	var queryResultValue int32
	row := db.QueryRow(queryStatement)
	err = row.Scan(&queryResultName, &queryResultTimestamp, &queryResultValue)
	handleError("query single row", err)
	if err != nil {
		panic(fmt.Errorf("error scanning db: %w", err))
	}
	fmt.Println(fmt.Sprintf("The name: %s, time: %s, value:%d ", queryResultName, queryResultTimestamp, queryResultValue))
	_, err = stmt.Exec("Jane", 69)
	handleError("execute insert statement", err)
	_, err = stmt.Exec("Malcolm", 13)
	handleError("execute insert statement", err)

	// fetching multiple rows
	theRows, err := db.Query(queryStatement)
	handleError("Query for multiple rows", err)
	defer theRows.Close()
	var (
		name  string
		value int32
		ts    time.Time
	)
	for theRows.Next() {
		err := theRows.Scan(&name, &ts, &value)
		handleError("next row in multiple rows", err)
		fmt.Println(fmt.Sprintf("The name: %s and value:%d created at time: %s ", name, value, ts))
	}
	err = theRows.Err()
	handleError("next row in multiple rows", err)
}

func handleError(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		//os.Exit(1)
	}
}
