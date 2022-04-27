package main

import (
	"context"
	"fmt"
	"strconv"

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
	// Create five putMessage requests with each two enclosed messages.
	for i := 0; i < 5; i++ {
		putMsgReq := streaming.PutMessagesRequest{StreamId: common.String(streamOCID),
			PutMessagesDetails: streaming.PutMessagesDetails{
				// we are batching 2 messages for each Put Request
				Messages: []streaming.PutMessagesDetailsEntry{
					{Key: []byte("key dummy-0-" + strconv.Itoa(i)),
						Value: []byte("my happy message-" + strconv.Itoa(i))},
					{Key: []byte("key dummy-1-" + strconv.Itoa(i)),
						Value: []byte("hello dolly and others-" + strconv.Itoa(i))}}},
		}

		// Send the request using the service client
		putMsgResp, err := streamClient.PutMessages(context.Background(), putMsgReq)
		if err != nil {
			fmt.Println("Sad, we ran into an error: ", err)
		}

		// Retrieve value from the response.
		fmt.Println(putMsgResp)
	}
}
