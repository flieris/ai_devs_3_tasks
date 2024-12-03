package main

import (
	"ai_devs_3_tasks/helpers"
	"encoding/json"
	"errors"
	"log"
	"os"
)

type DbQuery struct {
	Task   string `json:"task"`
	ApiKey string `json:"apikey"`
	Query  string `json:"query"`
}

type DbResponse struct {
	Reply []interface{} `json:"reply"`
	Error string        `json:"error"`
}

type QueryRequest struct {
	ApiKey string `json:"apikey"`
	Query  string `json:"query,omitempty"`
	UserId string `json:"userID,omitempty"`
}

type QueryResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type GpsResponse struct {
	Code    int64       `json:"code"`
	Message Coordinates `json:"message"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func SendDbQuery(query string) (string, error) {
	queryToSend := DbQuery{
		Task:   "database",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Query:  query,
	}
	var responseObj DbResponse
	err := helpers.SendRequest(os.Getenv("S03E03_URL"), queryToSend, &responseObj)
	if err != nil {
		return "", err
	}

	responseJson, err := json.Marshal(responseObj)
	if err != nil {
		return "", err
	}

	return string(responseJson), nil
}

func SendAPIQuery(api string, query string) (string, error) {
	var url string
	var queryToSend QueryRequest
	switch api {
	case "PLACES":
		url = os.Getenv("PLACES_URL")
		queryToSend = QueryRequest{
			ApiKey: os.Getenv("CENTRAL_API_KEY"),
			Query:  query,
		}
	case "PEOPLE":
		url = os.Getenv("PEOPLE_URL")
		queryToSend = QueryRequest{
			ApiKey: os.Getenv("CENTRAL_API_KEY"),
			Query:  query,
		}
	case "GPS":
		url = os.Getenv("GPS_URL")
		queryToSend = QueryRequest{
			ApiKey: os.Getenv("CENTRAL_API_KEY"),
			UserId: query,
		}
		log.Printf("Query: %v", queryToSend)
		var responseObj GpsResponse
		err := helpers.SendRequest(url, queryToSend, &responseObj)
		if err != nil {
			return "", err
		}
		responseJson, err := json.Marshal(responseObj)
		if err != nil {
			return "", err
		}

		return string(responseJson), nil
	default:
		return "", errors.New("Unknown API")
	}
	var responseObj QueryResponse
	err := helpers.SendRequest(url, queryToSend, &responseObj)
	if err != nil {
		return "", err
	}
	responseJson, err := json.Marshal(responseObj)
	if err != nil {
		return "", err
	}

	return string(responseJson), nil
}
