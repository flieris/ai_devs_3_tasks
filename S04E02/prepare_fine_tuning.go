package main

import (
	"ai_devs_3_tasks/helpers"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type FineTuningJsonl struct {
	Messages []FineTuningMessages `json:"messages"`
}

type FineTuningMessages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main2() {
	labDataZipPath := "lab_data.zip"
	labDataPath := "lab_data"
	err := helpers.GetZip(os.Getenv("S04E02_URL"), labDataZipPath)
	if err != nil {
		log.Fatalf("Error getting zip: %v", err)
	}

	err = helpers.Unzip(labDataZipPath, labDataPath)
	if err != nil {
		log.Fatalf("Error getting zip: %v", err)
	}

	files, err := os.ReadDir(labDataPath)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	tuningMap := make(map[string][]string)
	for _, file := range files {
		if file.Name() == "verify.txt" {
			continue
		}
		dataType := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		fileFd, err := os.Open(labDataPath + "/" + file.Name())
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		defer fileFd.Close()

		output, err := io.ReadAll(fileFd)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		tuningMap[dataType] = strings.Split(string(output), "\n")
	}

	var fineTuningStruct []FineTuningJsonl

	f, _ := os.Create("fine_tuning.jsonl")
	defer f.Close()
	for dataType, values := range tuningMap {
		for _, value := range values {
			if value == "" {
				continue
			}
			var interactionStruct FineTuningJsonl
			interactionStruct.Messages = []FineTuningMessages{
				{
					Role:    "system",
					Content: "Sklasyfikuj, czy podane próbki sa poprawne",
				},
				{
					Role:    "user",
					Content: value,
				},
				{
					Role:    "assistant",
					Content: dataType,
				},
			}
			fineTuningStruct = append(fineTuningStruct, interactionStruct)

			fineTuningJsonData, err := json.Marshal(interactionStruct)
			if err != nil {
				log.Fatalf("Error converting metadata to JSON: %v", err)
			}
			f.Write(fineTuningJsonData)
			f.Write([]byte("\n"))
		}
	}
}
