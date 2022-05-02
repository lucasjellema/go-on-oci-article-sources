package main

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/secrets"

	"github.com/oracle/oci-go-sdk/v65/streaming"
)

var streamDetailsSecretOCID string

type StreamConnectDetails struct {
	StreamMessagesEndpoint string `json:"streamMessagesEndpoint"`
	StreamOCID             string `json:"streamOCID"`
}

func getStreamConnectDetails(ociConfigurationProvider common.ConfigurationProvider) StreamConnectDetails {
	secretsClient, err := secrets.NewSecretsClientWithConfigurationProvider(ociConfigurationProvider)
	if err != nil {
		fmt.Printf("failed to get secretsclient : %s", err)
	}
	secretReq := secrets.GetSecretBundleRequest{SecretId: common.String(streamDetailsSecretOCID)}
	secretResponse, _ := secretsClient.GetSecretBundle(context.Background(), secretReq)
	contentDetails := secretResponse.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)
	decodedSecretContents, _ := b64.StdEncoding.DecodeString(*contentDetails.Content)
	var streamConnectDetails StreamConnectDetails
	err = json.Unmarshal(decodedSecretContents, &streamConnectDetails)
	if err != nil {
		fmt.Printf("failed to unmarshal secret : %s", err)
	}
	return streamConnectDetails
}

const MAX_AGE = 90

func getFirstNames() []string {
	return []string{"Hans", "Brian", "Janet", "Wilma", "Barry", "Wodan", "Betty", "Daisy", "Caroline", "Karen", "Fonz", "Richard", "Thomas", "Frank", "Doris", "Michael", "Joel", "Taylor"}
}

func main() {
	fmt.Println("Welcome to the Person Producer from Deep Down in the Container - About to publish some person records to the stream")
	streamDetailsSecretOCID = os.Getenv("STREAM_DETAILS_SECRET_OCID")
	if streamDetailsSecretOCID == "" {
		fmt.Printf("No value set for environment variable STREAM_DETAILS_SECRET_OCID")
		panic("No value set for environment variable STREAM_DETAILS_SECRET_OCID")
	}
	var ociConfigurationProvider common.ConfigurationProvider
	var err error
	if os.Getenv("INSTANCE_PRINCIPAL_AUTHENTICATION") == "NO" {
		fmt.Println("INSTANCE_PRINCIPAL_AUTHENTICATION == NO; OCI_CONFIG_FILE (only relevant when not doing instance principal authentication):", os.Getenv("OCI_CONFIG_FILE"))
		ociConfigurationProvider = common.DefaultConfigProvider()
	} else {
		fmt.Println("Relying on Instance Principal Authentication")
		ociConfigurationProvider, err = auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			fmt.Printf("failed to create InstancePrincipalConfigurationProvider : %s", err)
			panic(err)
		}
	}
	streamConnectDetails := getStreamConnectDetails(ociConfigurationProvider)
	streamClient, err := streaming.NewStreamClientWithConfigurationProvider(ociConfigurationProvider, streamConnectDetails.StreamMessagesEndpoint)
	if err != nil {
		fmt.Printf("failed to create streamClient : %s", err)
	}
	firstNames := getFirstNames()
	for i := 0; i < 5; i++ {
		person := Person{Name: getFirstNames()[rand.Intn(len(firstNames))], Age: rand.Intn(MAX_AGE) + 3, JuicyDetails: "created from canned Person Producer application at " + time.Now().String()}
		producePersonMessage(person, streamClient, streamConnectDetails.StreamOCID)
		time.Sleep(time.Second * 5)
	}
}

type Person struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	JuicyDetails string `json:"comment"`
}

func producePersonMessage(person Person, streamClient streaming.StreamClient, streamOCID string) {
	personMessage, err := json.Marshal(person)
	if err != nil {
		fmt.Println("Producing JSON message failed ", err)
	}
	putMessagesRequest := streaming.PutMessagesRequest{StreamId: common.String(streamOCID),
		PutMessagesDetails: streaming.PutMessagesDetails{
			Messages: []streaming.PutMessagesDetailsEntry{
				{Key: []byte(person.Name),
					Value: personMessage},
			},
		},
	}
	// Send the request using the service client
	putMsgResp, err := streamClient.PutMessages(context.Background(), putMessagesRequest)
	if err != nil {
		fmt.Println("Sad, we ran into an error: ", err)
	}

	// Retrieve value from the response.
	fmt.Println(putMsgResp)

}
