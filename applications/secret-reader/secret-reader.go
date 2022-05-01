package main

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"os"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/secrets"
)

const (
	secretOCID = "ocid1.vaultsecret.oc1.iad.amaaaaaa6sde7caazzhfhfsy2v6tqpr3velezxm4r7ld5alifmggjv3le2cq"
)

// type DatabaseConnectDetails struct {
// 	Username       string `json:username`
// 	Server         string `json:server`
// 	Port           string `json:port`
// 	Password       string `json:password`
// 	WalletLocation string `json:walletLocation`
// }

func main() {
	secretsClient, err := secrets.NewSecretsClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		fmt.Printf("failed to get secretsclient : %s", err)
	}
	secretReq := secrets.GetSecretBundleRequest{SecretId: common.String(secretOCID)}
	secretResponse, err := secretsClient.GetSecretBundle(context.Background(), secretReq)
	if err != nil {
		fmt.Printf("failed to get secretsbundle : %s", err)
	}
	contentDetails := secretResponse.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)
	decodedSecretContents, _ := b64.StdEncoding.DecodeString(*contentDetails.Content)
	fmt.Println("Secret Contents:", string(decodedSecretContents))
	err = os.WriteFile("./cwallet-from-secret.sso", decodedSecretContents, 0644)
	// var dbCredentials DatabaseConnectDetails

	// err = json.Unmarshal(decodedSecretContents, &dbCredentials)
	if err != nil {
		fmt.Printf("failed to write : %s", err)
	}
	// fmt.Println("username" + dbCredentials.Username)
}
