package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type RobotDescription struct {
	Description string `json:"description"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	resp, err := helpers.GetData(os.Getenv("S02E03_URL"))
	if err != nil {
		log.Fatalf("Error getting message from api: %v", err)
	}
	var robot RobotDescription

	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&robot); err != nil {
		log.Fatalf("Error decoding json object: %v", err)
	}

	log.Println(robot.Description)

	openaiClient := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	req := openai.ImageRequest{
		Prompt:         robot.Description,
		Model:          openai.CreateImageModelDallE3,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respUrl, err := openaiClient.CreateImage(ctx, req)
	if err != nil {
		log.Fatalf("Error generating image: %v", err)
	}

	log.Println(respUrl.Data[0].URL)

	message := helpers.JsonAnswer{
		Task:   "robotid",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: respUrl.Data[0].URL,
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
