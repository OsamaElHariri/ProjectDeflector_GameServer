package repositories

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepositoryFactory struct {
	client *mongo.Client
}

func getMongoRepositoryFactory() MongoRepositoryFactory {
	url := os.Getenv("DB_URL")
	client, err := mongo.NewClient(options.Client().ApplyURI(url))

	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return MongoRepositoryFactory{
		client: client,
	}
}

func (factory MongoRepositoryFactory) GetRepository() (Repository, func(), error) {
	client := factory.client
	ctx, cancelContext := context.WithTimeout(context.TODO(), 5*time.Second)

	cleanUpFunc := func() {
		cancelContext()
	}

	repo := MongoRepository{
		client: client,
		ctx:    ctx,
	}

	return repo, cleanUpFunc, nil
}
