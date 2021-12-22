package webserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"nautobotor/nautobot"
	"net/http"
)

// handleWebhook are used to processed nautobot webhook
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: err=%s\n", err)
		return
	}
	defer r.Body.Close()

	// Unmarshal data to strcut
	ip := nautobot.NewIPaddress(payload)

	fmt.Println(ip)
}

// httpServer handle web server with routing
func HttpServer() {
	// API routes
	http.HandleFunc("/webhook", handleWebhook)

	// Start server on port specified bellow
	port := ":8080"
	log.Fatal(http.ListenAndServe(port, nil))
}
