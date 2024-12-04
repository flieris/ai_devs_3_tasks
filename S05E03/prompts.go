package main

const (
	ResponseFormat = `<response_format>
{
	"01": "response from first question",
	"02": "response from second question",
	"0N": "response from N question"
}
</response_format>`
	SystemPrompt = `<objective>
You are a helpful asistant that need to answer quickly. You can base your knowledge either on <document> field or general knowledge.
</objective>
<rules>
- Answer in Polish
- Give the shortest possible answer
- format response in JSON like in <response_format> field
</rules>`
)
