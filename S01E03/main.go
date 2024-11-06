package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/openaiservice"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type RootJson struct {
	ApiKey      string     `json:"apikey"`
	Description string     `json:"description"`
	Copyright   string     `json:"copyright"`
	TestData    []TestData `json:"test-data"`
}

type TestData struct {
	Question string            `json:"question"`
	Answer   int               `json:"answer"`
	Test     map[string]string `json:"test,omitempty"`
}

func sumExpression(expr string) (int, error) {
	expr = strings.TrimSpace(expr)

	parts := strings.Split(expr, "+")

	left, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, err
	}
	right, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, err
	}
	return left + right, nil
}

func sendQuestion(question string) (string, error) {
	openAI := openaiservice.New()

	req := openaiservice.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Answer the question in the most consice way.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
		Model:  openai.GPT4o, // or leave empty for default GPT-4
		Stream: false,
	}
	ctx := context.Background()
	response, _, err := openAI.Completion(ctx, req)
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

	calibDataRaw, err := helpers.GetData(os.Getenv("S01E03_URL_1"))

	var calibDataJson RootJson
	if err := json.NewDecoder(bytes.NewReader(calibDataRaw)).Decode(&calibDataJson); err != nil {
		log.Fatalf("Error decoding json data: %v", err)
	}
	calibDataJson.ApiKey = os.Getenv("CENTRAL_API_KEY")

	for testElement := range calibDataJson.TestData {
		elemSum, err := sumExpression(calibDataJson.TestData[testElement].Question)
		if err != nil {
			log.Printf("Something has gone wrong: %s", err)
			continue
		}
		if elemSum != calibDataJson.TestData[testElement].Answer {
			log.Printf("The following operation %s, has incorrect answer (wrong: %d, correct %d). Correcting...", calibDataJson.TestData[testElement].Question, calibDataJson.TestData[testElement].Answer, elemSum)
			calibDataJson.TestData[testElement].Answer = elemSum
			log.Printf("Current answer: %d", calibDataJson.TestData[testElement].Answer)
		}
		if calibDataJson.TestData[testElement].Test != nil {
			log.Println("Test field is present:", calibDataJson.TestData[testElement].Test)
			answer, err := sendQuestion(calibDataJson.TestData[testElement].Test["q"])
			if err != nil {
				log.Printf("Something has gone wrong: %s", err)
				continue
			}
			calibDataJson.TestData[testElement].Test["a"] = answer
		}
	}

	message := helpers.JsonAnswer{
		Task:   "JSON",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: calibDataJson,
	}

	resp, err := helpers.SendAnswer(os.Getenv("S01E03_URL_2"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", resp.Message)
}
