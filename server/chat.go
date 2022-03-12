package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

type Message struct {
	Id      string `json:"id"`
	Message string `json:"message"`
	Name    string `json:"name"`
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			var msgParsed Message
			err := json.Unmarshal(message, &msgParsed)
			if err != nil {
				panic(err)
			}
			for client := range h.clients {
				if client.id == msgParsed.Id {
					continue
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func main() {
	port := flag.String("port", "80", "port to listen")
	flag.Parse()

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	hub := newHub()
	go hub.run()
	/*http.HandleFunc("/client", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/client" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<input id="input" type="text" />
			<button onclick="send()">Send</button>
			<pre id="output"></pre>
			<script>
				var input = document.getElementById("input");
				var output = document.getElementById("output");
				var socket = new WebSocket("ws://localhost:80/echo");

				socket.onopen = function () {
					output.innerHTML += "Status: Connected\n";
				};

				socket.onmessage = function (e) {
					output.innerHTML += "Server: " + e.data + "\n";
				};

				function send() {
					socket.send(input.value);
					input.value = "";
				}
			</script>`)
		} else {
			http.Error(w, "404", http.StatusNotFound)
		}
	})*/
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("NEW CONNECTION")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, msg, err := conn.ReadMessage()
		// conn.RemoteAddr()
		if err != nil {
			fmt.Println(err)
			return
		}
		name := string(msg)
		clid := uuid.New().String()
		fmt.Println("new client: ", name, clid)
		serveWs(conn, hub, w, r, name, clid)
	})
	fmt.Println("starting vchat service on port:", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Println(err)
	}
}
