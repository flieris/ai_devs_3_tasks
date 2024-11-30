package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/qdrant/go-client/qdrant"
	"github.com/sashabaranov/go-openai"
)

type Question struct {
	Question   string    `json:"question"`
	Related    []string  `json:"related,omitempty"`
	QuestionId string    `json:"question_id"`
	Embeddings []float32 `json:"embeddings,omitempty"`
}

func ParsePDFPage(ctx context.Context, openAi *llmservices.OpenAiService, pdfFile *pdf.Reader, pageIndex int) (string, error) {
	p := pdfFile.Page(pageIndex)
	if p.V.IsNull() {
		return "", fmt.Errorf("Error reading page %d", pageIndex)
	}

	text, err := p.GetPlainText(nil)
	if err != nil {
		return "", err
	}

	message := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("<objective>Compose the given text in Polish loading. Output the text in Polish, and give the output in JSON format as in <response_format> field. Output as a single string.</objective>\n%s", ResponseFormat),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		},
	}
	openAi.SendMessage(ctx, &message, openai.GPT4oMini)
	composeResponse := llmservices.MessageContent{}
	response := message[len(message)-1].Content
	err = json.Unmarshal([]byte(response), &composeResponse)
	if err != nil {
		log.Fatalf("Error unmarshalling questions: %v", err)
	}
	return strings.Join(composeResponse.Answers, ""), nil
}

func GetImagesFromPdf(pdfFile string, outputDir string) ([]string, error) {
	_, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	err = api.ExtractImagesFile(pdfFile, outputDir, []string{"19"}, nil)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}

	images := []string{}
	for _, file := range files {
		images = append(images, outputDir+"/"+file.Name())
	}

	return images, nil
}

func IdentifyImages(images []string) ([]string, error) {
	outputImages := []string{}
	for _, img := range images {
		imageBytes, err := os.ReadFile(img)
		if err != nil {
			return nil, err
		}
		encoded := base64.StdEncoding.EncodeToString(imageBytes)
		req := llmservices.OllamaChatCompletion{
			Messages: []llmservices.OllamaChatCompletionMessage{
				{
					Role:    "user",
					Content: "Identify if the image is text or not. Output 0 if it is, and 1 if it is not. Output as a single string.",
					Images:  []string{encoded},
				},
			},
			Model:  "llama3.2-vision:11b",
			Stream: false,
		}
		reqUrl := "http://localhost:11434/api/chat"
		response, err := llmservices.SendChatCompletion(req, reqUrl)
		if err != nil {
			return nil, err
		}
		if response.Message.Content == "0" {
			outputImages = append(outputImages, img)
		}
	}
	return outputImages, nil
}

func AnalyzeImages(ctx context.Context, openAi *llmservices.OpenAiService, images []string) ([]string, error) {
	imageMultiContent := []openai.ChatMessagePart{
		{
			Type: openai.ChatMessagePartTypeText,
			Text: fmt.Sprintf("Please analyze this images: %s", strings.Join(images, ", ")),
		},
	}
	for _, image := range images {
		imageBytes, err := os.ReadFile(image)
		if err != nil {
			return nil, err
		}
		encoded := base64.StdEncoding.EncodeToString(imageBytes)
		imageMultiContent = append(imageMultiContent, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: fmt.Sprintf("data:image/jpeg;base64,%s", encoded),
			},
		})
	}
	message := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("<objective>You are given an image of Polish text. Output text in JSON format according to the format in <response_format> field</objective>\n%s", ResponseFormat),
		},
		{
			Role:         openai.ChatMessageRoleUser,
			MultiContent: imageMultiContent,
		},
	}
	err := openAi.SendMessage(ctx, &message, openai.GPT4o)
	if err != nil {
		return nil, err
	}
	var output llmservices.MessageContent
	err = json.Unmarshal([]byte(message[len(message)-1].Content), &output)
	if err != nil {
		return nil, err
	}
	return output.Answers, nil
}

func PrepareQuestions(ctx context.Context, openAI *llmservices.OpenAiService, urlToQuestions string) ([]Question, error) {
	var questions []Question
	questionsMap := map[string]string{}
	questionsRaw, err := helpers.GetData(urlToQuestions)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(questionsRaw, &questionsMap)
	if err != nil {
		return nil, err
	}

	for questionId, question := range questionsMap {

		embeddings, err := openAI.GetEmbeddings(ctx, question)
		if err != nil {
			return nil, err
		}
		questionStruct := Question{
			Question:   question,
			QuestionId: questionId,
			Embeddings: embeddings,
		}
		questions = append(questions, questionStruct)
	}

	return questions, nil
}

func GetRelatedText(questions []Question, collectionName string, client *qdrant.Client) ([]Question, error) {
	scoredPoints := []Question{}
	for _, embeddings := range questions {
		limit := uint64(10)
		searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
			CollectionName: collectionName,
			Query:          qdrant.NewQueryDense(embeddings.Embeddings),
			Limit:          &limit,
			WithPayload:    qdrant.NewWithPayload(true),
		})
		if err != nil {
			return nil, err
		}
		for _, result := range searchResult {
			embeddings.Related = append(embeddings.Related, result.Payload["content"].GetStringValue())
		}
		scoredPoints = append(scoredPoints, embeddings)
	}
	return scoredPoints, nil
}
