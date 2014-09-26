package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/NeilVallon/blstr"

	"github.com/gorilla/websocket"
)

var count = 0

func chatws(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Create chat user info
	ch := make(chan []byte)
	id := count
	count++

	// Subscribe to chat
	if err := hub.Subscribe(id, ch); err != nil {
		log.Println(err)
		return
	}
	defer hub.Unsubscribe(id)

	// Create channel to close writer
	done := make(chan struct{})
	defer close(done)

	go writer(ch, done, conn)
	reader(id, conn)
}

func writer(ch <-chan []byte, done chan struct{}, conn *websocket.Conn) {
	for {
		select {
		case msg := <-ch:
			conn.WriteMessage(1, msg)
		case <-done:
			return
		}
	}
}

func reader(id int, conn *websocket.Conn) {
	for {
		t, msg, err := conn.ReadMessage()
		if err != nil || t != 1 {
			return
		}
		hub.Flood(id, msg)
	}
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/chatws", chatws)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, page)
}

var (
	hub = blstr.New()

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	page = `<!DOCTYPE html>
<html>
<head>
	<title>blstr - Chat Example</title>
	<script type="text/javascript">
		function addLogItem(li) {
			var ul = document.getElementById("log");
			ul.appendChild(li);
		}

		function addChatItem(msg) {
			var li = document.createElement("li");
			li.appendChild(document.createTextNode(msg));

			addLogItem(li);
		};

		function addServerItem(msg, color) {
			var li = document.createElement("li");
			li.appendChild(document.createTextNode(msg));
			li.style.color = color;

			addLogItem(li);
		};

		var websocket = new WebSocket("ws://127.0.0.1:8080/chatws");
		websocket.onopen = function(evt) {
			addServerItem("Connection opened.", "#5a5");
		};
		websocket.onclose = function(evt) {
			addServerItem("Connection closed.", "#aa5");
		};
		websocket.onerror = function(evt) {
			addServerItem("Connection error.", "#a55");
		};

		websocket.onmessage = function(evt) {
			addChatItem(evt.data);
		};

		function sendMessage() {
			var txtBox = document.getElementById("message");
			var msg = txtBox.value;

			websocket.send(msg);
			addChatItem(msg);

			txtBox.value = "";
		}
	</script>
</head>
<body>
	<input type="text" id="message" />
	<button type="button" onClick="sendMessage()" />send</button>

	<h4>Chat Log</h4>
	<ul id="log"></ul>
</body>
</html>`
)
