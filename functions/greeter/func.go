package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	fdk "github.com/fnproject/fdk-go"
)

func main() {
	fdk.Handle(fdk.HandlerFunc(myHandler))
}

type Person struct {
	Name string `json:"name"`
}

func myHandler(ctx context.Context, in io.Reader, out io.Writer) {
	p := &Person{Name: "World"}
	message := ""
	err := json.NewDecoder(in).Decode(p)
	if err != nil {
		message = fmt.Sprintf("Error on your input:  %s", err)
	} else {
		message = fmt.Sprintf("Extremely hot greetings from your automatically built and deployed function dear  %s", p.Name)
	}
	msg := struct {
		Msg string `json:"message"`
	}{
		Msg: message,
	}
	log.Print("Inside Go Greeter function new version")

	err = json.NewEncoder(out).Encode(&msg)
	if err != nil {
		log.Print("Error occurred in function greeter ", err)
	}

}
