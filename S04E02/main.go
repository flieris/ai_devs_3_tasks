package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	openAi := llmservices.NewOpenAiServcie()
	ctx := context.Background()

	labDataZipPath := "lab_data.zip"
	labDataPath := "lab_data"
	err := helpers.GetZip(os.Getenv("S04E02_URL"), labDataZipPath)
	if err != nil {
		log.Fatalf("Error getting zip: %v", err)
	}

	err = helpers.Unzip(labDataZipPath, labDataPath)
	if err != nil {
		log.Fatalf("Error getting zip: %v", err)
	}

	fileFd, err := os.Open(labDataPath + "/verify.txt")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer fileFd.Close()
	output, err := io.ReadAll(fileFd)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	valuesToVerify := make(map[string]string)
	for _, value := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		tmp := strings.Split(value, "=")
		valuesToVerify[tmp[0]] = tmp[1]
	}
	log.Println(valuesToVerify)
	var correct []string
	for id, verify := range valuesToVerify {
		msg := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Sklasyfikuj, czy podane próbki sa poprawne",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: verify,
			},
		}

		openAi.SendMessage(ctx, &msg, os.Getenv("FINE_TUNE_MODEL"))
		response := msg[len(msg)-1].Content
		if response == "correct" {
			correct = append(correct, id)
		}
	}

	message := helpers.JsonAnswer{
		Task:   "research",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: correct,
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
