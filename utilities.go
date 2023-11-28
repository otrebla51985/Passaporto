package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var inputPayload = ""

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
	content, err := os.ReadFile("payloadInput.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Convert the file content to a string
	inputPayloadString := string(content)
	inputPayload = inputPayloadString
}

func KeepAlive() {
	for {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://passaporto.onrender.com/", nil)
		if err != nil {
			log.Println("Error creating request to Render instance:", err)
			time.Sleep(pollingTime) // Retry after the pollingTime
			continue
		}

		// Make the API call to the Render instance
		res, err := client.Do(req)
		if err != nil {
			log.Println("Error calling Render instance:", err)
		} else {
			defer res.Body.Close()
			log.Println("API call to Render instance successful")
		}

		time.Sleep(pollingTime)
	}
}
