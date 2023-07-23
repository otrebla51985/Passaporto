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
	renderEndpoint  = "https://passaporto.onrender.com/"
	method          = "GET"
)

var bodyString = ""
var pollAPIFlag = false
var cookies string
var errorMsg string

func main() {
	bot, err := tgbotapi.NewBotAPI("5878994522:AAGAgNPCncWJxgMou5q0x6UOgkyUuD_99VA")
	if err != nil {
		log.Fatal("Error initializing Telegram bot:", err)
	}

	log.Printf("Connected to Telegram bot: %s", bot.Self.UserName)

	// Start the HTTP server to handle API requests and HTML page
	http.HandleFunc("/", handleIndexPage)
	http.HandleFunc(triggerEndpoint, handleTriggerRequest)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			panic(err)
		}
	}()

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

func handleIndexPage(w http.ResponseWriter, r *http.Request) {
	// Serve the index.html file when the root endpoint is accessed
	http.ServeFile(w, r, "index.html")
}

func checkAPI(cookies string) bool {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.passaportonline.poliziadistato.it/CittadinoAction.do?codop=resultRicercaRegistiProvincia&provincia=PD", nil)

	if err != nil {
		fmt.Println(err)
		return false
	}

	// Use the stored cookies in the request header
	req.Header.Add("Cookie", cookies)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	bodyString = string(body)

	// Check if the "Accesso Negato" substring is present in the XML response
	return !strings.Contains(bodyString, "Accesso Negato")
}

func pollAPI(w http.ResponseWriter, bot *tgbotapi.BotAPI, cookies string) {
	if !checkAPI(cookies) {
		sendErrorResponse(w, "Error: Invalid or expired cookies. Please try again with valid cookies.")
		return
	}

	for {

		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://www.passaportonline.poliziadistato.it/CittadinoAction.do?codop=resultRicercaRegistiProvincia&provincia=PD", nil)

		if err != nil {
			fmt.Println(err)
			return
		}
		req.Header.Add("Cookie", cookies)
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
		if strings.Contains(bodyString, "\"disponibilita\">No</td>") || strings.Contains(bodyString, "Accesso Negato") {
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

func sendErrorResponse(w http.ResponseWriter, message string) {
	// Send an error message back to the frontend
	errorMsg = message
	http.Error(w, message, http.StatusBadRequest)
}

func sendTelegramNotification(bot *tgbotapi.BotAPI, bodyString string) {
	log.Println("TROVATO UN POSTO - INVIO MESSAGGIO SU TELEGRAM")

	defer func() {
		if r := recover(); r != nil {
			log.Println("Panic occurred in sendTelegramNotification:", r)
		}
	}()

	log.Printf("Bot username: %s", bot.Self.UserName)
	log.Printf("Bot ID: %d", bot.Self.ID)

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

func handleTriggerRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Received trigger request - polling API and sending Telegram notification")

	// Extract the combined cookies from the URL parameters
	cookies :=
		"AGPID_FE=" + r.URL.Query().Get("AGPID_FE") + "; " +
			"AGPID=" + r.URL.Query().Get("AGPID") + "; " +
			"JSESSIONID=" + r.URL.Query().Get("JSESSIONID")

	cookies = strings.ReplaceAll(cookies, "%3B", ";")
	cookies = strings.ReplaceAll(cookies, "%26", "&")
	cookies = strings.ReplaceAll(cookies, " ", "")

	log.Println("cookies = " + cookies)

	if !checkAPI(cookies) {
		sendErrorResponse(w, "Error: Invalid or expired cookies. Please try again with valid cookies.")
		return
	}

	bot, err := tgbotapi.NewBotAPI("5878994522:AAGAgNPCncWJxgMou5q0x6UOgkyUuD_99VA")
	if err != nil {
		log.Println("Error initializing Telegram bot:", err)
		sendErrorResponse(w, "Error initializing Telegram bot. Please check the provided API token.")
		return
	}

	pollAPIFlag = true
	go pollAPI(w, bot, cookies) // Start the API polling in a separate goroutine

	// Respond to the trigger request with a success message
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("API trigger received. Polling has started with the provided cookies."))
}

func keepAlive() {
	for {
		client := &http.Client{}
		req, err := http.NewRequest(method, renderEndpoint, nil)
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
