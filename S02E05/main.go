package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

const (
	imageSysPrompt = `
<rules>
- Answer the questions in one sentence or shorter.
- When refering to fruits, provide the name of the fruit
- try to give an accurate name of the fruit
- anything referencing pineapple and cake might refer to a pineapple pizza
- for market square use hint: "Widok na kościół od strony 'Adasia'"
- try to guess the name of the city from the description
- Name "Bomba" refers to Rafał Bomba
</rules>
<context>
  `
)

func getMarkdownAndLinks(htmlData []byte) ([]string, map[string]string, error) {
	// Convert HTML to Markdown
	baseDataUrl := os.Getenv("S02E05_BASE_DATA_URL")
	markdown, err := htmltomarkdown.ConvertString(string(htmlData))
	if err != nil {
		return nil, nil, err
	}

	// Parse HTML to extract image and audio links
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlData))
	if err != nil {
		return nil, nil, err
	}

	links := make(map[string]string)
	doc.Find("figure").Each(func(index int, figure *goquery.Selection) {
		img := figure.Find("img")
		src, exists := img.Attr("src")
		if exists {
			links[baseDataUrl+src] = figure.Find("figcaption").Text()
		}
	})

	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		src, exists := item.Attr("href")
		if exists {
			links[baseDataUrl+src] = src
		}
	})

	// Print Markdown and links
	chapters := splitMarkdownIntoChapters(markdown)

	return chapters, links, nil
}

func splitMarkdownIntoChapters(markdown string) []string {
	// Regular expression to match headers (e.g., # Chapter, ## Section)
	re := regexp.MustCompile(`(?m)^#{1,6} .+`)
	indices := re.FindAllStringIndex(markdown, -1)

	var chapters []string
	var lastIndex int
	for _, index := range indices {
		if lastIndex != 0 {
			chapters = append(chapters, strings.TrimSpace(markdown[lastIndex:index[0]]))
		}
		lastIndex = index[0]
	}
	// Add the last chapter
	if lastIndex < len(markdown) {
		chapters = append(chapters, strings.TrimSpace(markdown[lastIndex:]))
	}

	return chapters
}

func DownloadAndTranscribeAudio(audioURL string) (string, error) {
	out, err := os.Create("rafal.mp3")
	if err != nil {
		return "", err
	}
	defer out.Close()
	resp, err := helpers.GetData(audioURL)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(out, bytes.NewReader(resp))
	if err != nil {
		return "", err
	}
	client := llmservices.NewOpenAiServcie()
	ctx := context.Background()
	response, err := client.Transcribe(ctx, openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: out.Name(),
	})
	if err != nil {
		return "", err
	}

	return response.Text, nil
}

func DownloadAndTranscribeImage(imageURL string, imageDescription string) (string, error) {
	resp, err := helpers.GetData(imageURL)
	if err != nil {
		return "", err
	}
	fmt.Println(imageURL)
	encoded := base64.StdEncoding.EncodeToString(resp)
	req := llmservices.OllamaChatCompletion{
		Messages: []llmservices.OllamaChatCompletionMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf("Describe the image concisely. You can base description image description: %s ", imageDescription),
				Images:  []string{encoded},
			},
		},
		Model:  "llama3.2-vision:11b",
		Stream: false,
	}
	reqUrl := "http://localhost:11434/api/chat"
	response, err := llmservices.SendChatCompletion(req, reqUrl)
	if err != nil {
		return "", err
	}

	return response.Message.Content, nil
}

func AskQuestion(document []string, question string) (string, error) {
	openAI := llmservices.NewOpenAiServcie()

	req := llmservices.CompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: imageSysPrompt + strings.Join(document, "\n") + "</context>",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
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
	articleHtml, err := helpers.GetData(os.Getenv("S02E05_URL_1"))
	if err != nil {
		log.Fatalf("Error getting html body: %v", err)
	}
	questions, err := helpers.GetData(os.Getenv("S02E05_URL_2"))
	if err != nil {
		log.Fatalf("Error getting questions: %v", err)
	}
	chapters, links, err := getMarkdownAndLinks(articleHtml)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	questionsList := strings.Split(string(questions), "\n")
	for link, description := range links {
		if strings.HasSuffix(link, "mp3") {
			//			transcritpion, err := DownloadAndTranscribeAudio(link)
			transcritpion := `No i co teraz? No to teraz mnie już nie powstrzymacie. Jesteśmy tutaj sami. I trzeba tylko wykonać plan. Jak ja się do tego w ogóle zmieszczę? W sumie dobrze zaplanowałem powstrzymanie. Adam miał rację. Jedna informacja powinna się nam w czasie
. Jedna informacja. Dwa lata wcześniej. Posunie całe badania do przodu i wtedy już będzie z górki. Czekaj, na odwagę. Z truskawką. Mowę nie wywoła. Ale z ludźmi? Dobra, jedna myfa mowę pewnie wywoła. Ale Adam mówi, że to jest stabilne. Że to si
ę wszystko uda. Trzeba tylko cofnąć się w czasie. Jeden i jedyny raz. Do Grudziądza. Znaleźć hotel. Ile może być hoteli w Grudziądzu? Ja nie wiem, ale na pewno znajdę jeden. I potem czekać. Spokojnie czekać dwa lata. Tyle jestem w stanie zrobić
. Resztę mam zapisane na kartce. No to co? No to siup. Wpisujemy. Czekajcie. Koordynaty są Grudziądz. Dobra. Batman nie wchodzi. A jest w menu. Człowiek. Dobra. Jeszcze gry. Wezmę ze sobą trochę. Dobra. Jestem gotowy. Jeszcze jedno. Na odwagę. 
Tak na cześć. O to by nie wszedł. Nie. Dobra. Naciskamy. Czekamy. To licz szybciej. Ile można czekać? Człowieku. Jestem gotowy. No to bziom.`
			//		if err != nil {
			//		log.Fatalf("Error: %v", err)
			//}
			chapters = append(chapters, transcritpion)
		}
		if strings.HasSuffix(link, "png") {
			transcriptionm, err := DownloadAndTranscribeImage(link, description)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			chapters = append(chapters, transcriptionm)
		}
	}
	fmt.Println(questionsList)
	for _, question := range questionsList {
		answer, err := AskQuestion(chapters, question)
		if err != nil {
			log.Fatalf("Error: %v ", err)
		}
		fmt.Println(answer)
		time.Sleep(60 * time.Second)
	}
}
