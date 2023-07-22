package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	pollingTime     = 30 * time.Second
	triggerEndpoint = "/trigger"
)

var bodyString = ""
var pollAPIFlag = false

func main() {
	bot, err := tgbotapi.NewBotAPI("5878994522:AAGAgNPCncWJxgMou5q0x6UOgkyUuD_99VA")
	if err != nil {
		log.Fatal("Error initializing Telegram bot:", err)
	}

	log.Printf("Connected to Telegram bot: %s", bot.Self.UserName)

	// Start the HTTP server to handle API requests
	http.HandleFunc(triggerEndpoint, func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received trigger API request")
		pollAPIFlag = true
		go handleTriggerRequest(bot)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("API trigger received"))
	})

	go startHTTPServer()

	// Block the main goroutine to keep the server running indefinitely
	select {}
}

func startHTTPServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func pollAPI(bot *tgbotapi.BotAPI) {
	for {

		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://www.passaportonline.poliziadistato.it/CittadinoAction.do?codop=resultRicercaRegistiProvincia&provincia=PD", nil)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Cookie", "AGPID_FE=AtngJSgKxwqA4yYbx6RLGA$$; AGPID=Ae14AJoLxwob7dw5PuiSXQ$$; JSESSIONID=clKrpYtuvJiQQE4yYTjdvi5W; AGPID=ANArJpoLxwrq5BstH+JHTA$$; AGPID_FE=AkiAQygKxwrZRAwRYn9wfQ$$; JSESSIONID=clKrpYtuvJiQQE4yYTjdvi5W")
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		bodyString = string(body)

		response := ""
		if strings.Contains(bodyString, "\"disponibilita\">No</td>") {
			response = "NO"
			log.Println("Ancora niente")
		} else {
			fmt.Println("Forse ho trovato")
			fmt.Println(bodyString)
			response = "YES"
		}

		if response == "YES" {
			sendTelegramNotification(bot, bodyString)
			pollAPIFlag = false
			break // Exit the loop when "YES" response is received
		}

		time.Sleep(pollingTime)
	}
}

func sendTelegramNotification(bot *tgbotapi.BotAPI, bodyString string) {
	log.Println("TROVATO UN POSTO - INVIO MESSAGGIO SU TELEGRAM")

	chatID := int64(-974313836) //YOUR_TELEGRAM_CHAT_ID
	//mio = 112845421
	//gruppo = -974313836

	//msg := tgbotapi.NewMessage(chatID, "API responded with 'YES'! "+bodyString)
	//_, err := bot.Send(msg)
	//if err != nil {
	//	log.Println("Error sending Telegram notification:", err)
	//}
	//os.Exit(3)

	// Create a temporary file to store the API response
	fileName := "api_response.xml"
	file, err := os.Create(fileName)
	if err != nil {
		log.Println("Error creating temporary file:", err)
		return
	}
	defer os.Remove(fileName)
	defer file.Close()

	result := getCharactersAfterSubstring(bodyString, "data=")
	fmt.Println("Characters after the substring:", result)

	// Write the API response to the file
	_, err = file.WriteString(bodyString)
	if err != nil {
		log.Println("Error writing to temporary file:", err)
		return
	}

	// Read the temporary file and get its content
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("Error reading temporary file:", err)
		return
	}

	// Create a tgbotapi.FileBytes instance with the file content
	fileBytes := tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: data,
	}

	// Create the Telegram document
	msg := tgbotapi.NewDocumentUpload(chatID, fileBytes)
	msg.Caption = "C'è posto" + "    \n\ndata = " + result
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Error sending Telegram document:", err)
	}
	os.Exit(3)
}

func getCharactersAfterSubstring(inputString, substring string) string {
	index := strings.Index(inputString, substring)

	if index == -1 {
		// Substring not found, return an empty string or handle the error accordingly.
		return ""
	}

	endPosition := index + len(substring) + 10

	// Check if the end position is within the bounds of the inputString.
	if endPosition > len(inputString) {
		endPosition = len(inputString)
	}

	// Extract the characters after the substring up to the 10th character.
	return inputString[index+len(substring) : endPosition]
}

func handleTriggerRequest(bot *tgbotapi.BotAPI) {
	log.Println("Received trigger request - polling API and sending Telegram notification")
	pollAPIFlag = true
	go pollAPI(bot) // Start the API polling in a separate goroutine
}
