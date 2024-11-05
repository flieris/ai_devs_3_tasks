package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/openaiservice"
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/net/html"
)

func findElementByID(n *html.Node, id string) string {
	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == id {
				// Found the element, now extract its text content
				var text string
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						text += c.Data
					}
				}
				return strings.TrimSpace(text)
			}
		}
	}

	// Recursively search children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findElementByID(c, id); result != "" {
			return result
		}
	}

	return ""
}

func getData(httpUrl string) (string, error) {
	resp, err := helpers.GetData(httpUrl)
	if err != nil {
		return "", err
	}
	doc, err := html.Parse(bytes.NewReader(resp))
	if err != nil {
		return "", err
	}

	question := findElementByID(doc, "human-question")

	return question, nil
}

func sendAnswer(httpUrl string, answer string) (string, error) {
	formData := url.Values{
		"username": {os.Getenv("S01E01_USER")},
		"password": {os.Getenv("S01E01_PASSWORD")},
		"answer":   {answer},
	}

	resp, err := http.PostForm(httpUrl, formData)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	taskUrl := os.Getenv("S01E01_URL")
	data, err := getData(taskUrl)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(data)

	openAI := openaiservice.New()

	req := openaiservice.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Odpowiadasz na pytania o date, zwracaj tylko rok",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: data,
			},
		},
		Model:  openai.GPT4, // or leave empty for default GPT-4
		Stream: false,
	}
	ctx := context.Background()
	response, _, err := openAI.Completion(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response
	if response != nil {
		log.Printf("AI Response: %s\n", response.Choices[0].Message.Content)
		answer := response.Choices[0].Message.Content
		key, err := sendAnswer(taskUrl, answer)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Response from xyz: %s", key)
	}
}
