package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("ws starting")
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // Allow connections from any origin
	conn, err := upgrader.Upgrade(w, r, nil)                          // Upgrade the HTTP server connection to the WebSocket protocol
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Create a file to save the audio
	outFile, err := os.Create("output_audio.webm") // Change the file extension based on the audio format you expect to receive
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outFile.Close()

	// Read messages from the WebSocket connection
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break // The client closed the connection
		}
		if messageType == websocket.BinaryMessage {
			// Write the binary data to the file
			log.Printf("%d bytes received", len(p))
			if _, writeErr := outFile.Write(p); writeErr != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("web starting: %q", r.RequestURI)

	switch r.URL.Path {
	case "/":
		http.ServeFile(w, r, "index.html")
	case "/script.js":
		http.ServeFile(w, r, "script.js")
	default:
		http.NotFound(w, r)
	}
}

func main() {
	http.HandleFunc("/", webHandler)    // Set the route handler
	http.HandleFunc("/echo", wsHandler) // Set the route handler
	fmt.Println("Server started on :8080")
	panic(http.ListenAndServe(":8080", nil))
}
