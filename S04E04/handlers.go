package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type InputMessage struct {
	Instruction string `json:"instruction"`
}

type OutputMessage struct {
	Description string `json:"description"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func RobotInstructionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		message := ErrorMessage{Message: "This is a GET request"}
		json.NewEncoder(w).Encode(message)
	case "POST":
		var message InputMessage
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println(message.Instruction)

		parsedInstruction := ParseInstruction(message.Instruction)

		scanResult := MoveAndScan(parsedInstruction.Answers[0])

		json.NewEncoder(w).Encode(OutputMessage{Description: scanResult})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
