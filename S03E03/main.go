package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

const (
	systemPrompt = `
<objective>You are a data retrival service that works with an sql database to get specific data. You task is to discover the database structure and retrive information from it</objective>
<rules>
  - First message you will get is <start>
  - If you think the task is done respond with <end>
  - We need to get the id of datacenter, which are managed by managers who are currently  on leave
  - Only respond with either sql statements, or end result
  - you can only perform operations like "table_name" refers to the table in the database: 
    - select
    - show tables, 
    - desc table_name, 
    - show create table table_name
</rules>
  `
)

type DbQuery struct {
	Task   string `json:"task"`
	ApiKey string `json:"apikey"`
	Query  string `json:"query"`
}

type DbResponse struct {
	Reply []interface{} `json:"reply"`
	Error string        `json:"error"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func sendMessage(ctx context.Context, client *llmservices.OpenAiService, messages *[]openai.ChatCompletionMessage) error {
	req := llmservices.CompletionRequest{
		Messages:    *messages,
		Model:       openai.GPT4oMini, // or leave empty for default GPT-4
		Stream:      false,
		Temperature: 0,
	}

	resp, _, err := client.Completion(ctx, req)
	if err != nil {
		return err
	}

	content := resp.Choices[0].Message.Content
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	})
	return nil
}

func sendQuery(query string) (string, error) {
	queryToSend := DbQuery{
		Task:   "database",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Query:  query,
	}
	var responseObj DbResponse
	err := helpers.SendRequest(os.Getenv("S03E03_URL"), queryToSend, &responseObj)
	if err != nil {
		return "", err
	}

	responseJson, err := json.Marshal(responseObj)
	if err != nil {
		return "", err
	}

	return string(responseJson), nil
}

func getDCFromString(jsonString string) ([]int, error) {
	type DC struct {
		DcID string `json:"dc_id"`
	}
	type Response struct {
		Reply []DC   `json:"reply"`
		Error string `json:"error"`
	}

	// Unmarshal JSON string into struct
	var response Response
	err := json.Unmarshal([]byte(jsonString), &response)
	if err != nil {
		return nil, err
	}

	// Convert the dc_id values to integers and store in a slice
	var result []int
	for _, item := range response.Reply {
		id, err := strconv.Atoi(item.DcID) // Convert string to int
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}

	return result, nil
}

func main() {
	openAi := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	var messages []openai.ChatCompletionMessage
	var dcIds []int

	systemMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	}
	messages = append(messages, systemMsg)
	startMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "<start>",
	}
	messages = append(messages, startMsg)
	for {
		err := sendMessage(ctx, openAi, &messages)
		if err != nil {
			log.Fatalf("Error: %v", err)
			break
		}
		lastMessage := messages[len(messages)-1]
		log.Printf("Last Message: %v", lastMessage.Content)
		if strings.Contains(lastMessage.Content, "end") {
			dcIds, err = getDCFromString(messages[len(messages)-2].Content)
			if err != nil {
				log.Fatalf("Error: %v", err)
				break
			}
			break
		}
		dbResponse, err := sendQuery(lastMessage.Content)
		if err != nil {
			log.Fatalf("Error getting response from db: %v", err)
		}
		log.Printf("Response from db: %s", dbResponse)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: dbResponse,
		})
	}

	message := helpers.JsonAnswer{
		Task:   "database",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: dcIds,
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
