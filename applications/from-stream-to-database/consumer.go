package main

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/secrets"
	"github.com/oracle/oci-go-sdk/v65/streaming"
)

const (
	streamDetailsSecretOCID = "ocid1.vaultsecret.oc1.iad.amaaaaaa6sde7caa6m5tuweeu3lbz22lf37y2dsbdojnhz2owmgvqgwwnvka"
)

type StreamConnectDetails struct {
	StreamMessagesEndpoint string `json:streamMessagesEndpoint`
	StreamOCID             string `json:streamOCID`
}

func getStreamConnectDetails() StreamConnectDetails {
	secretsClient, err := secrets.NewSecretsClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		fmt.Printf("failed to get secretsclient : %s", err)
	}
	secretReq := secrets.GetSecretBundleRequest{SecretId: common.String(streamDetailsSecretOCID)}
	secretResponse, _ := secretsClient.GetSecretBundle(context.Background(), secretReq)
	contentDetails := secretResponse.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)
	decodedSecretContents, _ := b64.StdEncoding.DecodeString(*contentDetails.Content)
	var streamConnectDetails StreamConnectDetails
	json.Unmarshal(decodedSecretContents, &streamConnectDetails)
	return streamConnectDetails
}

func main() {
	streamConnectDetails := getStreamConnectDetails()
	streamClient, err := streaming.NewStreamClientWithConfigurationProvider(common.DefaultConfigProvider(), streamConnectDetails.StreamMessagesEndpoint)
	if err != nil {
		fmt.Printf("failed to create streamClient : %s", err)
	}
	database = InitializeDatabase()
	defer func() {
		err := database.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()

	// Type can be CreateGroupCursorDetailsTypeTrimHorizon, CreateGroupCursorDetailsTypeAtTime, CreateGroupCursorDetailsTypeLatest
	createGroupCursorRequest := streaming.CreateGroupCursorRequest{
		StreamId: common.String(streamConnectDetails.StreamOCID),
		CreateGroupCursorDetails: streaming.CreateGroupCursorDetails{Type: streaming.CreateGroupCursorDetailsTypeLatest, // only consume messages produced after starting the consumer
			CommitOnGet: common.Bool(true), // when false, a consumer must manually commit their cursors (to move the offset).
			GroupName:   common.String("person-message-1"),
			TimeoutInMs: common.Int(1000),
		}}

	createGroupCursorResponse, err := streamClient.CreateGroupCursor(context.Background(), createGroupCursorRequest)
	if err != nil {
		fmt.Println(err)
	}
	consumeMessagesLoop(streamClient, streamConnectDetails.StreamOCID, *createGroupCursorResponse.Value)
}

func consumeMessagesLoop(streamClient streaming.StreamClient, streamOcid string, cursorValue string) {
	getMessagesFromCursorRequest := streaming.GetMessagesRequest{Limit: common.Int(15), // optional: The maximum number of messages to return, any value up to 10000. By default, the service returns as many messages as possible.
		StreamId: common.String(streamOcid),
		Cursor:   common.String(cursorValue)}
	for i := 0; i < 15; i++ {
		fmt.Println("starting iteration ", i)
		getMessagesFromCursorRequest.Cursor = common.String(cursorValue)
		// Send the request using the service client
		getMessagesFromCursorResponse, err := streamClient.GetMessages(context.Background(), getMessagesFromCursorRequest)
		if err != nil {
			fmt.Println(err)
		}
		for _, message := range getMessagesFromCursorResponse.Items {
			fmt.Println("Message consumed with Key : " + string(message.Key) + ", value : " + string(message.Value) + ", Partition " + *message.Partition)
			processPersonMessage(message.Value)
		}
		cursorValue = *getMessagesFromCursorResponse.OpcNextCursor
		// Pause for 10 seconds before doing the next iteration and poll
		time.Sleep(10 * time.Second)
	}
}

type Person struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	JuicyDetails string `json:"comment"`
}

func processPersonMessage(message []byte) {
	var person Person
	err := json.Unmarshal(message, &person)
	if err != nil {
		fmt.Println(err)
	}
	PersistPerson(person)
}
