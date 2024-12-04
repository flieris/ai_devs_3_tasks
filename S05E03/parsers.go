package main

import (
	"ai_devs_3_tasks/llmservices"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sashabaranov/go-openai"
)

func ParseWebsiteContent(url string) ([]string, error) {
	output := []string{}
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	doc.Find("body").Each(func(index int, item *goquery.Selection) {
		output = append(output, item.Text())
	})
	return output, nil
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

func HandleWebsiteChallange(challange RafalChallange, url string) map[string]string {
	openAI := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	// don't really care about error handling here
	content, _ := ParseWebsiteContent(url)
	msg := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s\n%s\n<document>\n%s\n</document>", SystemPrompt, ResponseFormat, strings.Join(content, "")),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: strings.Join(challange.Data, "\n"),
		},
	}
	err := sendMessage(ctx, openAI, &msg)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	lastMessage := msg[len(msg)-1]
	response := make(map[string]string)
	err = json.Unmarshal([]byte(lastMessage.Content), &response)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	return response
}

func HandleNormalChallange(challange RafalChallange) map[string]string {
	openAI := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	msg := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s\n%s", SystemPrompt, ResponseFormat),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: strings.Join(challange.Data, "\n"),
		},
	}
	err := sendMessage(ctx, openAI, &msg)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	lastMessage := msg[len(msg)-1]
	response := make(map[string]string)
	err = json.Unmarshal([]byte(lastMessage.Content), &response)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	return response
}
