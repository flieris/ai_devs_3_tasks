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

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
	"github.com/qdrant/go-client/qdrant"
	"github.com/sashabaranov/go-openai"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func collectionExists(ctx context.Context, client *qdrant.Client, collectionName string) bool {
	exists, _ := client.CollectionExists(ctx, collectionName)

	return exists
}

func hasPoints(ctx context.Context, client *qdrant.Client, collectionName string) bool {
	response, err := client.Count(ctx, &qdrant.CountPoints{
		CollectionName: collectionName,
	})
	if err != nil {
		log.Fatalf("Error counting points: %v", err)
	}
	return response > 0
}

func ParseQuestions(ctx context.Context, openAi *llmservices.OpenAiService, question Question) llmservices.MessageContent {
	questionResponse := llmservices.MessageContent{}
	fmt.Printf("%s\n%s\n<document>\n%s\n</document>\n", SystemPrompt, ResponseFormat, strings.Join(question.Related, ""))
	msg := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s\n%s\n<document>\n%s\n</document>", SystemPrompt, ResponseFormat, strings.Join(question.Related, "")),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: question.Question,
		},
	}
	openAi.SendMessage(ctx, &msg, openai.GPT4oMini)
	response := msg[len(msg)-1].Content
	log.Printf("Response: %s", response)
	err := json.Unmarshal([]byte(response), &questionResponse)
	if err != nil {
		log.Fatalf("Error unmarshalling questions: %v", err)
	}

	return questionResponse
}

func main() {
	collectionName := "S04E05"
	openAI := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: os.Getenv("QDRANT_URL"),
		Port: 6334,
	})
	if !collectionExists(ctx, client, collectionName) {
		client.CreateCollection(context.Background(), &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     1536,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		fmt.Printf("Collection %s created\n", collectionName)
	}
	data, err := helpers.DownloadFile(os.Getenv("S04E05_URL_1"))
	if err != nil {
		log.Fatalf("Error getting data: %v", err)
	}

	questions, err := PrepareQuestions(ctx, openAI, os.Getenv("S04E05_URL_2"))
	if err != nil {
		log.Fatalf("Error getting questions: %v", err)
	}
	f, r, err := pdf.Open(data)
	if err != nil {
		log.Fatalf("Error opening PDF: %v", err)
	}
	defer f.Close()

	// Read all pages
	totalPage := r.NumPage()
	pdfText := []string{}
	if !hasPoints(ctx, client, collectionName) {
		images, err := GetImagesFromPdf(data, "images")
		if err != nil {
			log.Fatalf("Error getting images: %v", err)
		}
		imagesToAnalyze, err := IdentifyImages(images)
		if err != nil {
			log.Fatalf("Error identifying images: %v", err)
		}
		analziedImage, err := AnalyzeImages(ctx, openAI, imagesToAnalyze)
		if err != nil {
			log.Fatalf("Error analyzing images: %v", err)
		}
		imageEmedding, err := openAI.GetEmbeddings(ctx, strings.Join(analziedImage, ""))
		if err != nil {
			log.Fatalf("Error getting embeddings: %v", err)
		}
		_, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{
			CollectionName: collectionName,
			Points: []*qdrant.PointStruct{
				{
					Id:      qdrant.NewIDUUID(uuid.New().String()),
					Vectors: qdrant.NewVectorsDense(imageEmedding),
					Payload: qdrant.NewValueMap(map[string]any{"content": strings.Join(analziedImage, "")}),
				},
			},
		})

		for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
			composeResponse, err := ParsePDFPage(ctx, openAI, r, pageIndex)
			if err != nil {
				log.Fatalf("Error parsing PDF page: %v", err)
			}
			pdfText = append(pdfText, composeResponse)
			embeddings, err := openAI.GetEmbeddings(ctx, composeResponse)
			if err != nil {
				log.Fatalf("Error getting embeddings: %v", err)
			}
			_, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{
				CollectionName: collectionName,
				Points: []*qdrant.PointStruct{
					{
						Id:      qdrant.NewIDUUID(uuid.New().String()),
						Vectors: qdrant.NewVectorsDense(embeddings),
						Payload: qdrant.NewValueMap(map[string]any{"content": composeResponse, "page_index": pageIndex}),
					},
				},
			})
		}
	}
	questionsToAsk, err := GetRelatedText(questions, collectionName, client)
	if err != nil {
		log.Fatalf("Error getting related text: %v", err)
	}

	answers := make(map[string]string)
	for _, question := range questionsToAsk {
		response := ParseQuestions(ctx, openAI, question)
		answers[question.QuestionId] = strings.Join(response.Answers, "")
	}

	message := helpers.JsonAnswer{
		Task:   "notes",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: answers,
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %v", respCentral)
}
