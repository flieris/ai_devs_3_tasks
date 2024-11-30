package main

const (
	ResponseFormat = `<response_format>
{
  "_thinking": "explanation of your interpretation and decision process",
  "answers": ["List of answers in string format"]
}
</response_format>`
	SystemPrompt = `<objective>
Analyze notes from a time-traveler named Rafał. Who time-traveled before 2024. Answer in JSON format without the markdown blocks.
</objective>
<rules>
- Answer only in Polish
- In the field "_thinking" put your interpretation process. 
- Output only the specified JSON format in <response_format> field
- Your task is to identify the specific year Rafał is describing as the one he has traveled to. Derive the year based on contextual historical knowledge, particularly technological advancements like the release of GPT-2 in 2019.
- Pay close attention to mentions of significant events and milestones, prioritizing those tied to known historical timelines (e.g., GPT-2 in 2019). Use the exact year of the technology.
- Rafał might have traveled to the specific point in time or before it. 
- Avoid bias toward future years (like 2024 or beyond) unless explicitly supported by the text. Instead, focus on identifying the year Rafał would most likely have traveled to based on described events.
- Provide reasoning for the chosen year, directly linking it to contextual clues in the notes.
- When asked about location include an explicit mention of the environment
- seek details about the location's physical characteristics, the location might not be a town name but a place near the town.
- Rafał's last know locations was around some caves
  </rules>`
)
