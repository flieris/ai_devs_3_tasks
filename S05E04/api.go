package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Agent struct {
	Client *llmservices.OpenAiService
}

type ResponseMessage struct {
	Tool     string `json:"tool,omitempty"`
	Thinking string `json:"_thinking"`
	Answer   string `json:"answer,omitempty"`
	Query    string `json:"query,omitempty"`
}

func InitAgent() *Agent {
	return &Agent{
		Client: llmservices.NewOpenAiServcie(),
	}
}

func (a *Agent) Process(question string) (string, error) {
	response := ""
	context := question
	query := strings.Clone(question)
	convo := []openai.ChatCompletionMessage{}
	for i := 0; i < 10; i++ {
		plan, err := a.planExecution(context, query, &convo)
		if err != nil {
			return "", err
		}
		response = plan.Answer
		log.Printf("Response: %s", response)
		log.Printf("Tool: %s", plan.Tool)
		switch plan.Tool {
		case "save_data":
			// save data
			response, err = a.processQueryDatabase(plan.Answer)
			log.Printf("Specified query: %s", response)
			if err != nil {
				return "", err
			}
			query = response
		case "query_database":
			// query database
			response, err = a.processQueryDatabase(plan.Answer)
			log.Printf("Specified query: %s", response)
			if err != nil {
				return "", err
			}
			query = response
		case "get_transcription":
			// get audio transcription
			response, err = a.processAudio(plan.Answer)
			if err != nil {
				return "", err
			}
			log.Printf("Audio transcription response: %s", response)
		case "image_processing":
			// process image
			response, err = a.processImage(plan.Answer)
			if err != nil {
				return "", err
			}
			log.Printf("Image processing response: %s", response)
		case "final_answer":
			return response, nil
		}
		break
	}
	return response, nil
}

func (a *Agent) planExecution(cont string, question string, messages *[]openai.ChatCompletionMessage) (ResponseMessage, error) {
	ctx := context.Background()
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("%s\n<context>\n%s\n</context>", PlanningPrompt, cont),
	})
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})
	err := a.Client.SendMessage(ctx, messages, openai.GPT4oMini)
	if err != nil {
		return ResponseMessage{}, err
	}
	var response ResponseMessage
	lastMessage := (*messages)[len(*messages)-1]
	err = json.Unmarshal([]byte(lastMessage.Content), &response)
	if err != nil {
		return ResponseMessage{}, err
	}
	return response, nil
}

func (a *Agent) processQueryDatabase(query string) (string, error) {
	ctx := context.Background()
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s\n", QueryDatabasePrompt),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: query,
		},
	}
	err := a.Client.SendMessage(ctx, &messages, openai.GPT4oMini)
	if err != nil {
		return "", err
	}
	var response ResponseMessage
	lastMessage := (messages)[len(messages)-1]
	err = json.Unmarshal([]byte(lastMessage.Content), &response)
	if err != nil {
		return "", err
	}
	data, err := QueryDatabase(response.Query)
	if err != nil {
		return "", err
	}
	return data, nil
}

func (a *Agent) processImage(query string) (string, error) {
	ctx := context.Background()
	urlRegex := regexp.MustCompile(`(https?://[^\s]+)`)
	response := ""
	match := urlRegex.FindString(query)

	if match != "" {
		data, err := helpers.GetData(match)
		if err != nil {
			return "", err
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		message := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("%s\n", ImageProcessingPrompt),
			},
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: fmt.Sprintf("Please analyze this image: %s", path.Base(match)),
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: fmt.Sprintf("data:image/png;base64,%s", encoded),
						},
					},
				},
			},
		}
		log.Printf("Image: %s", path.Base(match))
		err = a.Client.SendMessage(ctx, &message, openai.GPT4o)
		if err != nil {
			log.Printf("Error: %v", err)
			return "", err
		}
		var output ResponseMessage
		err = json.Unmarshal([]byte(message[len(message)-1].Content), &output)
		if err != nil {
			log.Printf("Error: %v", err)
			return "", err
		}
		log.Printf("Output: %v", output)
		return output.Answer, nil
	} else {
		return "No URL found", nil
	}
	return response, nil
}

func (a *Agent) processAudio(query string) (string, error) {
	ctx := context.Background()
	urlRegex := regexp.MustCompile(`(https?://[^\s]+)`)
	match := urlRegex.FindString(query)
	if match != "" {
		out, err := os.Create(path.Base(match))
		if err != nil {
			return "", err
		}
		defer out.Close()
		data, err := helpers.GetData(match)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(out, bytes.NewReader(data))
		if err != nil {
			return "", err
		}
		log.Printf("Audio: %s", path.Base(match))
		response, err := a.Client.Transcribe(ctx, openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: out.Name(),
		})
		if err != nil {
			return "", err
		}
		return response.Text, nil
	} else {
		return "No URL found", nil
	}
}
