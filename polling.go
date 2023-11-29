package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// returns an empty string if the call worked
// otherwise it returns the csrf token to try a new call
func CheckAPI(cookies string, csrfToken string) string {
	client := &http.Client{}
	req, err := http.NewRequest(method, passaportoEndpoint, strings.NewReader(inputPayload))

	if err != nil {
		fmt.Println(err)
		return ""
	}

	contentLength := len(inputPayload)
	req.ContentLength = int64(contentLength)
	contentLengthStr := strconv.Itoa(contentLength)

	// Use the stored cookies in the request header
	req.Header.Add("Cookie", cookies)
	req.Header.Add("X-CSRF-TOKEN", csrfToken)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/120.0")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Content-Length", contentLengthStr)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	bodyString = string(body)

	log.Printf("bodyString nuovo = " + bodyString)

	if strings.Contains(bodyString, "_csrf") {
		//return the csrf token:
		return GetCharactersAfterSubstring(bodyString, "_csrf\" content=\"", 36)
	}

	return ""
}

func PollAPI(w http.ResponseWriter, bot *tgbotapi.BotAPI, cookies string) {
	thisCsrfToken := CheckAPI(cookies, "")
	CheckAPI(cookies, thisCsrfToken)

	for {
		if pollAPIFlag {
			client := &http.Client{}
			req, err := http.NewRequest("POST", passaportoEndpoint, strings.NewReader(inputPayload))

			if err != nil {
				log.Println(err)
				return
			}

			contentLength := len(inputPayload)
			req.ContentLength = int64(contentLength)
			contentLengthStr := strconv.Itoa(contentLength)

			// Use the stored cookies in the request header
			req.Header.Add("Cookie", cookies)
			req.Header.Add("X-CSRF-TOKEN", thisCsrfToken)
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/120.0")
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Content-Length", contentLengthStr)

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
			if strings.Contains(bodyString, "dataPrimaDisponibilitaResidenti\":null") {
				response = "NO"
				LogToWebSocket("Nessun posto libero")
			} else if strings.Contains(bodyString, "_csrf") {
				response = "NO"
				LogToWebSocket("Cookies scaduti, qualcuno lo faccia ripartire pls")
			} else {
				result := GetCharactersAfterSubstring(bodyString, "dataPrimaDisponibilitaResidenti\":\"", 10)
				if strings.Contains(result, "null") {
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

				// Wait 3 minutes before resuming the API polling
				waitTime := 3 * time.Minute
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
