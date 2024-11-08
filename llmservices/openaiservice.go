package llmservices

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAiService struct {
	client *openai.Client
}

func NewOpenAiServcie() *OpenAiService {
	openAiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	serviceInstance := &OpenAiService{
		client: openAiClient,
	}

	return serviceInstance
}

type CompletionRequest struct {
	Messages []openai.ChatCompletionMessage
	Model    string
	Stream   bool
}

func (serviceInstance *OpenAiService) Completion(ctx context.Context, req CompletionRequest) (*openai.ChatCompletionResponse, *openai.ChatCompletionStream, error) {
	if req.Model == "" {
		req.Model = openai.GPT4
	}

	request := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   req.Stream,
	}

	if req.Stream {
		stream, err := serviceInstance.client.CreateChatCompletionStream(ctx, request)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating chat completion stream: %w", err)
		}
		return nil, stream, nil
	}

	resp, err := serviceInstance.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating chat completion: %w", err)
	}

	return &resp, nil, nil
}
