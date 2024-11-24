package main

import (
	"ai_devs_3_tasks/helpers"
	"ai_devs_3_tasks/llmservices"
	"ai_devs_3_tasks/services"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/sashabaranov/go-openai"
)

type DbRequest struct {
	Task   string `json:"task"`
	ApiKey string `json:"apikey"`
	Query  string `json:"query"`
}

type DbResponse struct {
	Reply []map[string]string `json:"reply"`
	Error string              `json:"error"`
}

type DbUser struct {
	UserId      int64
	UserName    string
	Connections []int64
	Embedding   []float32
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func getEmbeddings(content string) ([]float32, error) {
	openAI := llmservices.NewOpenAiServcie()
	ctx := context.Background()

	req, err := openAI.CreateEmbedding(ctx, openai.EmbeddingRequest{
		Model: openai.SmallEmbedding3,
		Input: content,
	})
	if err != nil {
		return nil, err
	}
	return req.Data[0].Embedding, nil
}

func sendQuery(query string) (*DbResponse, error) {
	queryToSend := DbRequest{
		Task:   "database",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Query:  query,
	}
	var responseObj DbResponse
	err := helpers.SendRequest(os.Getenv("S03E03_URL"), queryToSend, &responseObj)
	if err != nil {
		return nil, err
	}
	return &responseObj, nil
}

func main() {
	ctx := context.Background()
	graphDriver, err := services.NewDriver(services.Neo4jConfig{DbUri: os.Getenv("NEO4J_URL"), AuthUser: os.Getenv("NEO4J_USER"), AuthPass: os.Getenv("NEO4J_PASS"), Realm: ""})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer graphDriver.Driver.Close(ctx)

	err = graphDriver.CreateVectorIndex(ctx, "user_index", "User", "embedding", 1536, "cosine")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	test, err := sendQuery("select username, id from users;")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	var users []*DbUser
	for _, userMap := range test.Reply {
		id, err := strconv.Atoi(userMap["id"])
		if err != nil {
			log.Fatalf("Error converting value to int: %v", err)
		}
		embedding, err := getEmbeddings(userMap["username"])
		if err != nil {
			log.Fatalf("Error getting embedding: %v", err)
		}
		user := DbUser{
			UserId:    int64(id),
			UserName:  userMap["username"],
			Embedding: embedding,
		}
		users = append(users, &user)
	}

	for _, user := range users {
		sql := fmt.Sprintf(`
select user1_id as connection_id from connections where user2_id = %v
      UNION
select user2_id as connection_id from connections where user1_id = %v;
      `, user.UserId, user.UserId)
		connDbResponse, err := sendQuery(sql)
		if err != nil {
			log.Printf("Error gettin connections for userId %d : %v", user.UserId, err)
		}
		for _, conId := range connDbResponse.Reply {
			intConId, err := strconv.Atoi(conId["connection_id"])
			if err != nil {
				log.Printf("Error converting value %v to int %v", conId["connection_id"], err)
			}
			user.Connections = append(user.Connections, int64(intConId))
		}
	}
	for _, user := range users {
		_, err := graphDriver.InsertUserItem(ctx, (*user).UserId, (*user).UserName, (*user).Connections, (*user).Embedding)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	for _, user := range users {
		for _, connId := range (*user).Connections {
			err := graphDriver.ConnectNodes(ctx, (*user).UserId, connId, "CONNECTED_TO")
			if err != nil {
				log.Printf("Error: %v", err)
			}
		}
	}

	query := "MATCH p = shortestPath((person1:User)-[*]-(person2:User)) WHERE person1.name = 'Rafał' AND person2.name = 'Barbara' RETURN p"

	result, err := graphDriver.RunQuery(ctx, nil, query)
	if err != nil {
		log.Fatalf("Error executing query: %v", err)
	}

	var pathToUser []string
	for _, record := range result.Records {
		path, ok := record.Get("p")
		if !ok {
			continue
		}
		if pathObj, isPath := path.(neo4j.Path); isPath {
			for _, node := range pathObj.Nodes {
				if name, ok := node.Props["name"].(string); ok {
					pathToUser = append(pathToUser, name)
				}
			}
		} else {
			log.Println("Returned object is not a path")
		}
	}
	log.Println(strings.Join(pathToUser, ", "))
	message := helpers.JsonAnswer{
		Task:   "connections",
		ApiKey: os.Getenv("CENTRAL_API_KEY"),
		Answer: strings.Join(pathToUser, ", "),
	}
	respCentral, err := helpers.SendAnswer(os.Getenv("REPORT_URL"), message)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from central: %s", respCentral.Message)
}
