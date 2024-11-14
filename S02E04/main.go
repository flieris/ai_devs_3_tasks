package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"maps"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

const (
	systemPromptImage = `
  <rules>
  - Extract and output only the text from this image exactly as it appears
  - Do not translate, explain, or describe anything else—just provide the text exactly as written.
  </rules>`
	systemPrompt = `
<rules>
- you are going to get a json file in the following format:
{
"file_name_1": "file_content1",
 "file_name_2":"file_content2"
}
- you have to read the values of the json and categorize if the content of the file talks about people or machines
- file content sometimes do not refer to either people or machines, in that case ignore the file
- in people field only include files that refers about captured people or sightings
- in hardware field only include files about fixed hardware issues (ignore any related to software)
- the content can be in either Polish or English
- categorize the file in the following way:
{
  "people": ["plik1.txt", "plik2.mp3", "plikN.png"],
  "hardware": ["plik4.txt", "plik5.png", "plik6.mp3"],
}
- return just json, without any formatting
</rules>
  `
)

type CategorizedFiles struct {
	People   []string `json:"people"`
	Hardware []string `json:"hardware"`
}

func transcribeAudio(audioFiles []string, path string) (map[string]string, error) {
	client := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	transcriptions := make(map[string]string)
	for _, file := range audioFiles {
		response, err := client.Transcribe(ctx, openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: path + "/" + file,
		})
		if err != nil {
			return nil, err
		}

		transcriptions[file] = response.Text
	}

	return transcriptions, nil
}

func transcribeImage(imageFiles []string, path string) (map[string]string, error) {
	imagesMap := make(map[string]string)
	for _, image := range imageFiles {

		data, err := os.ReadFile(path + "/" + image)
		if err != nil {
			continue
		}

		encoded := base64.StdEncoding.EncodeToString(data)
		req := llmservices.OllamaChatCompletion{
			Messages: []llmservices.OllamaChatCompletionMessage{
				{
					Role:    "system",
					Content: systemPromptImage,
				},
				{
					Role:    "user",
					Content: "Output text from this picture:",
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
		imagesMap[image] = response.Message.Content
	}

	return imagesMap, nil
}

func transcribeText(textFiles []string, path string) (map[string]string, error) {
	textMap := make(map[string]string)

	for _, text := range textFiles {
		file, err := os.Open(path + "/" + text)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		output, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		textMap[text] = string(output)
	}

	return textMap, nil
}

func categorizeFiles(files []fs.DirEntry, path string) (textFiles, imageFiles, audioFiles, directories []string) {
	for _, file := range files {
		if file.IsDir() {
			directories = append(directories, file.Name())
		} else {
			ext := filepath.Ext(file.Name())
			switch ext {
			case ".txt":
				textFiles = append(textFiles, file.Name())
			case ".jpg", ".jpeg", ".png", ".gif":
				imageFiles = append(imageFiles, file.Name())
			case ".mp3", ".wav", ".flac":
				audioFiles = append(audioFiles, file.Name())
			}
		}
	}
	return
}

func categorize(data []byte) (string, error) {
	openAI := llmservices.NewOpenAiServcie()

	req := llmservices.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: string(data),
			},
		},
		Model:  openai.GPT4o, // or leave empty for default GPT-4
		Stream: false,
	}
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

func main() {
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
	textFiles, imageFiles, audioFiles, dirs := categorizeFiles(files, filesPath)

	fmt.Printf("Text files: %v\n", textFiles)
	fmt.Printf("Image files: %v\n", imageFiles)
	fmt.Printf("Audio files: %v\n", audioFiles)
	fmt.Printf("Directories: %v\n", dirs)

	imageTranscriptions, err := transcribeImage(imageFiles, filesPath)
	if err != nil {
		log.Fatalf("Error transcribing images: %v", err)
	}
	textTranscriptions, err := transcribeText(textFiles, filesPath)
	if err != nil {
		log.Fatalf("Error transcribing text: %v", err)
	}

	audioTranscriptions, err := transcribeAudio(audioFiles, filesPath)
	if err != nil {
		log.Fatalf("Error transcribing audio: %v", err)
	}

	transcriptions := make(map[string]string)

	maps.Copy(transcriptions, imageTranscriptions)
	maps.Copy(transcriptions, textTranscriptions)
	maps.Copy(transcriptions, audioTranscriptions)

	jsonData, err := json.Marshal(transcriptions)
	if err != nil {
		log.Fatalf("Error converting transcriptions into JSON: %v", err)
	}

	resp, err := categorize(jsonData)
	if err != nil {
		log.Fatalf("Error categorizing files: %v", err)
	}

	fmt.Println(resp)

	var categorizedFiles CategorizedFiles

	err = json.Unmarshal([]byte(resp), &categorizedFiles)
	if err != nil {
		log.Fatalf("Error unmashalling JSON response: %v", err)
	}
	fmt.Printf("People-related files: %v\n", categorizedFiles.People)
	fmt.Printf("Hardware-related files: %v\n", categorizedFiles.People)

	message := helpers.JsonAnswer{
		Task:   "kategorie",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: categorizedFiles,
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
