package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"

	fdk "github.com/fnproject/fdk-go"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(myHandler))
}

func myHandler(ctx context.Context, in io.Reader, out io.Writer) {
	objectName := "defaultObjectName.txt"
	bucketName := "the-bucket"
	fnctx := fdk.GetContext(ctx)             // fnctx contains relevant elements about the Function itself
	fnhttpctx, ok := fnctx.(fdk.HTTPContext) // fnhttpctx contains relevant elements about the HTTP Request that triggered it
	if ok {                                  // an HTTPContent was found which means that this was an HTTP request (not an fn invoke) that triggered the function
		u, err := url.Parse(fnhttpctx.RequestURL())
		if err == nil {
			// queries := u.Query()
			for key, value := range u.Query() {
				if key == "objectName" {
					objectName = value[0]
				}
				if key == "bucketName" {
					bucketName = value[0]
				}
			}
		}
	}
	var message string
	if compartmentOCID, ok := fnctx.Config()["compartmentOCID"]; ok { // assuming an Application or Function Configuration Parameter called compartmentOCID has been defined
		msg, err := CreateObject(objectName, bucketName, compartmentOCID)
		if err != nil {
			message = fmt.Sprintf("Error in function execution: %s", err)
		} else {
			message = msg
		}

	} else {
		message = fmt.Sprintf("Configuration parameter compartmentOCID was not found for the function or its application")
	}

	response := struct {
		Msg string `json:"message"`
	}{
		Msg: message,
	}
	err := json.NewEncoder(out).Encode(&response)
	if err != nil {
		log.Print("Error occurred in function during response encoding ", err)
	}

}
