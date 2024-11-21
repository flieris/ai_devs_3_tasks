package services

import (
	"github.com/qdrant/go-client/qdrant"
)

type QdrantService struct {
	Config         qdrant.Config
	CollectionName string
}

func (client QdrantService) CreateCollection(ctx context)
