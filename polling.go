package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func CheckAPI(cookies string) bool {
	client := &http.Client{}
	req, err := http.NewRequest(method, passaportoEndpoint, inputPayload)

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

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	bodyString = string(body)

	// Check if the "Accesso Negato" substring is present in the XML response
	return !strings.Contains(bodyString, "Accesso Negato")
}

func PollAPI(w http.ResponseWriter, bot *tgbotapi.BotAPI, cookies string) {
	if !CheckAPI(cookies) {
		SendErrorResponse(w, "Error: Invalid or expired cookies. Please try again with valid cookies.")
		return
	}

	for {
		if pollAPIFlag {
			client := &http.Client{}
			req, err := http.NewRequest("POST", passaportoEndpoint, inputPayload)

			if err != nil {
				log.Println(err)
				return
			}
			req.Header.Add("Cookie", cookies)
			res, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Println(err)
				return
			}
			bodyString = string(body)

			response := ""
			if strings.Contains(bodyString, "\"disponibilita\">No</td>") {
				response = "NO"
				LogToWebSocket("Nessun posto libero")
			} else if strings.Contains(bodyString, "Accesso Negato") {
				response = "NO"
				LogToWebSocket("Cookies scaduti, qualcuno lo faccia ripartire pls")
			} else {
				result := GetCharactersAfterSubstring(bodyString, "data=")
				if !strings.Contains(result, "-") {
					response = "NO"
					LogToWebSocket("Nessun posto libero")
				} else {
					fmt.Println("Forse ho trovato")
					fmt.Println(bodyString)
					response = "YES"
				}
			}

			if response == "YES" {
				sendTelegramNotification(bot, bodyString)
				pollAPIFlag = false // Stop calling the API until the user presses the submit button again

				// Wait for 8 minutes before resuming the API polling
				waitTime := 8 * time.Minute
				time.Sleep(waitTime)

				// Set pollAPIFlag to true after the wait time to resume API polling
				pollAPIFlag = true
			}
		}

		time.Sleep(pollingTime)
	}
}

func SendErrorResponse(w http.ResponseWriter, message string) {
	// Send an error message back to the frontend
	errorMsg = message
	http.Error(w, message, http.StatusBadRequest)
}
