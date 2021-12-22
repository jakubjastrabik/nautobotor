package main

import (
	"log"
	"nautobotor/webserver"
)

func main() {
	log.Println("Start app")

	// Call HTTP server with separated trade
	webserver.HttpServer()
}
