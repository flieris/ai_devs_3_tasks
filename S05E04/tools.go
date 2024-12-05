package main

import (
	"ai_devs_3_tasks/services"
	"context"
	"log"
	"os"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func SaveToDatabase(data string) error {
	ctx := context.Background()
	// save data to database
	graphDriver, err := services.NewDriver(services.Neo4jConfig{DbUri: os.Getenv("NEO4J_URL"), AuthUser: os.Getenv("NEO4J_USER"), AuthPass: os.Getenv("NEO4J_PASS"), Realm: ""})
	if err != nil {
		return err
	}
	defer graphDriver.Driver.Close(ctx)

	return nil
}

func QueryDatabase(query string) (string, error) {
	ctx := context.Background()
	// save data to database
	log.Printf("Query: %s", query)
	graphDriver, err := services.NewDriver(services.Neo4jConfig{DbUri: os.Getenv("NEO4J_URL"), AuthUser: os.Getenv("NEO4J_USER"), AuthPass: os.Getenv("NEO4J_PASS"), Realm: ""})
	if err != nil {
		return "", err
	}
	defer graphDriver.Driver.Close(ctx)

	session := graphDriver.Driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return "error", err
	}
	if strings.Contains(query, "MERGE") {
		return "OK", nil
	}
	var response []string
	for result.Next(ctx) {
		record := result.Record()
		for _, key := range record.Keys {
			value, _ := record.Get(key)
			response = append(response, value.(string))
		}
	}

	if len(response) == 0 {
		return "No records found", nil
	}
	log.Printf("Response: %v", response)
	return response[0], nil
}
