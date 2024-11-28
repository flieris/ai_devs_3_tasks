package main

import (
	"ai_devs_3_tasks/llmservices"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
)

const (
	ResponseFormat = `
{
  "_thinking": "explanation of your interpretation and decision process",
  "answers": ["List of answers in string format"]
}
  `
	SystemPrompt = `
<objective>
You are a message responder. You will be given an instruction in natural polish language. you have to decrypt set of instruction needed
</objective
<rules>
- The instruction will describe the movements of a drone
- In the field "_thinking" put your interpretation process. 
- Output only the specified JSON format in <response_format> field
- the inputed instructions will be in Polish
- the map size we are moving is a 4x4 grid
- The starting position is on (0,0)
- based on instruction output the final position in form of "(x,y)" include the parenthesis 
- in the answers field only output the position
</rules>
  `
)

var Mapa = map[string]string{
	"(0,0)": "start",
	"(0,1)": "łąka",
	"(0,2)": "drzewo",
	"(0,3)": "dom",
	"(1,0)": "łąka",
	"(1,1)": "młyn",
	"(1,2)": "łąka",
	"(1,3)": "łąka",
	"(2,0)": "łąka",
	"(2,1)": "łąka",
	"(2,2)": "skała",
	"(2,3)": "drzewa",
	"(3,0)": "pasmo górskie",
	"(3,1)": "pasmo górskie",
	"(3,2)": "samochód",
	"(3,3)": "jaskinia",
}

func ParseInstruction(instruction string) llmservices.MessageContent {
	openAi := llmservices.NewOpenAiServcie()
	ctx := context.Background()

	questionResponse := llmservices.MessageContent{}
	msg := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s\n%s", SystemPrompt, ResponseFormat),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: instruction,
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

func MoveAndScan(position string) string {
	return Mapa[position]
}
