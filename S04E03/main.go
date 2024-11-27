package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func CategorizeQuestions(ctx context.Context, openAi *llmservices.OpenAiService, questions map[string]string, links []string) map[string]llmservices.MessageContent {
	linksResponse := make(map[string]llmservices.MessageContent)
	for questionId, question := range questions {
		msg := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("%s<document>\n%s\n</document>%s", CategorizeLinksPrompt, strings.Join(links, ","), ResponseFormat),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		}
		openAi.SendMessage(ctx, &msg, openai.GPT4oMini)
		response := msg[len(msg)-1].Content
		tmp := llmservices.MessageContent{}
		err := json.Unmarshal([]byte(response), &tmp)
		if err != nil {
			log.Fatalf("Error unmarshalling questions: %v", err)
		}
		linksResponse[questionId] = tmp
	}

	return linksResponse
}

func AnswerQuestions(ctx context.Context, openAi *llmservices.OpenAiService, question string, documents []string) llmservices.MessageContent {
	questionResponse := llmservices.MessageContent{}
	msg := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s<document>\n%s\n</document>%s", QuestionPrompt, strings.Join(documents, ","), ResponseFormat),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: question,
		},
	}
	openAi.SendMessage(ctx, &msg, openai.GPT4oMini)
	response := msg[len(msg)-1].Content
	err := json.Unmarshal([]byte(response), &questionResponse)
	if err != nil {
		log.Fatalf("Error unmarshalling questions: %v", err)
	}

	return questionResponse
}

func main() {
	openAi := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	links := GetLinks(os.Getenv("S04E03_URL_2"), os.Getenv("S04E03_ALLOWED_DOMAINS"))
	questionsRaw, err := helpers.GetData(os.Getenv("S04E03_URL_1"))
	if err != nil {
		log.Fatalf("Error getting questions: %v", err)
	}
	questions := make(map[string]string)
	answers := make(map[string]string)
	err = json.Unmarshal(questionsRaw, &questions)
	if err != nil {
		log.Fatalf("Error unmarshalling questions: %v", err)
	}
	categorized := CategorizeQuestions(ctx, openAi, questions, links)
	for qId, links := range categorized {
		data := []string{}
		for _, link := range links.Answers {
			linkdata, err := GetMain(link)
			if err != nil {
				log.Fatalf("Error getting main body for link: %s %v", link, err)
			}
			data = append(data, linkdata...)
		}
		answerStruct := AnswerQuestions(ctx, openAi, questions[qId], data)
		answers[qId] = strings.Join(answerStruct.Answers, ",")
	}

	log.Println(answers)
	message := helpers.JsonAnswer{
		Task:   "softo",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: answers,
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
