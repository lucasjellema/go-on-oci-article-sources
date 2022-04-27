package main

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/streaming"
)

const (
	streamMessagesEndpoint = "https://cell-1.streaming.us-ashburn-1.oci.oraclecloud.com"
	streamOCID             = "ocid1.stream.oc1.iad.amaaaaaa6sde7caa56brreqvzptc37wytom7pjk7vx3qaflagk2t3syvk67q"
)

func main() {
	streamClient, err := streaming.NewStreamClientWithConfigurationProvider(common.DefaultConfigProvider(), streamMessagesEndpoint)
	if err != nil {
		fmt.Printf("failed to create streamClient : %s", err)
	}
	// Type can be CreateGroupCursorDetailsTypeTrimHorizon, CreateGroupCursorDetailsTypeAtTime, CreateGroupCursorDetailsTypeLatest
	createGroupCursorRequest := streaming.CreateGroupCursorRequest{
		StreamId: common.String(streamOCID),
		CreateGroupCursorDetails: streaming.CreateGroupCursorDetails{Type: streaming.CreateGroupCursorDetailsTypeTrimHorizon,
			CommitOnGet:  common.Bool(true), // when false, a consumer must manually commit their cursors (to move the offset).
			GroupName:    common.String("consumer-group-1"),
			InstanceName: common.String("go-instance-1"), // A unique identifier for the instance joining the consumer group. If an instanceName is not provided, a UUID will be generated
			TimeoutInMs:  common.Int(1000),
		}}

	createGroupCursorResponse, err := streamClient.CreateGroupCursor(context.Background(), createGroupCursorRequest)
	if err != nil {
		fmt.Println(err)
	}
	consumeMessagesLoop(streamClient, streamOCID, *createGroupCursorResponse.Value)

	// try again
	partition := "0"
	offset := common.Int64(5)
	// createCursorRequest := streaming.CreateCursorRequest{
	// 	StreamId: common.String(streamOCID),
	// 	CreateCursorDetails: streaming.CreateCursorDetails{Type: streaming.CreateCursorDetailsTypeTrimHorizon,
	// 		Partition: &partition,
	// 	}}
	createCursorRequest := streaming.CreateCursorRequest{
		StreamId: common.String(streamOCID),
		CreateCursorDetails: streaming.CreateCursorDetails{Type: streaming.CreateCursorDetailsTypeAfterOffset,
			Offset:    offset,
			Partition: &partition,
		}}

	// Send the request using the service client
	createCursorResponse, err := streamClient.CreateCursor(context.Background(), createCursorRequest)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(createCursorResponse)
	consumeMessagesLoop(streamClient, streamOCID, *createCursorResponse.Value)

}

func consumeMessagesLoop(streamClient streaming.StreamClient, streamOcid string, cursorValue string) {
	getMessagesFromCursorRequest := streaming.GetMessagesRequest{Limit: common.Int(5), // optional: The maximum number of messages to return, any value up to 10000. By default, the service returns as many messages as possible.
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
			fmt.Println("Key : " + string(message.Key) + ", value : " + string(message.Value) + ", Partition " + *message.Partition)
		}
		cursorValue = *getMessagesFromCursorResponse.OpcNextCursor
	}
}
