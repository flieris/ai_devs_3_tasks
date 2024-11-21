package helpers

import (
	"ai_devs_3_tasks/llmservices"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	keywordsPrompt = `
  <rules>
  - Your output should always be a comma seperated list of items
  - return just the keywords, do not add additional text
  - if there are any Polish names in the text, include them in the keywords
  - The documents are in Polish language.
  - Output must include a focused list of nouns in the Nominative Case, explicitly identifying occupations, technologies, and key themes mentioned in the document related to the characters referred to in the chunk.
  - Prioritize the extraction of occupations such as 'nauczyciel', 'programista Javascript', 'laborant', etc., ensuring they are clearly indicated in the output.
  - Include synonyms and related terms that capture broader contexts or variations of roles.
  - Output should be a comprehensive comma-separated list of at least 10 elements.
  - Emphasize the connections between the identified occupations and the characters in the document, providing clarity on their relevance.
  - Provide examples of keywords that would qualify as occupations related to the characters to guide the retrieval process.
  - if there is a date in the document name, add it to the keywords in the YYYY-MM-DD fromat
  - Do not add anything to the text. Like "Słowa kluczowe"
  </rules>
  `
)

type FileData struct {
	Content     string   `json:"content"`
	Keywords    string   `json:"keywords"`
	CommonFiles []string `json:"common_files,omitempty"`
}

func GetKeywords(document string) (keywords string, err error) {
	req := llmservices.OllamaChatCompletion{
		Messages: []llmservices.OllamaChatCompletionMessage{
			{
				Role:    "system",
				Content: keywordsPrompt,
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Wyciągnij słowa kluczowe z poniszego dokumentu:\n%s ", document),
			},
		},
		Model:  "SpeakLeash/bielik-11b-v2.3-instruct:Q4_K_M",
		Stream: false,
	}
	reqUrl := "http://localhost:11434/api/chat"
	response, err := llmservices.SendChatCompletion(req, reqUrl)
	if err != nil {
		return "", err
	}
	keywords = response.Message.Content
	return
}

func WriteMapToFile(filename string, data map[string]FileData) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Error converting map to JSON: %v", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON to file: %v", err)
	}
}

func ReadMapFromFile(filename string) (map[string]FileData, error) {
	data := make(map[string]FileData)
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
