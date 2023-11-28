package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	pollingTime        = 16 * time.Second
	triggerEndpoint    = "/trigger"
	passaportoEndpoint = "https://passaportonline.poliziadistato.it/cittadino/a/rc/v1/appuntamento/elenca-sede-prima-disponibilita"
	method             = "POST"
)

var bodyString = ""
var pollAPIFlag = false
var errorMsg string

var clients = make(map[*websocket.Conn]bool)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	bot, err := tgbotapi.NewBotAPI("5878994522:AAGAgNPCncWJxgMou5q0x6UOgkyUuD_99VA")
	if err != nil {
		log.Fatal("Error initializing Telegram bot:", err)
	}

	log.Printf("Connected to Telegram bot: %s", bot.Self.UserName)

	// Start the HTTP server to handle API requests and HTML page
	http.HandleFunc("/", HandleIndexPage)
	http.HandleFunc(triggerEndpoint, HandleTriggerRequest)
	http.HandleFunc("/ws", HandleWebSocket) // WebSocket endpoint

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			panic(err)
		}
	}()

	go KeepAlive()

	// Block the main goroutine to keep the server running indefinitely
	// and recover from any panics that may occur
	defer func() {
		if r := recover(); r != nil {
			log.Println("Panic occurred:", r)
			log.Println("Server is restarting...")
		}
	}()

	// Block the main goroutine to keep the server running indefinitely
	select {}
}

func HandleIndexPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Serve the index.html file when the root endpoint is accessed
	http.ServeFile(w, r, "index.html")
}
