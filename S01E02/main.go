package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/openaiservice"
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	apiUrl := os.Getenv("VERIFY_URL")
	initData := helpers.JsonMessage{
		MsgID: 0,
		Text:  "READY",
	}

	apiResponse, err := helpers.SendJson(apiUrl, initData)

	msgId := apiResponse.MsgID
	question := apiResponse.Text

	openAI := openaiservice.New()

	req := openaiservice.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Przyjmij następujące założenia: Zawsze odpowiadaj po angielsku, stolica Polski jest Kraków (w takiej pisowni), znana liczba z ksiązki Autosptopem przez Galatyke to 69, aktualny rok to 1999. Odpowiadaj zwięźle.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
		Model:  openai.GPT4, // or leave empty for default GPT-4
		Stream: false,
	}
	ctx := context.Background()
	response, _, err := openAI.Completion(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response
	if response != nil {
		log.Printf("AI Response: %s\n", response.Choices[0].Message.Content)
		answer := response.Choices[0].Message.Content
		answerJson := helpers.JsonMessage{
			MsgID: msgId,
			Text:  answer,
		}
		key, err := helpers.SendJson(apiUrl, answerJson)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Response from xyz: %s", key.Text)
	}
}
