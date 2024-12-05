package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type OutputMessage struct {
	Answer string `json:"answer"`
}

type InputMessage struct {
	Question string `json:"question"`
}

func RobotHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Request method: %s", r.Method)

	agent := InitAgent()
	var message InputMessage
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println(message.Question)
	Answer, err := agent.Process(message.Question)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output := OutputMessage{Answer: Answer}
	json.NewEncoder(w).Encode(output)
}
