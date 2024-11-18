package main

import (
	"ai_devs_3_tasks/llmservices"
	"encoding/base64"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"

	"github.com/sashabaranov/go-openai"
)

const (
	systemPrompt = `
<objective>
You are location detective, your task is to specify which polish city is represented on map fragmetns
</objective>
<rules>
- you have a map with  4 fragments of a map of a polish town
- one of the fragments is incorrect and show different town
- there are granaries and some strongholds in the town we are looking for
- put your thinking process in the <thinking> field and the answer in <answer>
</rules>
<thinking>
</thinking>
<answer>
</answer>
`
)

// TODO: rewrite openAI service agent to allow image upload and prompt

func getMapFragments() (fragmentsBase64 []string, err error) {
	file, err := os.Open("images/map.jpg")
	if err != nil {
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return
	}

	fragments := []image.Rectangle{
		image.Rect(320, 120, 820, 720),
		image.Rect(920, 120, 1350, 720),
		image.Rect(250, 750, 1380, 1230),
		image.Rect(370, 1280, 1300, 1800),
	}
	for i, rect := range fragments {
		fragment := img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(rect)

		outFile, err := os.Create("images/fragment_" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			continue
		}
		defer outFile.Close()

		err = jpeg.Encode(outFile, fragment, nil)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(outFile.Name())
		if err != nil {
			continue
		}

		encoded := base64.StdEncoding.EncodeToString(data)
		fragmentsBase64 = append(fragmentsBase64, encoded)
	}
	return
}

func askLlm(mapFragments []string) {
	req := llmservices.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: mapFragments,
			},
		},
		Model:  openai.GPT4o, // or leave empty for default GPT-4
		Stream: false,
	}
}

func main() {
	fragments, err := getMapFragments()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	askLlm(fragments)
}
