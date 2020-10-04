package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var ipu = make(map[net.Conn]string)

func main() {
	port := flag.Int("port", 3333, "Port to accept connections on.")
	host := flag.String("host", "192.168.40.13", "Host or IP to bind to")
	flag.Parse()

	l, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Listening to connections at '"+*host+"' on port", strconv.Itoa(*port))
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Panicln(err)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	var username string
	defer log.Println(username + " has left")
	defer func() {
		for c := range ipu {
			c.Write([]byte(username + " HAS LEFT\n"))
		}
	}()
	defer postMessage("SERVER", username+" HAS LEFT")
	defer delete(ipu, conn)
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			return
		}

		asdf := string(buf[:size])
		if len(asdf) > 6 && asdf[:5] == "/join" {
			subs := strings.Split(asdf, ":")
			ipu[conn] = strings.Trim(subs[1], "\n")
			log.Println(ipu[conn] + " has joined")
			for c := range ipu {
				c.Write([]byte(ipu[conn] + " HAS JOINED\n"))
			}
			username = ipu[conn]
			postMessage("SERVER", ipu[conn]+" HAS JOINED")
		} else {
			msg := ipu[conn] + ": " + asdf
			msgBuf := []byte(msg)
			newSize := len(msgBuf)
			data := msgBuf[:newSize]
			for c := range ipu {
				c.Write(data)
			}
			postMessage(ipu[conn], asdf)
			log.Println(msg)
		}
	}
}

func postMessage(author string, body string) {
	msg := message{
		Author: author,
		Body:   body,
	}

	msgj, _ := json.Marshal(msg)
	resp, err := http.Post("http://159.89.8.129:2814/message", "application/json", bytes.NewBuffer(msgj))
	if err != nil {
		log.Println("UNABLE TO SAVE MESSAGE")
	}
	defer resp.Body.Close()
}

type message struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}
