package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// We will add the WebSocket Hub here later.
	// For now, just a health check.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "DocStream Backend is Running!")
	})

	port := "8080"
	fmt.Printf("Server started on port %s\n", port)

	// Start the server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
