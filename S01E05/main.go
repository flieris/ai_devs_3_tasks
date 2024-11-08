package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	system_prompt = `
- Zamień imię i nazwisko na 'CENZURA'.
- Zamień adres (ulicę i numer domu) na 'CENZURA'. 
- Zamień nazwe miasta na 'CENZURA'
- Zamień wiek osoby na 'CENZURA'.
- Zachowaj oryginalną strukturę tekstu, zmieniając jedynie wskazane dane.
- Nie dodawaj nic do tekstu
- nie zmieniaj pozostałego tekstu

Przykłady:  

Informacje o podejrzanym: Marek Jankowski. Mieszka w Białymstoku na ulicy Lipowej 9. Wiek: 26 lat.
-> 
Informacje o podejrzanym: CENZURA. Mieszka w CENZURA na ulicy CENZURA. Wiek: CENZURA lat. 

Dane personalne podejrzanego: Wojciech Górski. Przebywa w Lublinie, ul. Akacjowa 7. Wiek: 27 lat.
->
Dane personalne podejrzanego: CENZURA. Przebywa w CENZURA, ul. CENZURA. Wiek: CENZURA lat.

Osoba podejrzana to Andrzej Mazur. Adres: Gdańsk, ul. Długa 8. Wiek: 29 lat.
->
Osoba podejrzana to CENZURA. Adres: CENZURA, ul. CENZURA. Wiek: CENZURA lat.
  `
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dataToAnon, err := helpers.GetData(os.Getenv("S01E05_URL"))
	if err != nil {
		log.Fatalf("Error getting data: %v", err)
	}

	fmt.Println(string(dataToAnon))

	req := llmservices.OllamaChatCompletion{
		Messages: []llmservices.OllamaChatCompletionMessage{
			{
				Role:    "system",
				Content: system_prompt,
			},
			{
				Role:    "user",
				Content: string(dataToAnon),
			},
		},
		Model:  "gemma2:2b",
		Stream: false,
	}
	reqUrl := os.Getenv("OLLAMA_URL") + "api/chat"
	response, err := llmservices.SendChatCompletion(req, reqUrl)
	if err != nil {
		log.Fatalf("Error getting data from ollama: %v", err)
	}
	fmt.Println(strings.TrimSuffix(response.Message.Content, "\n"))

	message := helpers.JsonAnswer{
		Task:   "CENZURA",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: strings.TrimSuffix(response.Message.Content, "\n"),
	}
	resp, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", resp.Message)
}
