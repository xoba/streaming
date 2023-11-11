package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xoba/open-golang/open"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	defer log.Printf("websocket done")
	log.Printf("handling websocket")
	SetCommonHeaders(w)
	if err := ws(w, r); err != nil {
		log.Printf("oops: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// pipes in -> c[0] -> c[1] -> ... -> c[n-1] -> out
func pipe(in io.Reader, out io.Writer, c ...*exec.Cmd) error {
	if len(c) == 0 {
		return nil
	}
	run := func(cmd *exec.Cmd, in io.Reader, out io.Writer) error {
		cmd.Stdin = in
		cmd.Stdout = out
		cmd.Stderr = os.Stderr
		return cmd.Start()
	}
	switch len(c) {
	case 0:
		return nil
	case 1:
		return run(c[0], in, out)
	default:
		r, w := io.Pipe()
		if err := run(c[0], in, w); err != nil {
			return err
		}
		return pipe(r, out, c[1:]...)
	}
}

func ws(w http.ResponseWriter, r *http.Request) error {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	c1 := exec.Command("ffmpeg", "-y", "-i", "-", "-filter:a", "atempo=0.75", "-f", "webm", "pipe:1")
	c2 := exec.Command("ffplay", "-")

	in, out := io.Pipe()

	if err := pipe(in, io.Discard, c1, c2); err != nil {
		return err
	}

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
			if _, writeErr := out.Write(p); writeErr != nil {
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
	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		return err
	}
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
