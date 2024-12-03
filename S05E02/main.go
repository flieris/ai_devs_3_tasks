package main

import (
	"ai_devs_3_tasks/helpers"
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Question struct {
	Question string `json:"question"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	questionRaw, _ := helpers.GetData(os.Getenv("S05E02_URL_Q"))
	question := Question{}
	err := json.Unmarshal(questionRaw, &question)
	if err != nil {
		log.Fatalf("Error unmarshalling question: %v", err)
	}
	log.Printf("Question: %v", question.Question)
	agent := InitAgent()

	response, err := agent.Process(question.Question)
	if err != nil {
		log.Fatalf("Error processing question: %v", err)
	}
	log.Printf("Response: %v", response)

	unfuck := make(map[string]map[string]float64)
	for _, item := range response.Answer {
		for key, value := range item {
			unfuck[key] = value
		}
	}
	message := helpers.JsonAnswer{
		Task:   "gps",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: unfuck,
	}
	log.Printf("Message: %v", message)
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %v", respCentral)
}
