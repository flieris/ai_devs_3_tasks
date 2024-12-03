package main

import (
	"ai_devs_3_tasks/llmservices"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Agent struct {
	Client *llmservices.OpenAiService
}

type ResponseMessage struct {
	Thinking string `json:"_thinking"`
	Task     string `json:"task,omitempty"`
	Content  string `json:"content"`
}

type FormatedResponse struct {
	Thinking string                          `json:"_thinking"`
	Answer   []map[string]map[string]float64 `json:"answer"`
}

type QuerySortResponse struct {
	Thinking string `json:"_thinking"`
	Answer   string `json:"answer"`
}

func InitAgent() *Agent {
	return &Agent{
		Client: llmservices.NewOpenAiServcie(),
	}
}

func (a *Agent) sendMessage(ctx context.Context, messages *[]openai.ChatCompletionMessage) error {
	req := llmservices.CompletionRequest{
		Messages:    *messages,
		Model:       openai.GPT4oMini, // or leave empty for default GPT-4
		Stream:      false,
		Temperature: 0,
	}

	resp, _, err := a.Client.Completion(ctx, req)
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

func (a *Agent) Process(question string) (FormatedResponse, error) {
	response := FormatedResponse{}
	context := question
	query := strings.Clone(question)
	convo := []openai.ChatCompletionMessage{}
	for i := 0; i < 10; i++ {
		time.Sleep(30 * time.Second)
		plan, err := a.planExecution(context, query, &convo)
		if err != nil {
			return FormatedResponse{}, err
		}
		log.Printf("Message: %v", plan)
		if strings.Contains(plan.Task, "QUERY_API") {
			apiResponse, err := a.processApiQuery(plan)
			if err != nil {
				return FormatedResponse{}, err
			}
			query = apiResponse
			log.Printf("API Response: %v", apiResponse)
			continue
		}
		if strings.Contains(plan.Task, "QUERY_DB") {
			dbResponse, err := SendDbQuery(plan.Content)
			if err != nil {
				return FormatedResponse{}, err
			}
			query = dbResponse
			log.Printf("DB Response: %v", dbResponse)
			continue
		}
		if strings.Contains(plan.Task, "COMPLETE") {
			response, err = a.formatResponse(plan.Content)
			if err != nil {
				return FormatedResponse{}, err
			}
			break
		}
	}
	return response, nil
}

func (a *Agent) planExecution(cont string, question string, messages *[]openai.ChatCompletionMessage) (ResponseMessage, error) {
	ctx := context.Background()
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("%s\n%s\n<context>\n%s\n</context>", PlanningSystemPrompt, PlanningResponseFormat, cont),
	})
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})
	err := a.Client.SendMessage(ctx, messages, openai.GPT4o)
	if err != nil {
		return ResponseMessage{}, err
	}
	var response ResponseMessage
	lastMessage := (*messages)[len(*messages)-1]
	//	log.Printf("Last message: %v", lastMessage)
	err = json.Unmarshal([]byte(lastMessage.Content), &response)
	if err != nil {
		return ResponseMessage{}, err
	}
	return response, nil
}

func (a *Agent) processApiQuery(input ResponseMessage) (string, error) {
	apiResponse, err := SendAPIQuery(strings.Split(input.Task, " ")[1], input.Content)
	if err != nil {
		return "", err
	}
	return apiResponse, nil
}

func (a *Agent) formatResponse(input string) (FormatedResponse, error) {
	ctx := context.Background()
	message := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: CleanUpSystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		},
	}
	err := a.Client.SendMessage(ctx, &message, openai.GPT4oMini)
	if err != nil {
		return FormatedResponse{}, err
	}
	formattedResponse := FormatedResponse{}
	response := message[len(message)-1].Content
	err = json.Unmarshal([]byte(response), &formattedResponse)
	if err != nil {
		return FormatedResponse{}, err
	}
	return formattedResponse, nil
}
