package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flieris/ai_devs_3_tasks/openaiservice"
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

func getData() (string, error) {
	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a request with custom headers
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return "", err
	}

	// Add headers to appear more like a regular browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	question := findElementByID(doc, "human-question")

	return question, nil
}

func main() {
	data, err := getData()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(data)

	openAI := openaiservice.New()

	req := openaiservice.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
		Model:  openai.GPT4, // or leave empty for default GPT-4
		Stream: false,
	}
	ctx := context.Background()
	response, stream, err := openAI.Completion(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response
	if response != nil {
		log.Printf("AI Response: %s\n", response.Choices[0].Message.Content)
	}
}
