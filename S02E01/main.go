package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

const (
	systemPrompt = `
<prompt_objective>
Your task is to retrive a street name of the institute of one of the polish univerisites where the person if interest works at.
</prompt_objective>
<prompt_rules>
  - The transcription is in Polish.
  - The person of interest is called "Andrzej Maj" or "profesor Maj."
  - Identify the university and the insitute where this person works.
  - based on the name of the institute, try to get the street name
  - if no street name can be deduced, return nothing
  - Output only the street name of the institute where the person works, with no additional text.
  - Deduce the correct street name from the context provided.
  - you can think aloud, just provide an definitive answer at the end.
  - some transcriptions may contain incorrect information
  - if the street name is a name of a person, return only surname
</prompt_rules>
  `
)

func reviewTranscription(ctx context.Context, transcription string, openaiClient *llmservices.OpenAiService) (string, error) {
	req := llmservices.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: transcription,
			},
		},
		Model:  openai.GPT4o, // or leave empty for default GPT-4
		Stream: false,
	}
	response, _, err := openaiClient.Completion(ctx, req)
	if err != nil {
		log.Printf("Error from OpenAI api: %v", err)
		return "", nil
	}

	// Print the response
	if response != nil {
		log.Printf("AI Response: %s\n", response.Choices[0].Message.Content)
		answer := response.Choices[0].Message.Content
		return answer, nil
	}
	return "", errors.New("No resposne from api")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	interZipPath := "przesluchania.zip"
	interPath := "przesluchania"
	err = helpers.GetZip(os.Getenv("S02E01_URL"), interZipPath)
	if err != nil {
		log.Fatalf("Error getting zip: %v", err)
	}

	err = helpers.Unzip(interZipPath, interPath)
	if err != nil {
		log.Fatalf("Error getting zip: %v", err)
	}

	interrogationFiles, err := os.ReadDir(interPath)
	if err != nil {
		log.Fatalf("Error reading files: %v", err)
	}

	client := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	var transcriptions []string
	for _, file := range interrogationFiles {
		log.Printf("Working on file: %s", file.Name())
		response, err := client.Transcribe(ctx, openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: interPath + "/" + file.Name(),
		})
		if err != nil {
			log.Fatalf("Error transcribing from openai: %v", err)
		}

		fmt.Println(response.Text)
		transcriptions = append(transcriptions, response.Text)
	}

	transcriptionsString := strings.Join(transcriptions, "\n")

	review, err := reviewTranscription(ctx, transcriptionsString, client)
	if err != nil {
		log.Fatalf("Error reviewing transcriptions: %v", err)
	}

	fmt.Println(review)

	message := helpers.JsonAnswer{
		Task:   "mp3",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: review,
	}

	resp, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", resp.Message)
}
