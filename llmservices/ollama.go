package llmservices

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type OllamaChatCompletion struct {
	Model    string                        `json:"model"`
	Messages []OllamaChatCompletionMessage `json:"messages"`
	Stream   bool                          `josn:"stream,omitempty"`
}

type OllamaChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatCompletionResponse struct {
	Model              string                      `json:"model"`
	CreatedAt          string                      `json:"created_at"`
	Message            OllamaChatCompletionMessage `json:"message"`
	DoneReason         string                      `json:"done_reason"`
	Done               bool                        `json:"done"`
	TotalDuration      int                         `json:"total_duration"`
	LoadDuration       int                         `json:"load_duration"`
	PromptEvalCount    int                         `json:"prompt_eval_count"`
	PromptEvalDuration int                         `json:"prompt_eval_duration"`
	EvalCount          int                         `json:"eval_count"`
	EvalDuration       int                         `json:"eval_duration"`
}

func SendChatCompletion(chatCompletionRequest OllamaChatCompletion, ollamaUrl string) (response OllamaChatCompletionResponse, err error) {
	chatCOmpletionRequestJson, err := json.Marshal(chatCompletionRequest)
	if err != nil {
		return
	}
	resp, err := http.Post(ollamaUrl, "application/json", bytes.NewBuffer(chatCOmpletionRequestJson))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return
	}
	return
}
