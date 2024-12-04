package main

import (
	"ai_devs_3_tasks/helpers"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type RafalResponse struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
	Hint    string      `json:"hint,omitempty"`
	Took    string      `json:"took,omitempty"`
	ReadIt  string      `json:"read-it,omitempty"`
}

type RafalChallange struct {
	Task string   `json:"task"`
	Data []string `json:"data"`
}

type RafalAnswer struct {
	ApiKey    string   `json:"apikey"`
	Timestamp int      `json:"timestamp"`
	Signature string   `json:"signature"`
	Answers   []string `json:"answer"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func getChallenge(challangeUrl string, responseChan chan<- map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()

	urlRegex := regexp.MustCompile(`(https?://[^\s]+)`)

	var challange RafalChallange

	err := helpers.SendRequest(challangeUrl, nil, &challange)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	log.Printf("Started processing challange: %s", challange.Task)
	match := urlRegex.FindString(challange.Task)
	if match != "" {
		response := HandleWebsiteChallange(challange, match)
		responseChan <- response
	} else {
		response := HandleNormalChallange(challange)
		responseChan <- response
	}
	log.Printf("Finished processing challange: %s", challange.Task)
}

func main() {
	var signResponse RafalResponse
	var wg sync.WaitGroup
	passwordPayload := map[string]string{
		"password": os.Getenv("S05E03_PASSWORD"),
	}
	start := time.Now()
	err := helpers.SendRequest(os.Getenv("S05E03_URL"), passwordPayload, &signResponse)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	var signedResponse RafalResponse
	signPayload := map[string]string{
		"sign": signResponse.Message.(string),
	}
	err = helpers.SendRequest(os.Getenv("S05E03_URL"), signPayload, &signedResponse)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	// this feels like an abomination
	challanges := signedResponse.Message.(map[string]interface{})["challenges"]
	timestamp := signedResponse.Message.(map[string]interface{})["timestamp"].(float64)
	wg.Add(len(challanges.([]interface{})))
	log.Printf("Timestamp: %d", int(timestamp))
	responseChannel := make(chan map[string]string, len(challanges.([]interface{})))
	for _, challange := range challanges.([]interface{}) {
		go getChallenge(challange.(string), responseChannel, &wg)
	}

	go func() {
		wg.Wait()
		close(responseChannel)
	}()
	answers := []string{}
	for response := range responseChannel {
		for _, value := range response {
			answers = append(answers, value)
		}
	}
	log.Printf("Final response: %v", answers)
	answerPayload := RafalAnswer{
		ApiKey:    os.Getenv("CENTRAL_API_KEY"),
		Timestamp: int(timestamp),
		Signature: signedResponse.Message.(map[string]interface{})["signature"].(string),
		Answers:   answers,
	}
	log.Printf("Answer payload: %v", answerPayload)
	var finalResponse RafalResponse
	err = helpers.SendRequest(os.Getenv("S05E03_URL"), answerPayload, &finalResponse)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	end := time.Now()
	log.Printf("Time elapsed: %v", end.Sub(start))
	log.Printf("Final response: %v", finalResponse)
}
