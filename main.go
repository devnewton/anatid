package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"./tribune"
	"github.com/gorilla/websocket"
)

var listenAddress string
var verboseMode bool

func init() {
	flag.StringVar(&listenAddress, "listen", ":6666", "TCP address to listen on")
	flag.BoolVar(&verboseMode, "verbose", false, "Verbose logging")
}

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
	tribunes map[string]*tribune.Tribune
	poll     chan *tribune.Tribune
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
		tribunes: map[string]*tribune.Tribune{
			"euromussels": &tribune.Tribune{Name: "euromussels", BackendURL: "https://faab.euromussels.eu/data/backend.tsv", PostURL: "https://faab.euromussels.eu/add.php", PostField: "message"},
			"sveetch":     &tribune.Tribune{Name: "sveetch", BackendURL: "http://sveetch.net/tribune/remote/tsv/", PostURL: "http://sveetch.net/tribune/post/tsv/?last_id=1", PostField: "content"},
			"moules":      &tribune.Tribune{Name: "moules", BackendURL: "http://moules.org/board/backend/tsv", PostURL: "http://moules.org/board/add.php?backend=tsv", PostField: "message"},
			"ototu":       &tribune.Tribune{Name: "ototu", BackendURL: "https://ototu.euromussels.eu/goboard/backend/tsv", PostURL: "https://ototu.euromussels.eu/goboard/post", PostField: "message"},
			"dlfp":        &tribune.Tribune{Name: "dlfp ", BackendURL: "https://linuxfr.org/board/index.tsv", PostURL: "https://linuxfr.org/api/v1/board", PostField: "message", AuthentificationType: tribune.OAuth2Authentification},
			"batavie":     &tribune.Tribune{Name: "batavie ", BackendURL: "http://batavie.leguyader.eu/remote.xml", PostURL: "http://batavie.leguyader.eu/index.php/add", PostField: "message", BackendType: tribune.XMLBackend},
		},
		poll: make(chan *tribune.Tribune),
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

func (a *anatid) handlePoll(w http.ResponseWriter, r *http.Request) {
	ws, err := a.upgrader.Upgrade(w, r, nil)
	if nil != err {
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

func (a *anatid) handlePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if nil != err {
		log.Println(err)
		return
	}
	t := a.tribunes[r.PostFormValue("tribune")]
	if nil != t {
		t.Post(r)
		a.poll <- t
	}
}

func (a *anatid) pollTribunes() {
	for _, t := range a.tribunes {
		err := a.pollTribune(t)
		if nil != err {
			log.Println(err)
			continue
		}
	}
}

func (a *anatid) pollTribune(t *tribune.Tribune) error {
	if verboseMode {
		log.Printf("Poll %s\n", t.Name)
	}
	posts, err := t.Poll()
	if nil != err {
		return err
	}
	postsJSON, err := json.Marshal(posts)
	if nil != err {
		return err
	}
	a.forward <- postsJSON
	return nil
}

func (a *anatid) pollLoop() {
	tick := time.Tick(30 * time.Second)
	for {
		select {
		case t := <-a.poll:
			a.pollTribune(t)
		case <-tick:
			a.pollTribunes()
		}
	}
}

func main() {
	flag.Parse()
	a := newAnatid()
	go a.forwardLoop()
	go a.pollLoop()
	http.HandleFunc("/anatid/poll", func(w http.ResponseWriter, r *http.Request) {
		a.handlePoll(w, r)
	})
	http.HandleFunc("/anatid/post", func(w http.ResponseWriter, r *http.Request) {
		a.handlePost(w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to anatid server."))
	})
	log.Printf("Listen to %s\n", listenAddress)
	err := http.ListenAndServe(listenAddress, nil)
	if nil != err {
		log.Fatal(err)
	}
}
