package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xoba/open-golang/open"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handling websocket")
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
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
		switch messageType {
		case websocket.BinaryMessage:
			// Write the binary data to the file
			log.Printf("%d bytes received", len(p))
			if _, writeErr := outFile.Write(p); writeErr != nil {
				fmt.Println(err)
				return
			}
		case websocket.TextMessage:
			log.Printf("text: %s\n", string(p))
		}
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("uri: %q", r.RequestURI)
	switch r.URL.Path {
	case "/":
		http.ServeFile(w, r, "index.html")
	case "/script.js":
		http.ServeFile(w, r, "script.js")
	case "/style.css":
		http.ServeFile(w, r, "style.css")
	default:
		http.NotFound(w, r)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	const port = 8080
	http.HandleFunc("/", webHandler)
	http.HandleFunc("/echo", wsHandler)
	log.Printf("Server starting on :%d", port)
	go func() {
		time.Sleep(time.Second / 3)
		open.Run(fmt.Sprintf("http://localhost:%d", port))
	}()
	return http.ListenAndServe(":8080", nil)
}
