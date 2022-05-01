package main

import (
	"context"
	"database/sql"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

var database *sql.DB

func InitializeDatabase() *sql.DB {
	initializeWallet()
	dbConnectDetails := getDatabaseConnectDetails()
	dbConnectDetails.WalletLocation = walletLocation
	db := GetSqlDBWithGoDrOrDriver(dbConnectDetails)
	database = db
	return database
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

const (
	PEOPLE_TABLE_NAME = "PEOPLE"
)

func PersistPerson(person Person) {
	ctx := context.Background()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = mergePerson(ctx, tx, person)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Merged record in table %s for person %s", PEOPLE_TABLE_NAME, person.Name)
	}
}

func mergePerson(ctx context.Context, tx *sql.Tx, person Person) error {
	mergeStatement := fmt.Sprintf(
		`MERGE INTO %s t using (select :name name, :age age, :description description from dual) person
		ON (t.name = person.name )
		WHEN MATCHED THEN UPDATE SET age = person.age, description = person.description
		WHEN NOT MATCHED THEN INSERT (t.name, t.age, t.description) values (person.name, person.age, person.description) `,
		PEOPLE_TABLE_NAME)
	_, err := tx.ExecContext(ctx, mergeStatement, person.Name, person.Age, person.JuicyDetails)
	return err
}
