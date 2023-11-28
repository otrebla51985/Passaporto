package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var inputPayload = strings.NewReader("")

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Add the client to the clients map
	clients[conn] = true
	defer delete(clients, conn)

	for {
		// Keep the WebSocket connection open to receive log messages
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			return
		}
	}
}

func LogToWebSocket(message string) {
	italianTZ, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		log.Println("Error loading Italian timezone:", err)
		return
	}

	currentTime := time.Now().In(italianTZ)

	dateTimeLayout := "02-01-2006 15:04:05"
	formattedDateTime := currentTime.Format(dateTimeLayout)
	message = formattedDateTime + " - " + message

	log.Println(message)

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("WebSocket write error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func CreateInputPayload() {
	content, err := ioutil.ReadFile("inputPayload.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Convert the file content to a string
	inputPayloadString := string(content)
	inputPayload = strings.NewReader(inputPayloadString)
}
