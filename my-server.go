package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	DEFAULT_HTTP_SERVER_PORT = "8080"
	ENV_KEY_HTTP_SERVER_PORT = "HTTP_SERVER_PORT"
	ENV_KEY_MYSERVER_VERSION = "VERSION_OF_MYSERVER"
)

const (
	GREET_PATH       = "/greet"
	STATIC_SITE_PATH = "/site/"
	ROOT_PATH        = "/"
)

func ComposeGreeting(name string) string {
	nameToGreet := "Stranger"
	if len(name) > 0 {
		nameToGreet = name
		log.Printf(" --  query parameter name is set to %s", name)
	}
	return fmt.Sprintf("Hello %s!", nameToGreet)
}

func greetHandler(response http.ResponseWriter, request *http.Request) {
	log.Printf("Handle Request for method %s on path %s", request.Method, request.URL.Path)
	if request.Method != "GET" {
		http.Error(response, "Method is not supported unfortunately. ", http.StatusNotFound)
		return
	}
	queryName := request.URL.Query().Get("name")
	fmt.Fprint(response, ComposeGreeting(queryName))
}

func fallbackHandler(response http.ResponseWriter, request *http.Request) {
	log.Printf("Warning: Request for unhandled method %s on path %s", request.Method, request.URL.Path)
	http.Error(response, "404 path/method combination not currently supported. Try /greet or /site", http.StatusNotFound)
}

func main() {
	httpServerPort, ok := os.LookupEnv(ENV_KEY_HTTP_SERVER_PORT)
	if !ok {
		httpServerPort = DEFAULT_HTTP_SERVER_PORT
		log.Printf("Environment Variable %s not set; using default value: %s", ENV_KEY_HTTP_SERVER_PORT, DEFAULT_HTTP_SERVER_PORT)
	}
	myserverVersion, ok := os.LookupEnv(ENV_KEY_MYSERVER_VERSION)
	if !ok {
		myserverVersion = "unknown"
		log.Printf("Environment Variable %s not set", ENV_KEY_MYSERVER_VERSION)
	}
	fileServer := http.FileServer(http.Dir("./website"))
	http.Handle(STATIC_SITE_PATH, http.StripPrefix("/site/", fileServer))
	http.HandleFunc(GREET_PATH, greetHandler)
	http.HandleFunc(ROOT_PATH, fallbackHandler)

	log.Printf("Starting my-server (version %s) listening for requests at port %s\n", myserverVersion, httpServerPort)
	if err := http.ListenAndServe(":"+httpServerPort, nil); err != nil {
		log.Fatal(err)
	}
}
