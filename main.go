package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"./tribune"
	"github.com/gorilla/websocket"
)

type client struct {
	ws   *websocket.Conn
	send chan []byte
}

func newClient(ws *websocket.Conn) *client {
	return &client{ws: ws, send: make(chan []byte, 8)}
}

func (c *client) readLoop() {
	defer c.ws.Close()
	c.ws.SetReadLimit(1024)
	for {
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *client) writeLoop() {
	for msg := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

type anatid struct {
	clients  map[*client]struct{}
	join     chan *client
	leave    chan *client
	forward  chan []byte
	upgrader websocket.Upgrader
}

func newAnatid() *anatid {
	return &anatid{
		clients: make(map[*client]struct{}),
		join:    make(chan *client),
		leave:   make(chan *client),
		forward: make(chan []byte),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 65536,
		},
	}
}

func (a *anatid) forwardLoop() {
	for {
		select {
		case c := <-a.join:
			a.clients[c] = struct{}{}
		case c := <-a.leave:
			delete(a.clients, c)
			close(c.send)
		case msg := <-a.forward:
			for c := range a.clients {
				select {
				case c.send <- msg:
				default:
					delete(a.clients, c)
					close(c.send)
				}
			}
		}
	}
}

func (a *anatid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	c := newClient(ws)
	a.join <- c
	defer func() { a.leave <- c }()
	go c.writeLoop()
	c.readLoop()
}

func (a *anatid) pollTribunes() {
	posts, err := tribune.PollTsv("https://faab.euromussels.eu/data/backend.tsv", "euromussels")
	if nil != err {
		log.Println(err)
		return
	}
	postsJSON, err := json.Marshal(posts)
	if nil != err {
		log.Println(err)
	}
	a.forward <- postsJSON
	log.Print("europosts: ")
	log.Println(len(posts))

}

func (a *anatid) pollLoop() {
	tick := time.Tick(30 * time.Second)
	for {
		select {
		case <-tick:
			a.pollTribunes()
		}
	}
}

func main() {
	a := newAnatid()
	go a.forwardLoop()
	go a.pollLoop()
	http.Handle("/anatid", a)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to anatid server."))
	})
	err := http.ListenAndServe(":6666", nil)
	if nil != err {
		log.Fatal(err)
	}
}
