package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/xoba/open-golang/open"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handling websocket")
	SetCommonHeaders(w)
	if err := ws(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ws(w http.ResponseWriter, r *http.Request) error {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	fifoPath := fmt.Sprintf("fifo_%s.webm", uuid.NewString()[:8])

	if err := syscall.Mkfifo(fifoPath, 0644); err != nil {
		return err
	}
	if err := exec.Command("ffplay", fifoPath).Start(); err != nil {
		return err
	}

	fifo, err := os.OpenFile(fifoPath, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return err
	}
	defer fifo.Close()

	conn.WriteMessage(websocket.TextMessage, []byte("launched ffplay etc"))

	// Read messages from the WebSocket connection
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return err
			}
			break
		}
		switch messageType {
		case websocket.BinaryMessage:
			log.Printf("%d bytes received", len(p))
			if _, writeErr := fifo.Write(p); writeErr != nil {
				return err
			}
		case websocket.TextMessage:
			log.Printf("received text: %s\n", string(p))
			switch msg := string(p); msg {
			case "stop":
				return nil
			default:
				log.Printf("unhandled message: %q", msg)
			}
		}
	}
	return nil
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	SetCommonHeaders(w)
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
	http.HandleFunc("/ws", wsHandler)
	log.Printf("Server starting on :%d", port)
	go func() {
		time.Sleep(time.Second / 3)
		open.Run(fmt.Sprintf("http://localhost:%d", port))
	}()
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func SetCommonHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Add("Access-Control-Allow-Origin", "*")
	h.Add("Referrer-Policy", "no-referrer")
	h.Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	h.Add("X-Content-Type-Options", "nosniff")
	h.Add("X-Frame-Options", "SAMEORIGIN")
	h.Add("X-Permitted-Cross-Domain-Policies", "none")
	h.Add("X-XSS-Protection", "1; mode=block")
	h.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")
}
