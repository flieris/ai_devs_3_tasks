package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

const (
	systemPrompt = `
  <objective>
    You are a keyword retrival service. You will be given a document and a chunk data. Your task is to output all keywords that relates chunk and document.
  </objective>
<rules>
- Base your work on the document in <document> field and report data in <chunk> field.
- your main source of information should be e report data in <chunk> field
- you can use the document in <document> field to assist you if there is any data that relatesto the <chunk>
- The documents are in Polish language.
- Output must include a focused list of nouns in the Nominative Case, explicitly identifying occupations, technologies, and key themes mentioned in the document related to the characters referred to in the chunk.
- Prioritize the extraction of occupations such as 'nauczyciel', 'programista Javascript', 'laborant', etc., ensuring they are clearly indicated in the output.
- instead of "frontend development" use "JavaScript"
- Include synonyms and related terms that capture broader contexts or variations of roles.
- Output should be a comprehensive comma-separated list of at least 10 elements.
- Emphasize the connections between the identified occupations and the characters in the document, providing clarity on their relevance.
- Provide examples of keywords that would qualify as occupations related to the characters to guide the retrieval process.
- there is also a filename provided ".txt", extract sektor name from it.
</rules>
  `
)

func createDocumentPrompt(document map[string]string) (documentPrompt string) {
	documentPrompt = "<document>\n"
	for _, content := range document {
		documentPrompt = documentPrompt + content
	}
	documentPrompt = documentPrompt + "\n</document>"
	return
}

func askLlm(documentPrompt, prompt string) (string, error) {
	openAI := llmservices.NewOpenAiServcie()

	req := llmservices.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("%s\n%s", systemPrompt, documentPrompt),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Model:  openai.GPT4oMini, // or leave empty for default GPT-4
		Stream: false,
	}
	fmt.Printf("%s\n%s\n", systemPrompt, documentPrompt)
	ctx := context.Background()
	response, _, err := openAI.Completion(ctx, req)
	if err != nil {
		log.Printf("Error from OpenAI api: %v", err)
		return "", nil
	}

	// Print the response
	if response != nil {
		log.Printf("AI Response: %s\n", response.Choices[0].Message.Content)
		answer := response.Choices[0].Message.Content
		return answer, nil
	}
	return "", errors.New("No resposne from api")
}

func getMetadata(reportFiles map[string]helpers.FileData, factFiles map[string]helpers.FileData) (metadata map[string]string) {
	metadata = make(map[string]string)
	for file, content := range reportFiles {
		// prepare documents to run with ai
		documentPrompt := ""
		for _, commonFile := range content.CommonFiles {
			if factData, exists := factFiles[commonFile]; exists {
				documentPrompt = fmt.Sprintf("<document>%s</document>", factData.Content)
			}
		}
		prompt := fmt.Sprintf("<chunk>\nfilename: %s: %s\n</chunk>\nPlease extract all keywords that might be of interest in relatation of a chunk.", file, content.Content)
		fmt.Println(prompt)
		answer, err := askLlm(documentPrompt, prompt)
		if err != nil {
			log.Printf("Warning, did not get response from openai: %v", err)
			continue
		}
		metadata[file] = answer
		// time.Sleep(60 * time.Second)
	}
	return
}

func main() {
	reportFiles := make(map[string]helpers.FileData)
	factFiles := make(map[string]helpers.FileData)
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	filesZipPath := "pliki-z-fabryki.zip"
	filesPath := "pliki-z-fabryki"
	err = helpers.GetZip(os.Getenv("S02E04_URL"), filesZipPath)
	if err != nil {
		log.Fatalf("Error getting zip: %v", err)
	}

	err = helpers.Unzip(filesZipPath, filesPath)
	if err != nil {
		log.Fatalf("Error getting zip: %v", err)
	}

	files, err := os.ReadDir(filesPath)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}
	if !helpers.FileExists("reports.json") || !helpers.FileExists("facts.json") {
		reportFiles, factFiles, err = CategorizeFiles(files, filesPath)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		helpers.WriteMapToFile("reports.json", reportFiles)
		helpers.WriteMapToFile("facts.json", factFiles)
	} else {
		reportFiles, err = helpers.ReadMapFromFile("reports.json")
		if err != nil {
			log.Fatalf("Error reading reports.json file: %v", err)
		}
		factFiles, err = helpers.ReadMapFromFile("facts.json")
		if err != nil {
			log.Fatalf("Error reading facts.json file: %v", err)
		}
	}

	CorrelateFilesBetweenGroups(reportFiles, factFiles)

	//	documentPrompt := createDocumentPrompt(factFiles)
	//	fmt.Println(documentPrompt)

	metadata := getMetadata(reportFiles, factFiles)
	//	metadataJson, err := json.Marshal(metadata)
	//if err != nil {
	//	log.Fatalf("Error converting metadata to JSON: %v", err)
	//}
	message := helpers.JsonAnswer{
		Task:   "dokumenty",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: metadata,
	}
	fmt.Println(message)
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
