package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	conn_counter = 0
	connections  map[int]*websocket.Conn
	addr         = flag.String("addr", ":8080", "websocket server address")
	upgrader     = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // CORS: Allow all origins
		},
	}
)

func main() {
	connections = make(map[int]*websocket.Conn)
	flag.Parse()
	go serverInput()
	http.HandleFunc("/", wsHandler)
	log.Printf("Started listening on %s\n", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	conn_counter++
	cid := conn_counter
	connections[conn_counter] = c
	defer func() {
		c.Close()
		delete(connections, cid)
		log.Printf("Connection lost: %s\n", c.RemoteAddr())
	}()
	log.Printf("New connection: %s", c.RemoteAddr())
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		if mt == websocket.TextMessage {
			log.Printf("recv[%v]: %s", cid, message)
		}
	}
}

func send_message(params []string) {
	cid, err := strconv.Atoi(params[0])
	if err != nil {
		log.Printf("Invalid cid: %s\n", params[0])
	}
	conn, ok := connections[cid]
	if ok == false {
		log.Printf("[!] No connection with cid = %v\n", cid)
		return
	}
	msg := strings.Join(params[1:], " ")
	conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func send_file(params []string) {
	cid, err := strconv.Atoi(params[0])
	if err != nil {
		log.Printf("Invalid cid: %s\n", params[0])
	}
	conn, ok := connections[cid]
	if ok == false {
		log.Printf("[!] No connection with cid = %v\n", cid)
		return
	}
	data, err := os.ReadFile(params[1])
	if err != nil {
		log.Printf("Error reading contents of file %s\n", params[1])
		return
	}
	conn.WriteMessage(websocket.TextMessage, data)
}

func broadcast_message(params []string) {
	msg := strings.Join(params, " ")
	for _, conn := range connections {
		conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}

func list_clients() {
	n := len(connections)
	if n == 0 {
		log.Println("No clients connected")
		return
	}
	log.Printf("%v clients connected:\n", n)
	for cid, conn := range connections {
		log.Printf("[%v] = %s\n", cid, conn.RemoteAddr())
	}
}

func serverInput() {
	commands := map[string]func([]string){
		"broadcast": broadcast_message,
		"send":      send_message,
		"sendf":     send_file,
		"clients":   func(params []string) { list_clients() },
	}
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		cmd_fields := strings.Fields(scanner.Text())
		command, ok := commands[cmd_fields[0]]
		if ok == false {
			log.Printf("[!] Unkown command: %s\n", cmd_fields[0])
			continue
		}
		command(cmd_fields[1:])
	}
}
