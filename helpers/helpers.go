package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type JsonMessage struct {
	MsgID int64  `json:"msgID"`
	Text  string `json:"text"`
}

func SendJson(apiUrl string, message JsonMessage) (*JsonMessage, error) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseObj JsonMessage
	if err := json.NewDecoder(resp.Body).Decode(&responseObj); err != nil {
		return nil, err
	}
	return &responseObj, nil
}

func GetData(apiUrl string) ([]byte, error) {
	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a request with custom headers
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
