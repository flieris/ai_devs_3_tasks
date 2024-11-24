package services

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jDriver struct {
	Config Neo4jConfig
	Driver neo4j.DriverWithContext
}

type Neo4jConfig struct {
	DbUri    string
	AuthUser string
	AuthPass string
	Realm    string
}

func NewDriver(config Neo4jConfig) (*Neo4jDriver, error) {
	driver, err := neo4j.NewDriverWithContext(config.DbUri, neo4j.BasicAuth(config.AuthUser, config.AuthPass, config.Realm))
	if err != nil {
		return nil, err
	}
	neoDriver := &Neo4jDriver{
		Config: config,
		Driver: driver,
	}
	return neoDriver, nil
}

type UserNodeItem struct {
	UserId      int64
	UserName    string
	Connections []any
	Embedding   []any
}

func (n *Neo4jDriver) CreateVectorIndex(ctx context.Context, indexName string, label string, propertyName string, dimension int64, similarityFunction string) error {
	query := fmt.Sprintf("CREATE VECTOR INDEX %s IF NOT EXISTS\n"+
		"FOR (n:%s) on (n.%s)\n"+
		"OPTIONS {\n"+
		"    indexConfig: {\n"+
		"  `vector.dimensions`: $dimension,\n"+
		"  `vector.similarity_function`: $similarityFunction\n"+
		"    }\n"+
		"}", indexName, label, propertyName)
	_, err := neo4j.ExecuteQuery(ctx, n.Driver, query,
		map[string]any{
			"dimension":          dimension,
			"similarityFunction": similarityFunction,
		}, neo4j.EagerResultTransformer)
	if err != nil {
		return err
	}
	return nil
}

func (n *Neo4jDriver) InsertUserItem(ctx context.Context, userId int64, userName string, userConns []int64, embedding []float32) (*UserNodeItem, error) {
	result, err := neo4j.ExecuteQuery(ctx, n.Driver,
		"MERGE (n:User { id: $id, name: $name, connections: $connections, embedding: $embedding }) RETURN n",
		map[string]any{
			"id":          userId,
			"name":        userName,
			"connections": userConns,
			"embedding":   embedding,
		}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}
	userNode, _, err := neo4j.GetRecordValue[neo4j.Node](result.Records[0], "n")
	if err != nil {
		return nil, err
	}
	id, err := neo4j.GetProperty[int64](userNode, "id")
	if err != nil {
		return nil, err
	}
	name, err := neo4j.GetProperty[string](userNode, "name")
	if err != nil {
		return nil, err
	}

	connections, err := neo4j.GetProperty[[]any](userNode, "connections")
	if err != nil {
		return nil, err
	}
	embed, err := neo4j.GetProperty[[]any](userNode, "embedding")
	if err != nil {
		return nil, err
	}
	return &UserNodeItem{UserId: id, UserName: name, Connections: connections, Embedding: embed}, nil
}

func (n *Neo4jDriver) ConnectNodes(ctx context.Context, srcNodeId int64, toNodeId int64, relationshipType string) error {
	query := fmt.Sprintf(`MATCH (a), (b) 
WHERE a.id = $srcNodeId and b.id = $toNodeId 
MERGE (a)-[:%s]->(b)`, relationshipType)
	_, err := neo4j.ExecuteQuery(ctx, n.Driver, query,
		map[string]any{
			"srcNodeId": srcNodeId,
			"toNodeId":  toNodeId,
		}, neo4j.EagerResultTransformer)
	if err != nil {
		return err
	}
	return nil
}

func (n *Neo4jDriver) RunQuery(ctx context.Context, params map[string]any, query string) (*neo4j.EagerResult, error) {
	result, err := neo4j.ExecuteQuery(ctx, n.Driver, query, params, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *UserNodeItem) String() string {
	return fmt.Sprintf("User (id: %d, name: %q, connections: %v)", i.UserId, i.UserName, i.Connections)
}
