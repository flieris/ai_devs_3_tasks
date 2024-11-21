package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/qdrant/go-client/qdrant"
	"github.com/sashabaranov/go-openai"
)

type FileData struct {
	FileName  string    `json:"file_name"`
	Content   string    `json:"content"`
	Embedding []float32 `json:"embedding"`
}

func getEmbeddings(content string) ([]float32, error) {
	openAI := llmservices.NewOpenAiServcie()
	ctx := context.Background()

	req, err := openAI.CreateEmbedding(ctx, openai.EmbeddingRequest{
		Model: openai.SmallEmbedding3,
		Input: content,
	})
	if err != nil {
		return nil, err
	}
	return req.Data[0].Embedding, nil
}

func convertToInterfaceSlice(strings []string) []interface{} {
	interfaces := make([]interface{}, len(strings))
	for i, s := range strings {
		interfaces[i] = s
	}
	return interfaces
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	factoryDirPath := "pliki-z-fabryki/do-not-share"
	query := "W raporcie, z którego dnia znajduje się wzmianka o kradzieży prototypu broni?"
	queryEmbedding, err := getEmbeddings(query)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: os.Getenv("QDRANT_URL"),
		Port: 6334,
	})
	client.CreateCollection(context.Background(), &qdrant.CreateCollection{
		CollectionName: "S03E02",
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     1536,
			Distance: qdrant.Distance_Cosine,
		}),
	})

	files, err := os.ReadDir(factoryDirPath)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}
	var weaponFiles []FileData
	// var weaponFilesPoints []*qdrant.PointStruct
	var index uint64 = 0
	for _, file := range files {
		fileFd, err := os.Open(factoryDirPath + "/" + file.Name())
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		defer fileFd.Close()
		output, err := io.ReadAll(fileFd)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		baseName := strings.Split(strings.TrimSuffix(file.Name(), ".txt"), "_")
		reportDate := fmt.Sprintf("%s-%s-%s", baseName[0], baseName[1], baseName[2])
		keywords, err := helpers.GetKeywords(fmt.Sprintf("File %s, content: %s", file.Name(), string(output)))
		embed, err := getEmbeddings(fmt.Sprintf("File %s, Report Date: %s, content: %s", file.Name(), reportDate, string(output)))
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		data := FileData{
			FileName:  file.Name(),
			Content:   string(output),
			Embedding: embed,
		}
		weaponFiles = append(weaponFiles, data)

		_, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{
			CollectionName: "S03E02",
			Points: []*qdrant.PointStruct{
				{
					Id:      qdrant.NewIDNum(uint64(index)),
					Vectors: qdrant.NewVectorsDense(embed),
					Payload: qdrant.NewValueMap(map[string]any{"file_name": file.Name(), "content": string(output), "metadata": keywords, "report_date": reportDate}),
				},
			},
		})
		index++
	}
	limit := uint64(1)
	searchResult, err := client.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: "S03E02",
		Query:          qdrant.NewQueryDense(queryEmbedding),
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{qdrant.NewMatchText("metadata", "kradzież")},
		},
		WithPayload: qdrant.NewWithPayload(true),
		Limit:       &limit,
	})
	dateOfTheft := searchResult[0].GetPayload()["report_date"].GetStringValue()
	fmt.Println(dateOfTheft)

	message := helpers.JsonAnswer{
		Task:   "wektory",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: dateOfTheft,
	}
	fmt.Println(message)
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
