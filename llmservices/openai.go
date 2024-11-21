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
	Messages    []openai.ChatCompletionMessage
	Model       string
	Stream      bool
	Temperature float32
}

type EmbeddingRequest struct {
	Input string
	Model string
}

func (serviceInstance *OpenAiService) Completion(ctx context.Context, req CompletionRequest) (*openai.ChatCompletionResponse, *openai.ChatCompletionStream, error) {
	if req.Model == "" {
		req.Model = openai.GPT4
	}

	request := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
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

func (serviceInstance *OpenAiService) Transcribe(ctx context.Context, req openai.AudioRequest) (*openai.AudioResponse, error) {
	if req.Model == "" {
		req.Model = openai.Whisper1
	}

	resp, err := serviceInstance.client.CreateTranscription(ctx, req)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (serviceInstance *OpenAiService) CreateImage(ctx context.Context, req openai.ImageRequest) (*openai.ImageResponse, error) {
	image, err := serviceInstance.client.CreateImage(ctx, req)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (serviceInstance *OpenAiService) CreateEmbedding(ctx context.Context, req openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	if req.Model == "" {
		req.Model = openai.SmallEmbedding3
	}

	resp, err := serviceInstance.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
