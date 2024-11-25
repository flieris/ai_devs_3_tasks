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
	"path"
	"slices"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

const (
	responseFormat = `
<response_format>
{
  "_thinking": "explanation of your interpretation and decision process",
  "answers": ["List of answers in string format"]
}
</response_format>
`
	systemPrompt = `
<objective>
You are image detecting service. Based on attached image you have to determine what task need to be done with it.
Interpret the attached image and then generate JSON object (without markdown/xml block).
<objective>
<rules>
- Analyze inputted photo
- In the field "_thinking" put your interpretation process. 
- Output only the specified JSON format in <response_format> field
- Determine on the state of the photo which operation to do
  - Operation you can do:
  - REPAIR FILE_NAME - for photos with visable white noise/glitches
  - DARKEN FILE_NAME - for photos that are way too bright
  - BRIGHTEN FILE_NAME - for photos that are way too dark
- The FILE_NAME will be in the url of the image you will get
- If there is nothing wrong with the image put "NOP" as the first element of answer array, and as second element describe the image in Polish language
- Format your output in the <response_format> field without the markdown block
</rules>
  `
	centralParserSystemPrompt = `
<objective>
Your task is to parse the text and get construct an url to an images based on url and .PNG files in the message. Format your output like in <response_format> field (without markdown block)
</objective>
<rules>
- If you can't determine url, use the default url in <default> field
- Format your output like in <response_format> field (without markdown block)
- images can be either just filenames with .PNG extension or whole url
- if there are NO files with .PNG extension and there is no url in the response, return empty response
</rules>
<default>
  `
	barbDescSystemPrompt = `
<objective>
Your task is to describe a woman named Barbara from the group of photos
</objective>
<rules>
- Describe Barbara in Polish language
- If the photo do not show a woman, ignore it.
- If there is more than 1 person in the photo, only focus on the person that appears on more than 1 photo
- In the field "_thinking" put your interpretation process. 
- Output only the specified JSON format in <response_format> field
- If Barbara has some characteristic points, please add them in description.
- characteristic can be stuff like: hair colour, tattos (describe shape, and what the tattoo presents), glassess etc.
</rules>
  `
)

type MessageContent struct {
	Thinking string   `json:"_thinking"`
	Answers  []string `json:"answers"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func analyzeImage(ctx context.Context, openAi *llmservices.OpenAiService, image string) ([]string, error) {
	resp, err := helpers.GetData(image)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(resp)
	message := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s\n%s", systemPrompt, responseFormat),
		},
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: fmt.Sprintf("Please analyze this image: %s", path.Base(image)),
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: fmt.Sprintf("data:image/jpeg;base64,%s", encoded),
					},
				},
			},
		},
	}
	err = openAi.SendMessage(ctx, &message, openai.GPT4o)
	if err != nil {
		return nil, err
	}
	var output MessageContent
	err = json.Unmarshal([]byte(message[len(message)-1].Content), &output)
	if err != nil {
		return nil, err
	}
	return output.Answers, nil
}

func analyzeImages(ctx context.Context, openAi *llmservices.OpenAiService, images []string) ([]string, error) {
	imageMultiContent := []openai.ChatMessagePart{
		{
			Type: openai.ChatMessagePartTypeText,
			Text: fmt.Sprintf("Please analyze this images: %s", strings.Join(images, ", ")),
		},
	}
	for _, image := range images {
		resp, err := helpers.GetData(image)
		if err != nil {
			return nil, err
		}
		encoded := base64.StdEncoding.EncodeToString(resp)
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
			Content: fmt.Sprintf("%s\n%s", barbDescSystemPrompt, responseFormat),
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
	log.Println(message[len(message)-1])
	var output MessageContent
	err = json.Unmarshal([]byte(message[len(message)-1].Content), &output)
	if err != nil {
		return nil, err
	}
	return output.Answers, nil
}

func sendQueryAndParseResponse(ctx context.Context, openAi *llmservices.OpenAiService, query string) (*MessageContent, error) {
	message := helpers.JsonAnswer{
		Task:   "photos",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: query,
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		return nil, err
	}
	log.Printf("Response from central: %v", respCentral.Message)
	initialMsg := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s\n%s</default>\n%s", centralParserSystemPrompt, os.Getenv("S04E01_URL"), responseFormat),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: respCentral.Message,
		},
	}

	openAi.SendMessage(ctx, &initialMsg, openai.GPT4oMini)
	var inputPhotosStruct MessageContent
	err = json.Unmarshal([]byte(initialMsg[len(initialMsg)-1].Content), &inputPhotosStruct)
	if err != nil {
		return nil, err
	}
	return &inputPhotosStruct, nil
}

func main() {
	openAi := llmservices.NewOpenAiServcie()
	ctx := context.Background()

	inputPhotosStruct, err := sendQueryAndParseResponse(ctx, openAi, "START")
	if err != nil {
		log.Fatalf("Error getting initial photos: %v", err)
	}
	var finalImages []string
	images := inputPhotosStruct.Answers
	for {
		var tmpImages []string
		for _, image := range images {
			if slices.Contains(finalImages, image) {
				continue
			}
			output, err := analyzeImage(ctx, openAi, image)
			if err != nil {
				log.Printf("Error: %v", err)
				break
			}
			if output[0] == "NOP" {
				finalImages = append(finalImages, image)
				continue
			}
			photoStruct, err := sendQueryAndParseResponse(ctx, openAi, output[0])
			if len(photoStruct.Answers) == 0 {
				finalImages = append(finalImages, image)
				continue
			}
			tmpImages = append(tmpImages, photoStruct.Answers[0])
		}
		images = tmpImages
		if len(images) == 0 {
			break
		}
	}
	finalOutput, err := analyzeImages(ctx, openAi, finalImages)
	if err != nil {
		log.Fatalf("Error analizing final images: %v", err)
	}
	log.Println(finalOutput)

	message := helpers.JsonAnswer{
		Task:   "photos",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: finalOutput[0],
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
