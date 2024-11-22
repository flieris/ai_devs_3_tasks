package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

// TODO: finish this, 1st get list of people and places from openai and then send it as a document
// maybe try to get just the mapping of api responses and then work from there? like a vector db or something like that

const (
	systemPrompt1 = `
<objective> Get a list of people and places from the imputted text </objective>
<rules>
- Get the list of first names only and town from the text
- the names should be in the Nominative Case, without any polish special characters
- output in items in UPPERCASE
- output like this:
{
  "places": ["place1", "place2"],
  "people": ["person1", "person2"]
</rules>
  `
	systemPrompt2 = `
<objective>You are a person locator. Based on 2 apis: places, people. You need to find a specified person </objective>
<rules>
  - First message you will get is <start>
  - If you think the task is done respond with name of a town and <end> marker
  - We need to find a TOWN where BARBARA is located at
  - BARBARA MUST be the only person in that town/city
  - If your query to the PLACES api finds BARBARA, respond with the name of the town you asked about and <end> marker
  - use polish names for towns/cities
  - Your output can only be:
API: PLACES OR PEOPLE
QUERY: PERSON/PLACE
- You can ask either places api or people api
- You are also given an input document you can base your queries on
- If you get this response: '{"code":0,"message":"[**RESTRICTED DATA**]"', you **MUST**:
    - Record this query as RESTRICTED
    - Check against the list of RESTRICTED queries before making any new query
    - **DO NOT** repeat any query marked as RESTRICTED
- If you query about RAFAŁ as input, convert the name to RAFAL before querying
- put your reasoning in <thinking></thinking> field
</rules>
<examples>
  - If you queried 'API: PEOPLE, QUERY: BARBARA' and got '{"code":0,"message":"[**RESTRICTED DATA**]"}', you must not query 'API: PEOPLE, QUERY: BARBARA' again.
</examples>
<thinking>
</thinking>
  `
)

type QueryRequest struct {
	ApiKey string `json:"apikey"`
	Query  string `json:"query"`
}

type QueryResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func getFormattedText(ctx context.Context, client *llmservices.OpenAiService, message string) (string, error) {
	req := llmservices.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("%s\n", systemPrompt1),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: message,
			},
		},
		Model:  openai.GPT4o, // or leave empty for default GPT-4
		Stream: false,
	}
	response, _, err := client.Completion(ctx, req)
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

func sendQuery(api string, query string) (string, error) {
	queryToSend := QueryRequest{
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Query:  query,
	}
	var responseObj QueryResponse
	err := helpers.SendRequest(api, queryToSend, &responseObj)
	if err != nil {
		return "", err
	}

	responseJson, err := json.Marshal(responseObj)
	if err != nil {
		return "", err
	}

	return string(responseJson), nil
}

func parseInput(input string) map[string]string {
	lines := strings.Split(input, "\n")
	inputMap := make(map[string]string)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "API:") {
			inputMap["api"] = strings.TrimSpace(strings.TrimPrefix(line, "API:"))
		} else if strings.HasPrefix(line, "QUERY:") {
			inputMap["query"] = strings.TrimSpace(strings.TrimPrefix(line, "QUERY:"))
		}
	}
	return inputMap
}

func main() {
	barbsData, err := helpers.GetData(os.Getenv("S03E04_HINT_URL"))
	if err != nil {
		log.Fatalf("Error getting questions: %v", err)
	}

	openAi := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	formattedData, err := getFormattedText(ctx, openAi, string(barbsData))
	if err != nil {
		log.Fatalf("Error getting formatted text: %v", err)
	}
	var messages []openai.ChatCompletionMessage
	var dcIds []int
	log.Printf("%s\n<document>\n%s\n</document>", systemPrompt2, formattedData)
	systemMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("%s\n<document>\n%s\n</document>", systemPrompt2, formattedData),
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
			break
		}
		input := parseInput(lastMessage.Content)
		var apiUrl string
		switch input["api"] {
		case "PLACES":
			apiUrl = os.Getenv("PLACES_URL")
		case "PEOPLE":
			apiUrl = os.Getenv("PEOPLE_URL")
		default:
			log.Printf("Unknown key: %s\n", input["api"])

		}
		log.Printf("asking api %s, about %s\n", apiUrl, input["query"])
		queryResp, err := sendQuery(apiUrl, input["query"])
		if err != nil {
			log.Fatalf("Error getting response from db: %v", err)
			break
		}
		log.Printf("Response from db: %s", queryResp)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: queryResp,
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
